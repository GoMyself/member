package controller

import (
	"github.com/valyala/fasthttp"
	"member2/contrib/helper"
	"member2/contrib/validator"
	"member2/model"
)

type CommissionController struct{}

// 佣金提取
func (that *CommissionController) Draw(ctx *fasthttp.RequestCtx) {

	withdrawPwd := string(ctx.PostArgs().Peek("withdraw_pwd"))
	amount := string(ctx.PostArgs().Peek("amount"))
	err := model.CommissionDraw(withdrawPwd, amount, ctx)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *CommissionController) Ration(ctx *fasthttp.RequestCtx) {

	withdrawPwd := string(ctx.PostArgs().Peek("withdraw_pwd"))
	username := string(ctx.PostArgs().Peek("username"))
	amount := string(ctx.PostArgs().Peek("amount"))

	if !validator.CheckUName(username, 4, 9) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	err := model.CommissionRation(withdrawPwd, username, amount, ctx)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
