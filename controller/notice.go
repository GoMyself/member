package controller

import (
	"github.com/valyala/fasthttp"
	"member2/contrib/helper"
	"member2/model"
)

type NoticeController struct{}

func (that *NoticeController) List(ctx *fasthttp.RequestCtx) {

	data, err := model.Notices()
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.PrintJson(ctx, true, data)
}
