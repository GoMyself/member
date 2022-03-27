package controller

import (
	"github.com/valyala/fasthttp"
	"member2/contrib/helper"
	"member2/model"
)

type UpgradeController struct{}

var (
	apps = map[string]string{
		"26": "ios",
		"27": "android",
		"35": "android",
		"36": "ios",
	}
)

func (that *UpgradeController) Info(ctx *fasthttp.RequestCtx) {

	d := string(ctx.Request.Header.Peek("d"))
	device, ok := apps[d]
	if !ok {
		helper.PrintJson(ctx, false, "{}")
		return
	}

	s, err := model.UpgradeInfo(device)
	if err != nil {
		helper.PrintJson(ctx, false, "{}")
		return
	}

	helper.PrintJson(ctx, true, s)
}
