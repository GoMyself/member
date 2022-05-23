package controller

import (
	"github.com/valyala/fasthttp"
	"member2/contrib/helper"
	"member2/model"
)

type VipController struct{}

func (that *VipController) Config(ctx *fasthttp.RequestCtx) {
	helper.PrintJson(ctx, true, model.VipConfig())
}

func (that *VipController) Info(ctx *fasthttp.RequestCtx) {

	mb, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	data, err := model.VipInfo(mb)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}
