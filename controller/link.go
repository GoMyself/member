package controller

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"member/contrib/helper"
	"member/contrib/validator"
	"member/model"
	"strconv"
)

type LinkController struct{}

func (that *LinkController) Insert(ctx *fasthttp.RequestCtx) {

	params := model.Link_t{}
	err := validator.Bind(ctx, &params)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	uri := string(ctx.PostArgs().Peek("url"))

	device := string(ctx.Request.Header.Peek("d"))
	i, err := strconv.Atoi(device)
	if err != nil {
		helper.Print(ctx, false, helper.DeviceTypeErr)
		return
	}

	if _, ok := model.Devices[i]; !ok {
		helper.Print(ctx, false, helper.DeviceTypeErr)
		return
	}

	params.ID = helper.GenId()
	params.CreatedAt = fmt.Sprintf("%d", ctx.Time().Unix())
	err = model.LinkInsert(ctx, uri, i, params)
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
