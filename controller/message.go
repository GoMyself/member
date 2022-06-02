package controller

import (
	"github.com/valyala/fasthttp"
	"member/contrib/helper"
	"member/contrib/validator"
	"member/model"
	"strings"
)

type MessageController struct{}

// 内信列表
func (that *MessageController) List(ctx *fasthttp.RequestCtx) {

	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")
	ty := ctx.QueryArgs().GetUintOrZero("ty") //1 站内消息 2 活动消息

	if ty > 0 {
		tys := map[int]bool{
			1: true,
			2: true,
		}
		if _, ok := tys[ty]; !ok {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}
	}

	mb, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	if page == 0 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	s, err := model.MessageList(ty, page, pageSize, mb.Username)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.PrintJson(ctx, true, s)
}

// 站内信已读
func (that *MessageController) Num(ctx *fasthttp.RequestCtx) {

	mb, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	num, err := model.MessageNum(mb.Username)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, num)
}

// 站内信已读
func (that *MessageController) Read(ctx *fasthttp.RequestCtx) {

	id := string(ctx.QueryArgs().Peek("id"))
	mb, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	err = model.MessageRead(id, mb.Username)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 站内信删除
func (that *MessageController) Delete(ctx *fasthttp.RequestCtx) {

	flag := ctx.QueryArgs().GetUintOrZero("flag") // 1 精确删除 2 删除所有已读
	ids := string(ctx.QueryArgs().Peek("ids"))

	flags := map[int]bool{
		1: true,
		2: true,
	}
	if _, ok := flags[flag]; !ok {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	var s []interface{}
	if flag == 1 {
		if ids == "" {
			helper.Print(ctx, false, helper.IDErr)
			return
		}

		for _, v := range strings.Split(ids, ",") {
			if !validator.CtypeDigit(v) {
				helper.Print(ctx, false, helper.IDErr)
				return
			}

			s = append(s, v)
		}
	}

	mb, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	err = model.MessageDelete(s, mb.Username, flag)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
