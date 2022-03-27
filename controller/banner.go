package controller

import (
	"github.com/valyala/fasthttp"
	"member2/contrib/helper"
	"member2/model"
	"strconv"
)

type BannerController struct{}

func (that *BannerController) Images(ctx *fasthttp.RequestCtx) {

	flags := string(ctx.QueryArgs().Peek("flags"))
	d := string(ctx.Request.Header.Peek("d"))
	i, err := strconv.Atoi(d)
	if err != nil {
		helper.Print(ctx, false, helper.DeviceTypeErr)
		return
	}

	if _, ok := model.Devices[i]; !ok {
		helper.Print(ctx, false, helper.DeviceTypeErr)
		return
	}

	f, err := strconv.Atoi(flags)
	if err != nil || (f < 1 || f > 5) {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	data, err := model.BannerImages(flags, d)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.PrintJson(ctx, true, data)
}
