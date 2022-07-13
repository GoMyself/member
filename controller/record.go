package controller

import (
	"member/contrib/helper"
	"member/model"
	"strconv"

	"github.com/valyala/fasthttp"
)

// 投注记录
type RecordController struct{}

// 投注列表
func (that *RecordController) Game(ctx *fasthttp.RequestCtx) {

	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")
	flag := string(ctx.QueryArgs().Peek("flag"))
	platformID := ctx.QueryArgs().GetUintOrZero("platform_id")
	startTime := string(ctx.QueryArgs().Peek("start_time"))
	endTime := string(ctx.QueryArgs().Peek("end_time"))
	ty := ctx.QueryArgs().GetUintOrZero("ty")                 // 1直属下级
	playerName := string(ctx.QueryArgs().Peek("player_name")) //下级会员名

	user, err := model.MemberCache(ctx, "")
	if err != nil {
		//fmt.Println("Game MemberInfo err = ", err.Error())
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	flags := -1
	if len(flag) > 0 {
		flags, _ = strconv.Atoi(flag)
	}

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	data, err := model.RecordGame(ty, user.UID, playerName, startTime, endTime, flags, platformID, uint(pageSize), uint(page))
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// Transfer 场馆转账列表
func (that *RecordController) Transfer(ctx *fasthttp.RequestCtx) {

	billNo := string(ctx.QueryArgs().Peek("bill_no"))
	pidIn := string(ctx.QueryArgs().Peek("pid_in"))
	pidOut := string(ctx.QueryArgs().Peek("pid_out"))
	transferType := string(ctx.QueryArgs().Peek("transfer_type"))
	startTime := string(ctx.QueryArgs().Peek("start_time"))
	endTime := string(ctx.QueryArgs().Peek("end_time"))
	state := string(ctx.QueryArgs().Peek("state"))
	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	user, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	data, err := model.RecordTransfer(user.Username, billNo, state, transferType, pidIn, pidOut, startTime, endTime, uint(page), uint(pageSize))
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 账变记录列表
func (that *RecordController) Transaction(ctx *fasthttp.RequestCtx) {

	startTime := string(ctx.QueryArgs().Peek("start_time"))
	endTime := string(ctx.QueryArgs().Peek("end_time"))
	cashTypes := string(ctx.QueryArgs().Peek("types"))
	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")

	user, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	data, err := model.RecordTransaction(user.UID, cashTypes, startTime, endTime, uint(page), uint(pageSize))
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 交易记录
func (that *RecordController) Trade(ctx *fasthttp.RequestCtx) {

	startTime := string(ctx.QueryArgs().Peek("start_time"))
	endTime := string(ctx.QueryArgs().Peek("end_time"))
	flag := ctx.QueryArgs().GetUintOrZero("flag") //账变类型 271 存款 272 取款 273 转账 274 红利 275 佣金返水 278 调整
	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")

	user, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	data, err := model.RecordTrade(user.UID, startTime, endTime, flag, uint(page), uint(pageSize))
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

// 交易记录 详情
func (that *RecordController) TradeDetail(ctx *fasthttp.RequestCtx) {

	id := ctx.QueryArgs().Peek("id")
	if string(id) == "" {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	flagByte := ctx.QueryArgs().Peek("flag")
	flag, err := strconv.Atoi(string(flagByte))
	if err != nil || flag < model.RecordTradeDeposit || flag > model.RecordTradeTransfer {
		helper.Print(ctx, false, helper.TradeTypeErr)
		return
	}

	user, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	data, err := model.RecordTradeDetail(flag, user.UID, string(id))
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}
