package model

import (
	"errors"
	"fmt"
	"member/contrib/helper"
	"member/contrib/validator"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/olivere/elastic/v7"
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
}

type TradeData struct {
	T   int64             `json:"t"`
	D   []trade           `json:"d"`
	S   uint              `json:"s"`
	Agg map[string]string `json:"agg"`
}

func RecordTransfer(username, billNo, state, transferType, pidIn, pidOut, startTime, endTime string, page, pageSize int) (TransferData, error) {

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
	param := map[string]interface{}{
		"username": username,
		"prefix":   meta.Prefix,
	}

	if billNo != "" {
		param["bill_no"] = billNo
	}

	if transferType != "" {
		param["transfer_type"] = transferType
	}

	if pidIn != "" && pidOut == "" {
		param["platform_id"] = pidIn
	}

	if pidIn == "" && pidOut != "" {
		param["platform_id"] = pidOut
	}

	if transferType == "" && pidIn != "" && pidOut != "" {
		param["platform_id"] = []interface{}{pidIn, pidOut}
	}

	if state != "" {
		param["state"] = state
	}

	rangeParam := map[string][]interface{}{
		"created_at": {startAt, endAt},
	}
	agg := map[string]string{
		"amount_agg": "amount",
	}
	sortFields := map[string]bool{
		"created_at": false,
	}
	data, err = esTransferQuery(esPrefixIndex("tbl_member_transfer"), sortFields, page, pageSize, param, rangeParam, agg)
	if err != nil {
		return data, err
	}

	return data, nil
}

// pull data from es
func RecordGame(ty int, uid, playerName, startTime, endTime string, flag, gameID int, pageSize, page int) (GameRecordData, error) {

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
	params := map[string]interface{}{}
	if ty == 1 {
		// 直属下级游戏记录
		params["parent_uid"] = uid
		if playerName != "" && validator.CheckUName(playerName, 5, 14) {
			params["player_name"] = playerName
		}
	} else { // 查自己的游戏记录
		params["uid"] = uid
	}

	if flag != -1 {
		params["flag"] = flag
	}

	if gameID > 0 {
		params["api_type"] = gameID
	}

	rangeParam := map[string][]interface{}{
		"bet_time": {startAt, endAt},
	}

	agg := map[string]string{
		"bet_amount_agg":       "bet_amount",
		"net_amount_agg":       "net_amount",
		"valid_bet_amount_agg": "valid_bet_amount",
	}

	sortFields := map[string]bool{
		"bet_time": false,
	}
	data, err = esGameRecordQuery(pullPrefixIndex("tbl_game_record"), sortFields, page, pageSize, params, rangeParam, agg)
	if err != nil {
		return data, err
	}

	return data, nil
}

// RecordTransaction es 账变记录列表
func RecordTransaction(uid, cashTypes, startTime, endTime string, page, pageSize int) (TransactionData, error) {

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

	param := map[string]interface{}{
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
		param["cash_type"] = types
	}

	rangeParam := map[string][]interface{}{
		"created_at": {startAt, endAt},
	}

	sortFields := map[string]bool{
		"created_at": false,
	}
	data, err = esTransactionQuery(esPrefixIndex("tbl_member_transaction"), sortFields, page, pageSize, param, rangeParam, map[string]string{})
	if err != nil {
		return data, err
	}

	return data, nil
}

