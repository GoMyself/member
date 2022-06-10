package controller

import (
	"github.com/valyala/fasthttp"
	"member/contrib/helper"
	"member/model"
	"strings"
)

type ShortURLController struct{}

func (that *ShortURLController) Gen(ctx *fasthttp.RequestCtx) {

	id := string(ctx.PostArgs().Peek("id"))
	uri := strings.TrimSpace(string(ctx.PostArgs().Peek("url")))
	if !helper.CtypeDigit(id) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	resc, err := model.ShortURLGen(id, uri)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.PrintJson(ctx, false, resc)
}
