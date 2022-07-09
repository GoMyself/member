package controller

import (
	"github.com/valyala/fasthttp"
	"member/contrib/helper"
	"member/model"
)

type TutorialController struct{}

// 新手教程已读
func (that *TutorialController) Read(ctx *fasthttp.RequestCtx) {

	mb, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	err = model.TutorialRead(mb.UID)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *TutorialController) State(ctx *fasthttp.RequestCtx) {

	mb, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	data, err := model.TutorialState(mb.UID)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}
