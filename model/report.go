package model

import (
	"database/sql"
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/shopspring/decimal"
	"github.com/valyala/fasthttp"
	"member/contrib/helper"
	"member/contrib/validator"
	"strconv"
	"strings"
)

type ReportAgencyData struct {
	T int64             `json:"t"`
	D []ReportSubAgency `json:"d"`
}

type ReportSubAgency struct {
	Uid            string  `json:"uid" db:"uid"`
	Username       string  `json:"username" db:"username"`
	Lvl            int     `json:"lvl" db:"lvl"`
	BetAmount      float64 `json:"bet_amount" db:"bet_amount"`
	BetMemCount    int64   `json:"bet_mem_count" db:"bet_mem_count"`
	Rebate         float64 `json:"rebate" db:"rebate"`
	DividendAmount float64 `json:"dividend_amount" db:"dividend_amount"`
	WinAmount      float64 `json:"win_amount" db:"win_amount"`
	Profit         float64 `json:"profit" db:"profit"`
	CgRebate       float64 `json:"cg_rebate" db:"cg_rebate"`
	IsOnline       int     `json:"is_online" db:"-"`
}

type ReportAgency struct {
	BetAmount         float64 `json:"bet_amount" db:"bet_amount"`
	Deposit           float64 `json:"deposit" db:"deposit"`
	Withdraw          float64 `json:"withdraw" db:"withdraw"`
	BetMemCount       int64   `json:"bet_mem_count" db:"bet_mem_count"`
	FirstDepositCount int64   `json:"first_deposit_count" db:"first_deposit_count"`
	RegistCount       int64   `json:"regist_count" db:"regist_count"`
	MemCount          int64   `json:"mem_count" db:"mem_count"`
	Rebate            float64 `json:"rebate" db:"rebate"`
	TeamRebate        float64 `json:"team_rebate" db:"team_rebate"`
	NetAmount         float64 `json:"net_amount" db:"net_amount"`
	Presettle         float64 `json:"-" db:"presettle"`
	DividendAmount    float64 `json:"dividend_amount" db:"dividend_amount"`
	BalanceTotal      float64 `json:"balance_total" db:"balance_total"`
	WinAmount         float64 `json:"win_amount" db:"win_amount"`
	CgRebate          float64 `json:"cg_rebate" db:"cg_rebate"`
	Profit            float64 `json:"profit" db:"profit"`
}

func AgencyReport(ty string, fCtx *fasthttp.RequestCtx, username string) (ReportAgency, error) {

	data := ReportAgency{}
	mb, err := MemberCache(fCtx, "")
	if err != nil {
		return data, errors.New(helper.AccessTokenExpires)
	}
	userId := mb.UID
	if len(username) > 0 && username != mb.Username {

		var count int64
		mb, err = MemberCache(nil, username)
		if err != nil {
			return data, errors.New(helper.UsernameExist)
		}
		ex := g.Ex{
			"ancestor":   userId,
			"descendant": mb.UID,
			"prefix":     meta.Prefix,
		}
		query, _, _ := dialect.From("tbl_members_tree").Select(g.COUNT("*")).Where(ex).Limit(1).ToSQL()
		err := meta.MerchantDB.Get(&count, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}
		if count == 0 {
			return data, errors.New(helper.NotDirectSubordinate)
		}
	}

	var startAt int64
	var reportType int
	switch ty {
	case "1": //??????
		startAt = helper.DayTST(0, loc).Unix()
		reportType = 2
	case "2": //??????
		startAt = helper.DayTST(0, loc).Unix() - 24*60*60
		reportType = 2
	case "3": //??????
		startAt = helper.MonthTST(0, loc).Unix()
		reportType = 4
	case "4": //??????
		startAt = helper.MonthTST(helper.MonthTST(0, loc).Unix()-1, loc).Unix()
		reportType = 4
	default:
		startAt = helper.DayTST(0, loc).Unix()
		reportType = 2
	}
	// ??????????????????
	and := g.And(
		g.C("report_type").Eq(reportType),
		g.C("prefix").Eq(meta.Prefix),
		g.C("report_time").Eq(startAt),
		g.C("uid").Eq(mb.UID),
		g.C("data_type").Eq(1),
	)

	query, _, _ := dialect.From("tbl_report_agency").Where(and).
		Select(
			g.C("bet_amount").As("bet_amount"),                   //????????????
			g.C("deposit_amount").As("deposit"),                  //????????????
			g.C("withdrawal_amount").As("withdraw"),              //????????????
			g.C("bet_mem_count").As("bet_mem_count"),             //????????????
			g.C("first_deposit_count").As("first_deposit_count"), //????????????
			g.C("regist_count").As("regist_count"),               //????????????
			g.C("mem_count").As("mem_count"),                     //????????????
			g.C("rebate_amount").As("rebate"),                    //??????
			g.C("company_net_amount").As("net_amount"),           //??????
			g.C("presettle").As("presettle"),                     //????????????
			g.C("dividend_amount").As("dividend_amount"),         //????????????
			g.C("balance_total").As("balance_total"),             //????????????
			g.C("win_amount").As("win_amount"),                   //????????????
			g.C("cg_rebate").As("cg_rebate"),                     //????????????
			g.C("company_revenue").As("profit"),                  //????????????
		).
		ToSQL()
	fmt.Println(query)
	err = meta.ReportDB.Get(&data, query)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err.Error())
		return data, pushLog(err, helper.DBErr)
	}
	data.NetAmount, _ = (decimal.NewFromFloat(data.NetAmount).Add(decimal.NewFromFloat(data.Presettle))).Mul(decimal.NewFromFloat(-1)).Float64()
	data.Profit, _ = decimal.NewFromFloat(data.Profit).Mul(decimal.NewFromFloat(-1)).Float64()

	var myRebate sql.NullFloat64
	// ??????????????????
	and = g.And(
		g.C("report_type").Eq(reportType),
		g.C("prefix").Eq(meta.Prefix),
		g.C("report_time").Eq(startAt),
		g.C("uid").Eq(mb.UID),
		g.C("data_type").Eq(2),
	)
	query, _, _ = dialect.From("tbl_report_agency").Where(and).
		Select(
			g.C("rebate_amount").As("rebate"), //??????
		).
		ToSQL()
	fmt.Println(query)
	err = meta.ReportDB.Get(&myRebate, query)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err.Error())
		return data, pushLog(err, helper.DBErr)
	}
	data.TeamRebate, _ = decimal.NewFromFloat(data.Rebate).Sub(decimal.NewFromFloat(myRebate.Float64)).Float64()

	return data, nil
}

