package model

import (
	"database/sql"
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"
	"member2/contrib/helper"
)

type ReportAgency struct {
	BetAmount         float64 `json:"bet_amount,omitempty" db:"bet_amount"`
	Deposit           float64 `json:"deposit,omitempty" db:"deposit"`
	Withdraw          float64 `json:"withdraw,omitempty" db:"withdraw"`
	BetMemCount       int64   `json:"bet_mem_count,omitempty" db:"bet_mem_count"`
	FirstDepositCount int64   `json:"first_deposit_count,omitempty" db:"first_deposit_count"`
	RegistCount       int64   `json:"regist_count,omitempty" db:"regist_count"`
	MemCount          int64   `json:"mem_count,omitempty" db:"mem_count"`
	Rebate            float64 `json:"rebate,omitempty" db:"rebate"`
	NetAmount         float64 `json:"net_amount,omitempty" db:"net_amount"`
	DividendAmount    float64 `json:"dividend_amount,omitempty" db:"dividend_amount"`
	BalanceTotal      float64 `json:"balance_total,omitempty" db:"balance_total"`
}

func AgencyReport(ty string, fCtx *fasthttp.RequestCtx) (ReportAgency, error) {

	data := ReportAgency{}
	mb, err := MemberCache(fCtx, "")
	if err != nil {
		return data, errors.New(helper.AccessTokenExpires)
	}

	var startAt int64
	var reportType int
	switch ty {
	case "1": //今天
		startAt = helper.DayTST(0, loc).Unix()
		reportType = 2
	case "2": //昨天
		startAt = helper.DayTST(0, loc).Unix() - 24*60*60
		reportType = 2
	case "3": //本月
		startAt = helper.MonthTST(0, loc).Unix()
		reportType = 4
	case "4": //上月
		startAt = helper.MonthTST(helper.MonthTST(0, loc).Unix()-1, loc).Unix()
		reportType = 4
	default:

	}
	// 获取统计数据
	and := g.And(
		g.C("report_type").Eq(reportType),
		g.C("prefix").Eq(meta.Prefix),
		g.C("report_time").Eq(startAt),
		g.C("uid").Eq(mb.UID),
		g.C("data_type").Eq(1),
	)

	query, _, _ := dialect.From("tbl_report_agency").Where(and).
		Select(
			g.C("bet_amount").As("bet_amount"),                   //投注金额
			g.C("deposit_amount").As("deposit"),                  //充值金额
			g.C("withdrawal_amount").As("withdraw"),              //提现金额
			g.C("bet_mem_count").As("bet_mem_count"),             //投注人数
			g.C("first_deposit_count").As("first_deposit_count"), //首存人数
			g.C("regist_count").As("regist_count"),               //注册人数
			g.C("mem_count").As("mem_count"),                     //下级人数
			g.C("rebate_amount").As("rebate"),                    //返水
			g.C("company_net_amount").As("net_amount"),           //输赢
			g.C("dividend_amount").As("dividend_amount"),         //活动礼金
			g.C("balance_total").As("balance_total"),             //团队余额
		).
		ToSQL()
	err = meta.ReportDB.Get(&data, query)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err.Error())
		return data, pushLog(err, helper.DBErr)
	}
	return data, nil
}