//交易记录
func RecordTrade(uid, startTime, endTime string, flag, page, pageSize int) (TradeData, error) {

	data := TradeData{}
	startAtMs, err := helper.TimeToLocMs(startTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	endAtMs, err := helper.TimeToLocMs(endTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	startAt, err := helper.TimeToLoc(startTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	endAt, err := helper.TimeToLoc(endTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	if startAtMs >= endAtMs || startAt >= endAt {
		return data, errors.New(helper.QueryTimeRangeErr)
	}

	param := map[string]interface{}{"uid": uid}
	switch flag {
	case RecordTradeWithdraw: // 取款
		rangeParam := map[string][]interface{}{"created_at": {startAt, endAt}}
		aggField := map[string]string{"amount_agg": "amount"}
		return recordTradeWithdraw(flag, page, pageSize, param, rangeParam, aggField)

	case RecordTradeDeposit: // 存款
		rangeParam := map[string][]interface{}{"created_at": {startAt, endAt}}
		aggField := map[string]string{"amount_agg": "amount"}
		return recordTradeDeposit(flag, page, pageSize, param, rangeParam, aggField)

	case RecordTradeTransfer: // 转账
		rangeParam := map[string][]interface{}{"created_at": {startAtMs, endAtMs}}
		aggField := map[string]string{"amount_agg": "amount"}
		return recordTradeTransfer(flag, page, pageSize, param, rangeParam, aggField)

	case RecordTradeDividend: // 红利
		param["state"] = DividendReviewPass
		param["hand_out_state"] = DividendSuccess
		rangeParam := map[string][]interface{}{"apply_at": {startAtMs, endAtMs}}
		return recordTradeDividend(flag, page, pageSize, param, rangeParam, nil)

	case RecordTradeRebate: // 返水/佣金
		return recordTradeRebate(flag, page, pageSize, uid, startAtMs, endAtMs)

	case RecordTradeAdjust: // 加币 减币 调整
		return recordTradeAdjust(uid, flag, page, pageSize, startAt, endAt)
	}

	return data, errors.New(helper.GetDataFailed)
}

// 交易详情
func RecordTradeDetail(flag int, uid string, id string) (TradeData, error) {

	param := map[string]interface{}{
		"uid": uid,
	}
	meta.MerchantRedis.SetNX(ctx, "1", "1", 1000*time.Second).Result()
	data := TradeData{}
	switch flag {
	case RecordTradeWithdraw:
		//取款
		param["_id"] = id
		return recordTradeWithdraw(flag, 1, 15, param, nil, nil)

	case RecordTradeDeposit:
		//存款
		param["_id"] = id
		return recordTradeDeposit(flag, 1, 1, param, nil, nil)

	case RecordTradeTransfer:
		param["_id"] = id
		//转账
		aggField := map[string]string{"amount_agg": "amount"}
		return recordTradeTransfer(flag, 1, 15, param, nil, aggField)

	default:
		//红利 返水  调整
		return data, nil
	}
}

// 取款
func recordTradeWithdraw(flag, page, pageSize int,
	param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (TradeData, error) {

	data := TradeData{}
	param["prefix"] = meta.Prefix
	sortFields := map[string]bool{
		"created_at": false,
	}
	body, err := esWithdrawQuery(esPrefixIndex("tbl_withdraw"), sortFields, page, pageSize, param, rangeParam, aggField)
	if err != nil {
		return data, err
	}

	data.Agg = body.Agg
	data.T = body.T

	for _, v := range body.D {
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
		}

		data.D = append(data.D, item)
	}

	return data, nil
}

// 存款
func recordTradeDeposit(flag, page, pageSize int,
	param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (TradeData, error) {

	data := TradeData{}
	param["prefix"] = meta.Prefix
	sortFields := map[string]bool{
		"created_at": false,
	}
	body, err := esDepositQuery(esPrefixIndex("tbl_deposit"), sortFields, page, pageSize, param, rangeParam, aggField)
	if err != nil {
		return data, err
	}

	data.Agg = body.Agg
	data.T = body.T

	for _, v := range body.D {
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
		}

		data.D = append(data.D, item)
	}

	return data, nil
}

// 转账
func recordTradeTransfer(flag, page, pageSize int,
	param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (TradeData, error) {

	data := TradeData{}
	param["prefix"] = meta.Prefix
	sortFields := map[string]bool{
		"created_at": false,
	}
	body, err := esTransferQuery(esPrefixIndex("tbl_member_transfer"), sortFields, page, pageSize, param, rangeParam, aggField)
	if err != nil {
		return data, err
	}

	data.Agg = body.Agg
	data.T = body.T

	for _, v := range body.D {
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
func recordTradeDividend(flag, page, pageSize int,
	param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (TradeData, error) {

	data := TradeData{}
	param["prefix"] = meta.Prefix
	sortFields := map[string]bool{
		"apply_at": false,
	}
	body, err := esDividendQuery(esPrefixIndex("tbl_member_dividend"), sortFields, page, pageSize, param, rangeParam, aggField)
	if err != nil {
		return data, err
	}

	data.T = body.T
	for _, v := range body.D {
		item := trade{
			Flag:       flag,
			ID:         v.ID,
			Ty:         v.Wallet,
			BillNo:     v.ID,
			PlatformId: v.PlatformID,
			Amount:     fmt.Sprintf("%.4f", v.Amount),
			CreatedAt:  fmt.Sprintf("%d", v.ApplyAt),
			State:      v.State,
		}

		// 中心钱包
		if v.Wallet == 1 {
			// 中心钱包红利
			item.TransferType = helper.TransactionDividend
		} else { //场馆钱包
			// 场馆红利
			item.TransferType = TransferDividend
		}

		data.D = append(data.D, item)
	}

	return data, nil
}

// 返水
func recordTradeRebate(flag, page, pageSize int, uid string, startAt, endAt int64) (TradeData, error) {

	data := TradeData{}
	query := elastic.NewBoolQuery()

	query.Must(
		elastic.NewRangeQuery("created_at").Gte(startAt).Lte(endAt),
		elastic.NewTermQuery("uid", uid),
		elastic.NewBoolQuery().Should(
			elastic.NewTermQuery("cash_type", 161),
			elastic.NewTermQuery("cash_type", 170),
		),
	)

	total, esData, _, err := esQuerySearch(esPrefixIndex("tbl_balance_transaction"), "created_at",
		page, pageSize, colsEsCommissionTransaction, query, nil)

	if err != nil {
		return data, err
	}

	data.T = total
	for _, v := range esData {

		fmt.Println(string(v.Source))
		comm := BalanceTransaction{}
		_ = helper.JsonUnmarshal(v.Source, &comm)
		fmt.Println(comm)
		item := trade{
			Flag:         flag,
			ID:           v.Id,
			Ty:           1, // 佣金钱包
			BillNo:       v.Id,
			PlatformId:   "",
			TransferType: comm.CashType,
			Amount:       fmt.Sprintf(`%f`, comm.Amount),
			CreatedAt:    fmt.Sprintf(`%d`, comm.CreatedAt),
			State:        1,
		}

		data.D = append(data.D, item)
	}

	return data, nil
}

// 加币 减币 调整
func recordTradeAdjust(uid string, flag, page, pageSize int, startAt, endAt int64) (TradeData, error) {

	data := TradeData{}

	query := elastic.NewBoolQuery()

	query.Must(
		elastic.NewRangeQuery("apply_at").Gte(startAt).Lte(endAt),
		elastic.NewTermQuery("uid", uid),
		elastic.NewTermQuery("prefix", meta.Prefix),
		elastic.NewBoolQuery().Should(
			elastic.NewTermQuery("adjust_mode", AdjustDownMode),
			elastic.NewTermQuery("adjust_mode", AdjustUpMode),
		),
		elastic.NewBoolQuery().Must(
			elastic.NewTermQuery("state", AdjustReviewPass),
		))

	total, esData, _, err := esQuerySearch(esPrefixIndex("tbl_member_adjust"), "apply_at", page, pageSize, adjustFields, query, nil)
	if err != nil {
		return data, err
	}

	data.T = total
	for _, v := range esData {

		adjust := MemberAdjust{}
		adjust.ID = v.Id
		_ = helper.JsonUnmarshal(v.Source, &adjust)

		t := trade{
			Flag:       flag,
			ID:         adjust.ID,
			Ty:         1,
			BillNo:     adjust.ID,
			PlatformId: "1",
			Amount:     fmt.Sprintf("%.4f", adjust.Amount),
			CreatedAt:  fmt.Sprintf("%d", adjust.ApplyAt),
			State:      adjust.State,
		}

		if adjust.AdjustMode == AdjustUpMode {
			t.TransferType = helper.TransactionUpPoint
		} else {
			t.TransferType = helper.TransactionDownPoint
		}

		data.D = append(data.D, t)
	}

	return data, nil
}

func CheckSmsCaptcha(ip, sid, phone, code string) error {

	key := fmt.Sprintf("%s:sms:%s%s%s", meta.Prefix, phone, ip, sid)
	fmt.Println("CheckSmsCaptcha", key)
	val, err := meta.MerchantRedis.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		_ = pushLog(err, helper.RedisErr)
		return errors.New(helper.PhoneVerificationErr)
	}

	if code == val {
		return nil
	}

	return errors.New(helper.PhoneVerificationErr)
}

// starc 会员列表查询执行
func EsMemberList(page, pageSize int, ascending bool, username, startTime, endTime, sortField string, query *elastic.BoolQuery) (MemberListData, error) {

	data := MemberListData{}
	if startTime != "" && endTime != "" {

		startAt, err := helper.TimeToLoc(startTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		endAt, err := helper.TimeToLoc(endTime, loc)
		if err != nil {
			return data, errors.New(helper.TimeTypeErr)
		}

		if startAt >= endAt {
			return data, errors.New(helper.QueryTimeRangeErr)
		}
		query.Filter(elastic.NewRangeQuery("created_at").Gte(startAt).Lte(endAt))
	}
	data.S = pageSize

	query.Filter(elastic.NewTermQuery("prefix", meta.Prefix))
	var t int64
	var esResult []*elastic.SearchHit
	var err2 error

	if sortField != "" && username == "" {
		t, esResult, _, err2 = EsMemberListSort(
			esPrefixIndex("tbl_report_agency"), sortField, ascending, page, pageSize, reportAgencyListFields, query, nil)

		if err2 != nil {
			return data, pushLog(err2, helper.DBErr)
		}
	} else {
		t, esResult, _, err2 = EsMemberListSearch(
			esPrefixIndex("tbl_members"), "created_at", page, pageSize, []string{"uid", "username"}, query, nil)

		if err2 != nil {
			return data, pushLog(err2, helper.DBErr)
		}
	}

	var names []string
	data.T = int(t)
	for _, v := range esResult {

		record := MemberListCol{}
		_ = helper.JsonUnmarshal(v.Source, &record)
		data.D = append(data.D, record)
		names = append(names, record.Username)
	}

	if len(data.D) == 0 {
		return data, nil
	}

	// 获取用户的反水比例
	var ids []string
	for _, v := range data.D {
		ids = append(ids, v.UID)
	}
	rebates, err := MemberRebateExistRedis(ids)
	if err != nil {
		return data, err
	}

	for i, v := range data.D {
		if rb, ok := rebates[v.UID]; ok {
			data.D[i].DJ = rb.DJ
			data.D[i].TY = rb.TY
			data.D[i].ZR = rb.ZR
			data.D[i].QP = rb.QP
			data.D[i].DZ = rb.DZ
			data.D[i].CP = rb.CP
			data.D[i].FC = rb.FC
			data.D[i].BY = rb.BY
			data.D[i].CGHighRebate = rb.CGHighRebate
			data.D[i].CGOfficialRebate = rb.CGOfficialRebate
		}
	}
	return data, nil
}

func MemberRebateExistRedis(ids []string) (map[string]MemberRebate, error) {
	pipe := meta.MerchantRedis.Pipeline()
	defer pipe.Close()
	mm := make(map[string]MemberRebate)
	//// 任何一个错误的id 都将返回一个错误
	for i, idd := range ids {
		m, ee := MemberRebateGetCache(idd)
		if ee != nil {
			msg := fmt.Sprintf("%d ,errtype:%s", i, helper.RedisErr)
			return nil, pushLog(ee, msg)
		} else {
			mm[idd] = m
		}

	}
	return mm, nil
}
