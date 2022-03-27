package controller

import (
	"github.com/valyala/fasthttp"
	"member2/contrib/helper"
	"member2/model"
)

type RebateController struct{}

func (that *RebateController) Scale(ctx *fasthttp.RequestCtx) {

	username := string(ctx.UserValue("token").([]byte))
	if username == "" {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	mb, err := model.MemberFindOne(username)
	if username == "" {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	vs, err := model.RebateScale(mb.UID)
	if err != nil {
		helper.Print(ctx, false, err.Error())
	}

	helper.Print(ctx, true, vs)
}
