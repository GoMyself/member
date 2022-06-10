package controller

import (
	"github.com/valyala/fasthttp"
)

type ShortURLController struct{}

func (that *ShortURLController) Domain(ctx *fasthttp.RequestCtx) {

	//resc, err := model.ShortURLGen(uri)
	//if err != nil {
	//	helper.Print(ctx, false, err.Error())
	//	return
	//}

	//helper.PrintJson(ctx, false, resc)
}
