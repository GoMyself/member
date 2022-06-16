package controller

import (
	"database/sql"
	"fmt"
	"github.com/valyala/fasthttp"
	"member/contrib/helper"
	"member/model"
	"strings"
)

type MessageController struct{}

// 内信列表
func (that *MessageController) List(ctx *fasthttp.RequestCtx) {

	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")
	ty := ctx.QueryArgs().GetUintOrZero("ty")         //1 站内消息 2 活动消息
	isRead := string(ctx.QueryArgs().Peek("is_read")) //0 未读 1已读

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

	if isRead != "0" && isRead != "1" {
		helper.Print(ctx, false, helper.ParamErr)
		return
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
	data, err := model.MessageList(ty, page, pageSize, isRead, mb.Username)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 站内信已读
func (that *MessageController) Emergency(ctx *fasthttp.RequestCtx) {

	mb, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	data, err := model.MessageEmergency(mb.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			helper.Print(ctx, true, nil)
			return
		}

		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
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

	ts := string(ctx.PostArgs().Peek("ts"))
	_, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	fmt.Println(ts)
	err = model.MessageRead(ts)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 站内信删除
func (that *MessageController) Delete(ctx *fasthttp.RequestCtx) {

	flag := ctx.PostArgs().GetUintOrZero("flag") // 1 精确删除 2 删除所有已读
	tss := string(ctx.PostArgs().Peek("tss"))

	fmt.Println(flag, tss)
	flags := map[int]bool{
		1: true,
		2: true,
	}
	if _, ok := flags[flag]; !ok {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	mb, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	err = model.MessageDelete(mb.Username, strings.Split(tss, ","), flag)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
