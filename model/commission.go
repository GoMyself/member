package model

import (
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/olivere/elastic/v7"
	"github.com/shopspring/decimal"
	"github.com/valyala/fasthttp"
	"member2/contrib/helper"
	"time"
)

func CommissionDraw(withdrawPwd, sAmount string, fCtx *fasthttp.RequestCtx) error {

	username := string(fCtx.UserValue("token").([]byte))
	if username == "" {
		return errors.New(helper.AccessTokenExpires)
	}

	mb, err := MemberFindOne(username)
	if err != nil {
		return errors.New(helper.AccessTokenExpires)
	}

	// 余额
	balance, _ := decimal.NewFromString(mb.Balance)
	// 佣金
	coAmount, _ := decimal.NewFromString(mb.Commission)
	amount, err := decimal.NewFromString(sAmount)
	if err != nil {
		return errors.New(helper.AmountErr)
	}

	// 余额不足
	if amount.GreaterThan(coAmount) {
		return errors.New(helper.LackOfBalance)
	}

	// 提款密码错误
	if mb.WithdrawPwd != MurmurHash(withdrawPwd, mb.CreatedAt) {
		return errors.New(helper.WithdrawPwdMismatch)
	}

	coAfter := coAmount.Sub(amount)
	balanceAfter := balance.Add(amount)
	id := helper.GenId()
	ts := time.Now()
	// 佣金账变记录
	coTrans := MemberTransaction{
		AfterAmount:  coAfter.String(),
		Amount:       amount.String(),
		BeforeAmount: coAmount.String(),
		BillNo:       id,
		CreatedAt:    ts.UnixMilli(),
		ID:           id,
		CashType:     COTransactionDraw,
		UID:          mb.UID,
		Username:     mb.Username,
		Prefix:       meta.Prefix,
	}
	// 佣金账变记录
	trans := MemberTransaction{
		AfterAmount:  balanceAfter.String(),
		Amount:       amount.String(),
		BeforeAmount: balance.String(),
		BillNo:       id,
		CreatedAt:    ts.UnixMilli(),
		ID:           id,
		CashType:     TransactionCommissionDraw,
		UID:          mb.UID,
		Username:     mb.Username,
	}

	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	query, _, _ := dialect.Insert("tbl_commission_transaction").Rows(coTrans).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	record := g.Record{
		"commission": g.L(fmt.Sprintf("commission-%s", amount.String())),
	}

	transfer := CommissionTransfer{
		ID:           id,
		UID:          mb.UID,
		Username:     mb.Username,
		ReceiveUID:   mb.UID,
		ReceiveName:  mb.Username,
		Amount:       amount.String(),
		TransferType: 2, //佣金提取
		CreatedAt:    ts.Unix(),
		State:        1, //审核中
		Automatic:    1,
		ReviewAt:     0,
		ReviewUID:    "0",
		ReviewName:   "",
		Prefix:       meta.Prefix,
	}
	// 不需要审核
	if meta.AutoCommission {
		query, _, _ = dialect.Insert("tbl_balance_transaction").Rows(trans).ToSQL()
		_, err = tx.Exec(query)
		if err != nil {
			_ = tx.Rollback()
			return pushLog(err, helper.DBErr)
		}

		transfer.State = 2
		transfer.ReviewAt = ts.Unix()

		record["balance"] = g.L(fmt.Sprintf("balance+%s", amount.String()))
	} else {
		record["lock_amount"] = g.L(fmt.Sprintf("lock_amount+%s", amount.String()))
		transfer.Automatic = 2
	}

	query, _, _ = dialect.Insert("tbl_commission_transfer").Rows(transfer).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	ex := g.Ex{
		"uid": mb.UID,
	}
	query, _, _ = dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

func CommissionRation(withdrawPwd, subName, sAmount string, fCtx *fasthttp.RequestCtx) error {

	username := string(fCtx.UserValue("token").([]byte))
	if username == "" {
		return errors.New(helper.AccessTokenExpires)
	}

	mb, err := MemberFindOne(username)
	if err != nil {
		return errors.New(helper.AccessTokenExpires)
	}

	smb, err := MemberFindOne(subName)
	if err != nil {
		return errors.New(helper.UsernameErr)
	}

	if smb.ParentUid != mb.UID {
		return errors.New(helper.NotDirectSubordinate)
	}

	// 余额
	balance, _ := decimal.NewFromString(mb.Balance)
	// 佣金
	coAmount, _ := decimal.NewFromString(mb.Commission)
	amount, err := decimal.NewFromString(sAmount)
	if err != nil {
		return errors.New(helper.AmountErr)
	}

	// 余额不足
	if amount.GreaterThan(coAmount) {
		return errors.New(helper.LackOfBalance)
	}

	// 提款密码错误
	if mb.WithdrawPwd != MurmurHash(withdrawPwd, mb.CreatedAt) {
		return errors.New(helper.WithdrawPwdMismatch)
	}

	coAfter := coAmount.Sub(amount)
	balanceAfter := balance.Add(amount)
	id := helper.GenId()
	ts := time.Now()
	// 佣金账变记录
	coTrans := MemberTransaction{
		AfterAmount:  coAfter.String(),
		Amount:       amount.String(),
		BeforeAmount: coAmount.String(),
		BillNo:       id,
		CreatedAt:    ts.UnixMilli(),
		ID:           id,
		CashType:     COTransactionRation,
		UID:          mb.UID,
		Username:     mb.Username,
	}
	// 佣金账变记录
	subCoTrans := MemberTransaction{
		AfterAmount:  balanceAfter.String(),
		Amount:       amount.String(),
		BeforeAmount: balance.String(),
		BillNo:       id,
		CreatedAt:    ts.UnixMilli(),
		ID:           id,
		CashType:     COTransactionReceive,
		UID:          smb.UID,
		Username:     smb.Username,
		Prefix:       meta.Prefix,
	}

	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	query, _, _ := dialect.Insert("tbl_commission_transaction").Rows(coTrans).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	record := g.Record{
		"commission": g.L(fmt.Sprintf("commission-%s", amount.String())),
	}

	transfer := CommissionTransfer{
		ID:           id,
		UID:          mb.UID,
		Username:     mb.Username,
		ReceiveUID:   smb.UID,
		ReceiveName:  smb.Username,
		Amount:       amount.String(),
		TransferType: 3, //佣金下发
		CreatedAt:    ts.Unix(),
		State:        1, //审核中
		Automatic:    1,
		ReviewAt:     0,
		ReviewUID:    "0",
		ReviewName:   "",
		Prefix:       meta.Prefix,
	}

	// 不需要审核
	if meta.AutoCommission {
		query, _, _ = dialect.Insert("tbl_commission_transaction").Rows(subCoTrans).ToSQL()
		_, err = tx.Exec(query)
		if err != nil {
			_ = tx.Rollback()
			return pushLog(err, helper.DBErr)
		}

		rec := g.Record{
			"commission": g.L(fmt.Sprintf("commission+%s", amount.String())),
		}
		query, _, _ = dialect.Update("tbl_members").Set(rec).Where(g.Ex{"uid": smb.UID}).ToSQL()
		_, err = tx.Exec(query)
		if err != nil {
			_ = tx.Rollback()
			return pushLog(err, helper.DBErr)
		}

		transfer.State = 2
		transfer.ReviewAt = ts.Unix()
	} else {
		record["lock_amount"] = g.L(fmt.Sprintf("lock_amount+%s", amount.String()))
		transfer.Automatic = 2
	}

	query, _, _ = dialect.Insert("tbl_commission_transfer").Rows(transfer).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	ex := g.Ex{
		"uid": mb.UID,
	}
	query, _, _ = dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

// 佣金记录
func CommissionRecord(uid, startTime, endTime string, flag, page, pageSize int) (CommissionTransactionData, error) {

	data := CommissionTransactionData{D: []CommissionTransaction{}}
	startAtMs, err := helper.TimeToLocMs(startTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	endAtMs, err := helper.TimeToLocMs(endTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	if startAtMs >= endAtMs {
		return data, errors.New(helper.QueryTimeRangeErr)
	}

	query := elastic.NewBoolQuery().Must(
		elastic.NewRangeQuery("created_at").Gte(startAtMs).Lte(endAtMs),
		elastic.NewTermQuery("uid", uid),
	)

	if flag != 0 {
		query = query.Must(
			elastic.NewTermQuery("cash_type", flag),
		)
	}

	agg := elastic.NewSumAggregation().Field("amount")
	total, esData, aggSum, err := esQuerySearch(esPrefixIndex("tbl_commission_transaction"), "created_at",
		page, pageSize, colsEsCommissionTransaction, query, map[string]*elastic.SumAggregation{
			"amount_agg": agg,
		})

	if err != nil {
		return data, err
	}

	if v, ok := aggSum.Sum("amount_agg"); ok {
		data.Agg, _ = decimal.NewFromFloat(*v.Value).Truncate(4).Float64()

	}

	data.T = total
	for _, v := range esData {
		comm := CommissionTransaction{}
		comm.ID = v.Id
		_ = helper.JsonUnmarshal(v.Source, &comm)
		if err != nil {
			return data, errors.New(helper.FormatErr)
		}
		data.D = append(data.D, comm)
	}

	return data, nil
}
