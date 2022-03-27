package controller

import (
	"github.com/valyala/fasthttp"
	"member2/contrib/helper"
	"member2/model"
)

type TreeController struct{}

func (that *TreeController) List(ctx *fasthttp.RequestCtx) {

	id := string(ctx.QueryArgs().Peek("id"))

	if !helper.CtypeDigit(id) {
		helper.Print(ctx, false, "id")
		return
	}

	data, err := model.TreeList(id)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}
	helper.PrintJson(ctx, true, data)
}
