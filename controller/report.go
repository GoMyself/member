package controller

import (
	"github.com/valyala/fasthttp"
	"member/contrib/helper"
	"member/model"
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

func (that *ReportController) SubReport(ctx *fasthttp.RequestCtx) {

	ty := string(ctx.QueryArgs().Peek("ty"))
	flag := string(ctx.QueryArgs().Peek("flag"))
	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")
	if flag == "1" || flag == "2" || flag == "3" {
		data, err := model.SubAgencyReport(ty, flag, page, pageSize, ctx)
		if err != nil {
			helper.Print(ctx, false, err.Error())
			return
		}
		helper.Print(ctx, true, data)

	} else {
		data, err := model.SubAgencyList(page, pageSize, ctx)
		if err != nil {
			helper.Print(ctx, false, err.Error())
			return
		}
		helper.Print(ctx, true, data)

	}

}

func (that *ReportController) SubGameRecord(ctx *fasthttp.RequestCtx) {

	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")
	flag := ctx.QueryArgs().GetUintOrZero("flag") //0全部1待开奖2未中奖3已中奖
	gameType := ctx.QueryArgs().GetUintOrZero("game_type")
	platformID := ctx.QueryArgs().GetUintOrZero("platform_id")
	dateFlag := ctx.QueryArgs().GetUintOrZero("date_flag")    //1今天2昨天3七天
	playerName := string(ctx.QueryArgs().Peek("player_name")) //下级会员名

	user, err := model.MemberCache(ctx, "")
	if err != nil {
		//fmt.Println("Game MemberInfo err = ", err.Error())
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	data, err := model.SubGameRecord(user.UID, playerName, gameType, dateFlag, flag, platformID, uint(pageSize), uint(page))
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *ReportController) SubTradeRecord(ctx *fasthttp.RequestCtx) {

	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")
	flag := ctx.QueryArgs().GetUintOrZero("flag")             //1充值2提现
	dateFlag := ctx.QueryArgs().GetUintOrZero("date_flag")    //1今天2昨天3七天
	playerName := string(ctx.QueryArgs().Peek("player_name")) //下级会员名

	user, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	data, err := model.SubTradeRecord(user.UID, playerName, dateFlag, flag, uint(pageSize), uint(page))
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}

func (that *ReportController) List(ctx *fasthttp.RequestCtx) {

	ty := string(ctx.QueryArgs().Peek("ty")) //1今天2昨天3本月4上月
	playerName := string(ctx.QueryArgs().Peek("player_name"))
	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")
	isOnline := ctx.QueryArgs().GetUintOrZero("is_online")

	data, err := model.AgencyReportList(ty, ctx, playerName, page, pageSize, isOnline)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, data)
}