func SubAgencyReport(ty, flag string, page, pageSize int, fCtx *fasthttp.RequestCtx, username string) (ReportSubMemberData, error) {

	data := ReportSubMemberData{}
	loginUser, err := MemberCache(fCtx, "")
	if err != nil {
		return data, errors.New(helper.AccessTokenExpires)
	}
	mb, err := MemberCache(fCtx, username)
	if err != nil {
		return data, errors.New(helper.AccessTokenExpires)
	}
	if len(username) > 0 {
		var count int64
		ex := g.Ex{
			"ancestor":   loginUser.UID,
			"descendant": mb.UID,
			"prefix":     meta.Prefix,
		}
		query, _, _ := dialect.From("tbl_members_tree").Select(g.COUNT("*")).Where(ex).Limit(1).ToSQL()
		err := meta.MerchantDB.Get(&count, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}
		if count == 0 {
			return data, errors.New(helper.NotDirectSubordinate)
		}
	}

	var startAt int64
	endAt := helper.DayTET(0, loc).Unix()
	var reportType int
	switch ty {
	case "1": //??????
		startAt = helper.DayTST(0, loc).Unix()
		reportType = 2
	case "2": //??????
		startAt = helper.DayTST(0, loc).Unix() - 24*60*60
		endAt = helper.DayTST(0, loc).Unix() - 1
		reportType = 2
	case "3": //??????
		startAt = helper.MonthTST(0, loc).Unix()
		reportType = 4
	case "4": //??????
		startAt = helper.MonthTST(helper.MonthTST(0, loc).Unix()-1, loc).Unix()
		reportType = 4
	default:
		startAt = helper.DayTST(0, loc).Unix()
		reportType = 2
	}

	// ??????????????????
	ex := g.Ex{
		"report_type": reportType,
		"prefix":      meta.Prefix,
		"report_time": startAt,
		"top_uid":     mb.UID,
	}
	orderBy := "uid"
	switch flag {
	case "1":
		orderBy = "bet_amount"
		ex["bet_amount"] = g.Op{"gt": 0}
	//????????????
	case "2":
		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
		orderBy = "created_at"
	//????????????
	case "3":
		ex["first_deposit_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
		orderBy = "first_deposit_at"
	default:

	}
	t := dialect.From("tbl_report_sub_member")
	if page == 1 {
		query, _, _ := t.Select(g.COUNT(1)).Where(ex).ToSQL()
		fmt.Println(query)
		err := meta.ReportDB.Get(&data.T, query)
		if err != nil {
			return data, pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
		}
		if data.T == 0 {
			return data, nil
		}
	}
	offset := (page - 1) * pageSize
	query, _, _ := t.Where(ex).Select(
		g.C("bet_count").As("bet_count"),                       //????????????
		g.C("bet_amount").As("bet_amount"),                     //????????????
		g.C("parent_name").As("parent_name"),                   //???????????????
		g.C("created_at").As("created_at"),                     //????????????
		g.C("username").As("username"),                         //????????????
		g.C("first_deposit_at").As("first_deposit_at"),         //????????????
		g.C("first_deposit_amount").As("first_deposit_amount"), //????????????
	).Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C(orderBy).Desc()).ToSQL()
	fmt.Println(query)
	err = meta.ReportDB.Select(&data.D, query)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err.Error())
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func SubAgencyList(page, pageSize int, fCtx *fasthttp.RequestCtx, username string) (ReportSubMemberData, error) {

	offset := (page - 1) * pageSize
	data := ReportSubMemberData{}
	loginUser, err := MemberCache(fCtx, "")
	if err != nil {
		return data, errors.New(helper.AccessTokenExpires)
	}
	mb, err := MemberCache(fCtx, username)
	if err != nil {
		return data, errors.New(helper.AccessTokenExpires)
	}
	if len(username) > 0 {
		var count int64
		ex := g.Ex{
			"ancestor":   loginUser.UID,
			"descendant": mb.UID,
			"prefix":     meta.Prefix,
		}
		query, _, _ := dialect.From("tbl_members_tree").Select(g.COUNT("*")).Where(ex).Limit(1).ToSQL()
		err := meta.MerchantDB.Get(&count, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}
		if count == 0 {
			return data, errors.New(helper.NotDirectSubordinate)
		}
	}

	if page == 1 {
		query := fmt.Sprintf(`select count(distinct descendant) from tbl_members_tree where ancestor = %s and prefix = '%s'`, mb.UID, meta.Prefix)
		fmt.Println(query)
		err = meta.MerchantDB.Get(&data.T, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}
		if data.T == 0 {
			return data, nil
		}
	}

	var uids []string
	query := fmt.Sprintf(`select descendant from tbl_members_tree where ancestor = %s and prefix = '%s' limit %d,%d`, mb.UID, meta.Prefix, offset, pageSize)
	fmt.Println(query)
	err = meta.MerchantDB.Select(&uids, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}
	ex := g.Ex{
		"uid":         uids,
		"report_time": helper.DayTST(0, loc).Unix(),
		"report_type": 2,
		"data_type":   1,
		"prefix":      meta.Prefix,
	}
	query, _, _ = dialect.From("tbl_report_agency").Where(ex).
		Select(
			g.C("mem_count").As("mem_count"), //????????????
			g.C("username").As("username"),
			g.C("parent_name").As("parent_name"),
		).
		ToSQL()
	fmt.Println(query)
	err = meta.ReportDB.Select(&data.D, query)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err.Error())
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func SubGameRecord(uid, playerName string, gameType, dateType int, flag string, gameID int, pageSize, page uint) (GameRecordData, error) {

	data := GameRecordData{}
	//????????????
	var startAt int64
	endAt := helper.DayTET(0, loc).UnixMilli()

	switch dateType {
	case 1: //??????
		startAt = helper.DayTST(0, loc).UnixMilli()
	case 2: //??????
		startAt = helper.DayTST(0, loc).UnixMilli() - 24*60*60*1000
		endAt = helper.DayTET(0, loc).UnixMilli() - 24*60*60*1000

	case 3: //??????
		startAt = helper.DayTST(0, loc).UnixMilli() - 6*24*60*60*1000
	default:
		startAt = helper.DayTST(0, loc).UnixMilli()
	}

	//????????????
	var uids []string
	ex := g.Ex{}
	if playerName != "" && validator.CheckUName(playerName, 5, 14) {
		//????????????????????????????????????????????????????????????
		var count int64
		mb, err := MemberCache(nil, playerName)
		if err != nil {
			return data, errors.New(helper.UsernameExist)
		}
		ex = g.Ex{
			"ancestor":   uid,
			"descendant": mb.UID,
			"prefix":     meta.Prefix,
		}
		query, _, _ := dialect.From("tbl_members_tree").Select(g.COUNT("*")).Where(ex).Limit(1).ToSQL()
		fmt.Println(query)
		err = meta.MerchantDB.Get(&count, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}
		if count == 0 {
			return data, errors.New(helper.NotDirectSubordinate)
		}
		ex = g.Ex{
			"uid": mb.UID,
		}
	} else {
		ex = g.Ex{
			"top_uid":     uid,
			"report_time": g.Op{"between": exp.NewRangeVal(startAt/1000, endAt/1000)},
			"report_type": 2,
			"prefix":      meta.Prefix,
			"bet_amount":  g.Op{"gt": 0},
		}
		query, _, _ := dialect.From("tbl_report_sub_member").Select(g.C("uid")).Where(ex).GroupBy("uid").Order(g.SUM("bet_amount").Desc()).Limit(100).ToSQL()
		fmt.Println(query)
		err := meta.ReportDB.Select(&uids, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}
		if len(uids) > 0 {
			uids = append(uids, uid)
			ex = g.Ex{
				"uid": uids,
			}
		} else {
			ex = g.Ex{
				"uid": uid,
			}
		}
	}
	ex["bet_time"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	if flag != "-1" {
		f, err := strconv.Atoi(flag)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}
		ex["flag"] = f
	}

	if gameType > 0 {
		ex["game_type"] = gameType
	}

	if gameID > 0 {
		ex["api_type"] = gameID
	}

	query, _, _ := dialect.From("tbl_game_record").Select(g.COUNT("bill_no")).Where(ex).Limit(1).ToSQL()
	fmt.Println(query)
	err := meta.TiDB.Get(&data.T, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}
	if data.T == 0 {
		return data, nil
	}

	offset := (page - 1) * pageSize
	query, _, _ = dialect.From("tbl_game_record").Select(colsGameRecord...).Where(ex).Order(g.C("bet_time").Desc()).Offset(offset).Limit(pageSize).ToSQL()
	fmt.Println(query)
	err = meta.TiDB.Select(&data.D, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	for i := 0; i < len(data.D); i++ {
		data.D[i].ApiTypes = fmt.Sprintf(`%d`, data.D[i].ApiType)
	}

	return data, nil
}

func SubTradeRecord(uid, playerName string, dateType, flag int, pageSize, page uint) (TradeData, error) {

	data := TradeData{}

	//????????????
	var startAt int64
	endAt := helper.DayTET(0, loc).Unix()

	switch dateType {
	case 1: //??????
		startAt = helper.DayTST(0, loc).Unix()
	case 2: //??????
		startAt = helper.DayTST(0, loc).Unix() - 24*60*60
		endAt = helper.DayTET(0, loc).Unix() - 24*60*60
	case 3: //??????
		startAt = helper.DayTST(0, loc).Unix() - 6*24*60*60
	default:
		startAt = helper.DayTST(0, loc).Unix()
	}

	ex := g.Ex{}
	if playerName != "" && validator.CheckUName(playerName, 5, 14) {
		//????????????????????????????????????????????????????????????
		var count int64
		mb, err := MemberCache(nil, playerName)
		if err != nil {
			return data, errors.New(helper.UsernameExist)
		}
		ex = g.Ex{
			"ancestor":   uid,
			"descendant": mb.UID,
			"prefix":     meta.Prefix,
		}
		query, _, _ := dialect.From("tbl_members_tree").Select(g.COUNT("*")).Where(ex).Limit(1).ToSQL()
		err = meta.MerchantDB.Get(&count, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}
		if count == 0 {
			return data, errors.New(helper.NotDirectSubordinate)
		}
		ex = g.Ex{
			"username": playerName,
		}
	} else {
		var uids []string
		ex = g.Ex{
			"top_uid":     uid,
			"report_time": g.Op{"between": exp.NewRangeVal(startAt, endAt)},
			"report_type": 2,
			"prefix":      meta.Prefix,
		}
		query, _, _ := dialect.From("tbl_report_sub_member").Select(g.C("uid")).Where(ex).GroupBy("uid").Limit(100).ToSQL()
		fmt.Println(query)
		err := meta.ReportDB.Select(&uids, query)
		if err != nil {
			return data, pushLog(err, helper.DBErr)
		}
		if len(uids) > 0 {
			uids = append(uids, uid)
			ex = g.Ex{
				"uid": uids,
			}
		} else {
			fmt.Println(uid)
			ex = g.Ex{
				"uid": uid,
			}
		}
	}
	switch flag {
	case RecordTradeWithdraw: // ??????
		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
		return recordTradeWithdraw(RecordTradeWithdraw, page, pageSize, ex)

	case RecordTradeDeposit: // ??????
		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
		return recordTradeDeposit(RecordTradeDeposit, page, pageSize, ex)
	default:
	}

	return data, errors.New(helper.GetDataFailed)
}

func AgencyReportList(ty string, fCtx *fasthttp.RequestCtx, playerName string, page, pageSize, isOnline int) (ReportAgencyData, error) {

	data := ReportAgencyData{}
	mb, err := MemberCache(fCtx, "")
	if err != nil {
		return data, errors.New(helper.AccessTokenExpires)
	}
	ex := g.Ex{}
	if len(playerName) > 0 {
		ex["parent_name"] = playerName
	} else {
		ex["parent_name"] = mb.Username
	}
	offset := (page - 1) * pageSize

	var startAt int64
	var reportType int
	endAt := helper.DayTET(0, loc).Unix()
	switch ty {
	case "1": //??????
		startAt = helper.DayTST(0, loc).Unix()
		reportType = 2
	case "2": //??????
		startAt = helper.DayTST(0, loc).Unix() - 24*60*60
		endAt = helper.DayTST(0, loc).Unix() - 1
		reportType = 2
	case "3": //??????
		startAt = helper.MonthTST(0, loc).Unix()
		reportType = 4
	case "4": //??????
		startAt = helper.MonthTST(helper.MonthTST(0, loc).Unix()-1, loc).Unix()
		reportType = 4
	case "5": //??????
		startAt = helper.DayTST(0, loc).AddDate(0, 0, -2).Unix()
		endAt = helper.DayTST(0, loc).Unix()
		reportType = 2
	case "6": //??????
		startAt = helper.DayTST(0, loc).AddDate(0, 0, -6).Unix()
		endAt = helper.DayTST(0, loc).Unix()
		reportType = 2
	default:
		startAt = helper.DayTST(0, loc).Unix()
		reportType = 2
	}
	// ??????????????????
	ex["report_type"] = reportType
	ex["prefix"] = meta.Prefix
	ex["report_time"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	ex["data_type"] = 1
	if isOnline == 1 {
		uids, err := allOnline()
		if err != nil {
			return data, pushLog(err, helper.RedisErr)
		}
		ex["uid"] = uids
	} else if isOnline == 2 {
		uids, err := allOnline()
		if err != nil {
			return data, pushLog(err, helper.RedisErr)
		}
		ex["uid"] = g.L("not in ?", uids)
	}

	query, _, _ := dialect.From("tbl_report_agency").Select(g.COUNT(g.DISTINCT("uid"))).Where(ex).Limit(1).ToSQL()
	query = strings.Replace(query, "= not in", "not in", 1)
	fmt.Println(query)
	err = meta.ReportDB.Get(&data.T, query)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err.Error())
		return data, pushLog(err, helper.DBErr)
	}

	if data.T == 0 {
		return data, nil
	}

	query, _, _ = dialect.From("tbl_report_agency").Where(ex).
		Select(
			g.C("uid").As("uid"),
			g.C("username").As("username"),
			g.C("lvl").As("lvl"),
			g.SUM("bet_mem_count").As("bet_mem_count"),     //????????????
			g.SUM("bet_amount").As("bet_amount"),           //????????????
			g.SUM("rebate_amount").As("rebate"),            //??????
			g.SUM("dividend_amount").As("dividend_amount"), //????????????
			g.SUM("win_amount").As("win_amount"),           //????????????
			g.SUM("cg_rebate").As("cg_rebate"),             //????????????
			g.SUM("company_revenue").As("profit"),          //??????
		).GroupBy("uid", "username", "lvl").Order(g.C("bet_mem_count").Desc()).Offset(uint(offset)).Limit(uint(pageSize)).
		ToSQL()
	query = strings.Replace(query, "= not in", "not in", 1)
	fmt.Println(query)
	err = meta.ReportDB.Select(&data.D, query)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err.Error())
		return data, pushLog(err, helper.DBErr)
	}
	for i := 0; i < len(data.D); i++ {
		data.D[i].Profit, _ = decimal.NewFromFloat(data.D[i].Profit).Mul(decimal.NewFromFloat(-1)).Float64()
	}

	for i := 0; i < len(data.D); i++ {
		ds, err := onlineDevices(data.D[i].Uid)
		if err != nil {
			pushLog(err, helper.RedisErr)
			return data, nil
		}

		if ds != "" {
			data.D[i].IsOnline = 1
		} else {
			data.D[i].IsOnline = 2
		}
	}

	return data, nil
}
