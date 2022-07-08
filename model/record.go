package model

import (
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"member/contrib/helper"
	"member/contrib/validator"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type trade struct {
	Flag         int    `json:"flag"`          //前端传入 1 存款 2 取款 3 转账 4 红利 5 反水 6 加币 7 减币 8 调整
	ID           string `json:"id"`            //id
	Ty           int    `json:"ty"`            //1 中心钱包 2 场馆转账 3 佣金钱包
	BillNo       string `json:"bill_no"`       //流水号
	PlatformId   string `json:"platform_id"`   //三方场馆ID' ty = 2 需要根据这个id判断是哪个场馆，可能在详情展示时使用
	TransferType int    `json:"transfer_type"` //ty = 2   转账类型(0:转入 1:转出 2:后台上分 3:场馆钱包清零 4:场馆红利 )', ty = 1 0:转入1:转出2:转入失败补回3:转出失败扣除4:存款5:提现 6:后台上分 7:后台下分 8:后台下分回退 9 红利派发
	Amount       string `json:"amount"`        //金额
	CreatedAt    string `json:"created_at"`    //创建时间
	State        int    `json:"state"`         //0:失败1:成功2:处理中3:脚本确认中4:人工确认中',  只有ty = 2时需要判断
	Remark       string `json:"remark"`
	Username     string `json:"username"`
	Balance      string `json:"balance"`
}

type TradeData struct {
	T   int64             `json:"t"`
	D   []trade           `json:"d"`
	S   uint              `json:"s"`
	Agg map[string]string `json:"agg"`
}

func RecordTransfer(username, billNo, state, transferType, pidIn, pidOut, startTime, endTime string, page, pageSize uint) (TransferData, error) {

	data := TransferData{}
	//判断日期
	startAt, err := helper.TimeToLocMs(startTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	endAt, err := helper.TimeToLocMs(endTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	if startAt >= endAt {
		return data, errors.New(helper.QueryTimeRangeErr)
	}

	//查询条件
	ex := g.Ex{
		"username": username,
		"prefix":   meta.Prefix,
	}

	if billNo != "" {
		ex["bill_no"] = billNo
	}

	if transferType != "" {
		ex["transfer_type"] = transferType
	}

	if pidIn != "" && pidOut == "" {
		ex["platform_id"] = pidIn
	}

	if pidIn == "" && pidOut != "" {
		ex["platform_id"] = pidOut
	}

	if transferType == "" && pidIn != "" && pidOut != "" {
		ex["platform_id"] = []interface{}{pidIn, pidOut}
	}

	if state != "" {
		ex["state"] = state
	}

	ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	query, _, _ := dialect.From("tbl_member_transfer").Select(g.COUNT("id")).Where(ex).Limit(1).ToSQL()
	fmt.Println(query)
	err = meta.TiDB.Get(&data.T, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}
	if data.T == 0 {
		return data, nil
	}

	offset := (page - 1) * pageSize
	query, _, _ = dialect.From("tbl_member_transfer").Select(colsTransfer...).Where(ex).Order(g.C("created_at").Desc()).Offset(offset).Limit(pageSize).ToSQL()
	fmt.Println(query)
	err = meta.TiDB.Select(&data.D, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

// pull data from es
func RecordGame(ty int, uid, playerName, startTime, endTime string, flag, gameID int, pageSize, page uint) (GameRecordData, error) {

	data := GameRecordData{}
	//判断日期
	startAt, err := helper.TimeToLocMs(startTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	endAt, err := helper.TimeToLocMs(endTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}
	if startAt >= endAt {
		return data, errors.New(helper.QueryTimeRangeErr)
	}

	//查询条件
	ex := g.Ex{}
	if ty == 1 {
		// 直属下级游戏记录
		ex["parent_uid"] = uid
		if playerName != "" && validator.CheckUName(playerName, 5, 14) {
			ex["player_name"] = playerName
		}
	} else { // 查自己的游戏记录
		ex["uid"] = uid
	}

	if flag != -1 {
		ex["flag"] = flag
	}

	if gameID > 0 {
		ex["api_type"] = gameID
	}

	ex["bet_time"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}

	query, _, _ := dialect.From("tbl_game_record").Select(g.COUNT("bill_no")).Where(ex).Limit(1).ToSQL()
	fmt.Println(query)
	err = meta.TiDB.Get(&data.T, query)
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

	return data, nil
}

// RecordTransaction es 账变记录列表
func RecordTransaction(uid, cashTypes, startTime, endTime string, page, pageSize uint) (TransactionData, error) {

	data := TransactionData{}
	startAt, err := helper.TimeToLocMs(startTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	endAt, err := helper.TimeToLocMs(endTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	if startAt >= endAt {
		return data, errors.New(helper.QueryTimeRangeErr)
	}

	ex := g.Ex{
		"uid": uid,
	}

	// 账变类型筛选
	if cashTypes != "" {
		var types []interface{}
		for _, v := range strings.Split(cashTypes, ",") {
			if !validator.CheckStringLength(v, 0, 5) {
				continue
			}
			types = append(types, v)
		}
		ex["cash_type"] = types
	}

	ex["prefix"] = meta.Prefix
	query, _, _ := dialect.From("tbl_member_transaction").Select(g.COUNT("id")).Where(ex).Limit(1).ToSQL()
	fmt.Println(query)
	err = meta.TiDB.Get(&data.T, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}
	if data.T == 0 {
		return data, nil
	}

	offset := (page - 1) * pageSize
	query, _, _ = dialect.From("tbl_member_transaction").Select(colsTransation...).Where(ex).Order(g.C("created_at").Desc()).Offset(offset).Limit(pageSize).ToSQL()
	fmt.Println(query)
	err = meta.TiDB.Select(&data.D, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

//交易记录
func RecordTrade(uid, startTime, endTime string, flag int, page, pageSize uint) (TradeData, error) {

	data := TradeData{}
	startAtMs, err := helper.TimeToLocMs(startTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	endAtMs := helper.DaySET(endTime, loc).UnixMilli()

	startAt, err := helper.TimeToLoc(startTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	endAt := helper.DaySET(endTime, loc).Unix()
	fmt.Println(endAt)
	if startAtMs >= endAtMs || startAt >= endAt {
		return data, errors.New(helper.QueryTimeRangeErr)
	}

	ex := g.Ex{"uid": uid}
	switch flag {
	case RecordTradeWithdraw: // 取款
		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
		return recordTradeWithdraw(flag, page, pageSize, ex)

	case RecordTradeDeposit: // 存款
		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
		return recordTradeDeposit(flag, page, pageSize, ex)

	case RecordTradeTransfer: // 转账
		ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAtMs, endAtMs)}
		return recordTradeTransfer(flag, page, pageSize, ex)

	case RecordTradeDividend: // 红利
		ex["state"] = DividendReviewPass
		ex["review_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
		return recordTradeDividend(flag, page, pageSize, ex)

	case RecordTradeRebate: // 返水/佣金
		return recordTradeRebate(flag, page, pageSize, uid, startAtMs, endAtMs)

	case RecordTradeAdjust: // 加币 减币 调整
		return recordTradeAdjust(uid, flag, page, pageSize, startAt, endAt)
	}

	return data, errors.New(helper.GetDataFailed)
}

// 交易详情
func RecordTradeDetail(flag int, uid string, id string) (TradeData, error) {

	ex := g.Ex{"uid": uid}
	meta.MerchantRedis.SetNX(ctx, "1", "1", 1000*time.Second).Result()
	data := TradeData{}
	switch flag {
	case RecordTradeWithdraw:
		//取款
		ex["id"] = id
		return recordTradeWithdraw(flag, uint(1), 15, ex)

	case RecordTradeDeposit:
		//存款
		ex["id"] = id
		return recordTradeDeposit(flag, uint(1), uint(1), ex)

	case RecordTradeTransfer:
		ex["id"] = id
		//转账
		return recordTradeTransfer(flag, uint(1), uint(15), ex)

	default:
		//红利 返水  调整
		return data, nil
	}
}

// 取款
func recordTradeWithdraw(flag int, page, pageSize uint,
	ex g.Ex) (TradeData, error) {

	data := TradeData{}
	var list []Withdraw
	ex["prefix"] = meta.Prefix
	query, _, _ := dialect.From("tbl_withdraw").Select(g.COUNT("id")).Where(ex).Limit(1).ToSQL()
	fmt.Println(query)
	err := meta.TiDB.Get(&data.T, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}
	if data.T == 0 {
		return data, nil
	}

	offset := (page - 1) * pageSize
	query, _, _ = dialect.From("tbl_withdraw").Select(colsWithdraw...).Where(ex).Order(g.C("created_at").Desc()).Offset(offset).Limit(pageSize).ToSQL()
	fmt.Println(query)
	err = meta.TiDB.Select(&list, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	for _, v := range list {
		item := trade{
			Flag:         flag,
			ID:           v.ID,
			Ty:           1,
			BillNo:       v.OID,
			PlatformId:   "0",
			TransferType: helper.TransactionWithDraw,
			Amount:       fmt.Sprintf("%.4f", v.Amount),
			CreatedAt:    fmt.Sprintf("%d", v.CreatedAt),
			State:        v.State,
			Remark:       v.WithdrawRemark,
			Username:     v.Username,
		}
		mb, err := MemberCache(nil, v.Username)
		if err != nil {
			return data, err
		}
		item.Balance = mb.Balance

		data.D = append(data.D, item)
	}

	return data, nil
}

// 存款
func recordTradeDeposit(flag int, page, pageSize uint,
	ex g.Ex) (TradeData, error) {

	data := TradeData{}
	var list []Deposit
	ex["prefix"] = meta.Prefix
	fmt.Println(ex["uid"])
	query, _, _ := dialect.From("tbl_deposit").Select(g.COUNT("id")).Where(ex).Limit(1).ToSQL()
	fmt.Println(query)
	err := meta.TiDB.Get(&data.T, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}
	if data.T == 0 {
		return data, nil
	}

	offset := (page - 1) * pageSize
	query, _, _ = dialect.From("tbl_deposit").Select(colsDeposit...).Where(ex).Order(g.C("created_at").Desc()).Offset(offset).Limit(pageSize).ToSQL()
	fmt.Println(query)
	err = meta.TiDB.Select(&list, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	for _, v := range list {
		item := trade{
			Flag:         flag,
			ID:           v.ID,
			Ty:           1,
			BillNo:       v.OID,
			PlatformId:   v.ChannelID,
			TransferType: helper.TransactionDeposit,
			Amount:       fmt.Sprintf("%.4f", v.Amount),
			CreatedAt:    fmt.Sprintf("%d", v.CreatedAt),
			State:        v.State,
			Username:     v.Username,
		}

		data.D = append(data.D, item)
	}

	return data, nil
}

// 转账
func recordTradeTransfer(flag int, page, pageSize uint,
	ex g.Ex) (TradeData, error) {

	data := TradeData{}
	var list []Transfer
	ex["prefix"] = meta.Prefix
	query, _, _ := dialect.From("tbl_member_transfer").Select(g.COUNT("id")).Where(ex).Limit(1).ToSQL()
	fmt.Println(query)
	err := meta.TiDB.Get(&data.T, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}
	if data.T == 0 {
		return data, nil
	}

	offset := (page - 1) * pageSize
	query, _, _ = dialect.From("tbl_member_transfer").Select(colsTransfer...).Where(ex).Order(g.C("created_at").Desc()).Offset(offset).Limit(pageSize).ToSQL()
	fmt.Println(query)
	err = meta.TiDB.Select(&list, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	for _, v := range list {
		item := trade{
			Flag:         flag,
			ID:           v.ID,
			Ty:           2,
			BillNo:       v.BillNo,
			PlatformId:   v.PlatformId,
			TransferType: v.TransferType,
			Amount:       fmt.Sprintf("%.4f", v.Amount),
			CreatedAt:    fmt.Sprintf("%d", v.CreatedAt),
			State:        v.State,
		}

		data.D = append(data.D, item)
	}

	return data, nil
}

// 红利
func recordTradeDividend(flag int, page, pageSize uint,
	ex g.Ex) (TradeData, error) {

	data := TradeData{}
	ex["prefix"] = meta.Prefix
	var list []Dividend
	query, _, _ := dialect.From("tbl_member_dividend").Select(g.COUNT("id")).Where(ex).Limit(1).ToSQL()
	fmt.Println(query)
	err := meta.TiDB.Get(&data.T, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}
	if data.T == 0 {
		return data, nil
	}

	offset := (page - 1) * pageSize
	query, _, _ = dialect.From("tbl_member_dividend").Select(colsDividend...).Where(ex).Order(g.C("review_at").Desc()).Offset(offset).Limit(pageSize).ToSQL()
	fmt.Println(query)
	err = meta.TiDB.Select(&list, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	for _, v := range list {
		item := trade{
			Flag:       flag,
			ID:         v.ID,
			Ty:         1,
			BillNo:     v.ID,
			PlatformId: "",
			Amount:     fmt.Sprintf("%.4f", v.Amount),
			CreatedAt:  fmt.Sprintf("%d", v.ApplyAt),
			State:      v.State,
		}

		item.TransferType = helper.TransactionDividend

		data.D = append(data.D, item)
	}

	return data, nil
}

// 返水
func recordTradeRebate(flag int, page, pageSize uint, uid string, startAt, endAt int64) (TradeData, error) {

	data := TradeData{}
	ex := g.Ex{
		"uid":       uid,
		"cash_type": []int{161, 170, 642, 643, 644, 645, 646, 647, 648, 649},
	}
	ex["created_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}

	var list []BalanceTransaction
	query, _, _ := dialect.From("tbl_balance_transaction").Select(g.COUNT("id")).Where(ex).Limit(1).ToSQL()
	fmt.Println(query)
	err := meta.TiDB.Get(&data.T, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}
	if data.T == 0 {
		return data, nil
	}

	offset := (page - 1) * pageSize
	query, _, _ = dialect.From("tbl_balance_transaction").Select(colsBalanceTransaction...).Where(ex).Order(g.C("created_at").Desc()).Offset(offset).Limit(pageSize).ToSQL()
	fmt.Println(query)
	err = meta.TiDB.Select(&list, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	for _, v := range list {

		item := trade{
			Flag:         flag,
			ID:           v.BillNo,
			Ty:           1, // 佣金钱包
			BillNo:       v.BillNo,
			PlatformId:   "",
			TransferType: v.CashType,
			Amount:       fmt.Sprintf(`%f`, v.Amount),
			CreatedAt:    fmt.Sprintf(`%d`, v.CreatedAt),
			State:        1,
		}

		data.D = append(data.D, item)
	}

	return data, nil
}

// 加币 减币 调整
func recordTradeAdjust(uid string, flag int, page, pageSize uint, startAt, endAt int64) (TradeData, error) {

	data := TradeData{}
	ex := g.Ex{
		"uid":    uid,
		"prefix": meta.Prefix,
	}
	ex["adjust_mode"] = []int{AdjustDownMode, AdjustUpMode}
	ex["state"] = AdjustReviewPass
	ex["apply_at"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}

	var list []MemberAdjust
	query, _, _ := dialect.From("tbl_member_adjust").Select(g.COUNT("id")).Where(ex).Limit(1).ToSQL()
	fmt.Println(query)
	err := meta.TiDB.Get(&data.T, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}
	if data.T == 0 {
		return data, nil
	}

	offset := (page - 1) * pageSize
	query, _, _ = dialect.From("tbl_member_adjust").Select(colsAdjust...).Where(ex).Order(g.C("apply_at").Desc()).Offset(offset).Limit(pageSize).ToSQL()
	fmt.Println(query)
	err = meta.TiDB.Select(&list, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	for _, v := range list {

		t := trade{
			Flag:       flag,
			ID:         v.ID,
			Ty:         1,
			BillNo:     v.ID,
			PlatformId: "1",
			Amount:     fmt.Sprintf("%.4f", v.Amount),
			CreatedAt:  fmt.Sprintf("%d", v.ApplyAt),
			State:      v.State,
		}

		if v.AdjustMode == AdjustUpMode {
			t.TransferType = helper.TransactionUpPoint
		} else {
			t.TransferType = helper.TransactionDownPoint
		}

		data.D = append(data.D, t)
	}

	return data, nil
}

func CheckSmsCaptcha(ip, sid, phone, code string) error {

	key := fmt.Sprintf("%s:sms:%s%s", meta.Prefix, phone, sid)
	cmd := meta.MerchantRedis.Get(ctx, key)
	val, err := cmd.Result()
	if err != nil && err != redis.Nil {
		_ = pushLog(err, helper.RedisErr)
		return errors.New(helper.PhoneVerificationErr)
	}

	if code == val {
		return nil
	}

	return errors.New(helper.PhoneVerificationErr)
}

//func MemberRebateExistRedis(ids []string) (map[string]MemberRebate, error) {
//	pipe := meta.MerchantRedis.Pipeline()
//	defer pipe.Close()
//	mm := make(map[string]MemberRebate)
//	//// 任何一个错误的id 都将返回一个错误
//	for i, idd := range ids {
//		m, ee := MemberRebateGetCache(idd)
//		if ee != nil {
//			msg := fmt.Sprintf("%d ,errtype:%s", i, helper.RedisErr)
//			return nil, pushLog(ee, msg)
//		} else {
//			mm[idd] = m
//		}
//
//	}
//	return mm, nil
//}
