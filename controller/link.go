package controller

import (
	"member2/contrib/helper"
	"member2/contrib/validator"
	"member2/model"

	"github.com/valyala/fasthttp"
)

type LinkController struct{}

func (that *LinkController) Insert(ctx *fasthttp.RequestCtx) {

	params := model.Link_t{}
	err := validator.Bind(ctx, &params)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	params.ID = helper.GenId()
	params.CreatedAt = uint32(ctx.Time().Unix())
	err = model.LinkInsert(ctx, params)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}
	helper.Print(ctx, true, helper.Success)
}

func (that *LinkController) List(ctx *fasthttp.RequestCtx) {

	data, err := model.LinkList(ctx)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}
	helper.Print(ctx, true, data)
}

func (that *LinkController) Delete(ctx *fasthttp.RequestCtx) {

	id := string(ctx.QueryArgs().Peek("id"))

	if !helper.CtypeDigit(id) {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	err := model.LinkDelete(ctx, id)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}
	helper.Print(ctx, true, helper.Success)
}
