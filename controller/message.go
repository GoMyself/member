package controller

import (
	"github.com/valyala/fasthttp"
	"member2/contrib/helper"
	"member2/contrib/validator"
	"member2/model"
	"strings"
)

type MessageController struct{}

// 内信列表
func (that *MessageController) List(ctx *fasthttp.RequestCtx) {

	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")
	ty := ctx.QueryArgs().GetUintOrZero("ty")
	username := string(ctx.UserValue("token").([]byte))
	if username == "" {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	tys := map[int]bool{
		1: true,
		2: true,
	}
	if _, ok := tys[ty]; !ok {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	mb, err := model.MemberFindOne(username)
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	if page == 0 {
		page = 1
	}
	if pageSize < 10 {
		pageSize = 10
	}
	s, err := model.MessageList(ty, page, pageSize, mb.Username)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.PrintJson(ctx, true, s)
}

// 站内信已读
func (that *MessageController) Read(ctx *fasthttp.RequestCtx) {

	id := string(ctx.QueryArgs().Peek("id"))
	username := string(ctx.UserValue("token").([]byte))
	if username == "" {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	mb, err := model.MemberFindOne(username)
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
	username := string(ctx.UserValue("token").([]byte))
	if username == "" {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	flags := map[int]bool{
		1: true,
		2: true,
	}
	if _, ok := flags[flag]; !ok {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	s := strings.Split(ids, ",")
	if flag == 1 {
		if len(s) == 0 {
			helper.Print(ctx, false, helper.IDErr)
			return
		}

		for _, v := range s {
			if !validator.CtypeDigit(v) {
				helper.Print(ctx, false, helper.IDErr)
				return
			}
		}
	}

	mb, err := model.MemberFindOne(username)
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
