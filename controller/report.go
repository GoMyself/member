package controller

import (
	"github.com/valyala/fasthttp"
	"member2/contrib/helper"
	"member2/model"
)

type ReportController struct{}

//Report 代理报表
func (that *ReportController) Report(ctx *fasthttp.RequestCtx) {

	ty := string(ctx.PostArgs().Peek("ty"))
	username := string(ctx.PostArgs().Peek("username"))
	data, err := model.AgencyReport(ty, ctx, username)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}
