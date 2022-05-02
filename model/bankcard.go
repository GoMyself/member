package model

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"bitbucket.org/nwf2013/schema"
	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"

	"member2/contrib/helper"
	"member2/contrib/validator"
)

func BankcardInsert(ctx *fasthttp.RequestCtx, phone, sid, code, realname, bankcardNo string, data BankCard) error {

	if len(bankcardNo) < 6 || len(bankcardNo) > 20 {
		return errors.New(helper.BalanceErr)
	}

	mb, err := MemberCache(ctx, "")
	if err != nil {
		return err
	}

	// 未绑定手机号 且 手机号不传  提示：请先绑定手机
	if phone == "" && mb.PhoneHash == "0" {
		return errors.New(helper.UserIDCheckInvalidPhone)
	}

	// 未绑定手机号 需要验证手机号是否已经被绑定
	if phone != "" && mb.PhoneHash == "0" {
		if !validator.IsVietnamesePhone(phone) {
			return errors.New(helper.PhoneFMTErr)
		}

		phoneHash := fmt.Sprintf("%d", MurmurHash(phone, 0))
		ex := g.Ex{
			"phone_hash": phoneHash,
		}

		if MemberBindCheck(ex) {
			return errors.New(helper.PhoneExist)
		}
	}

	// 判断卡号是否存在
	bankcardHash := MurmurHash(bankcardNo, 0)
	idx := bankcardHash % 10
	key := fmt.Sprintf("bl:bc%d", idx)
	ok, err := meta.MerchantRedis.SIsMember(ctx, key, bankcardHash).Result()
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	if ok {
		return errors.New(helper.BankcardBan)
	}

	var encRes []schema.Enc_t
	if phone == "" {
		var res []schema.Dec_t
		recs := schema.Dec_t{
			Field: "phone",
			Hide:  false,
			ID:    mb.UID,
		}

		res = append(res, recs)
		phoneRes, err := rpcGet(res)
		if err != nil && phoneRes[0].Err == "" {
			return errors.New(helper.GetRPCErr)
		}
		phone = phoneRes[0].Res
	}

	// 判断验证码
	ip := helper.FromRequest(ctx)
	err = emailCmp(sid, code, ip, phone)
	if err != nil {
		return err
	}

	// 判断会员银行卡数目
	if mb.BankcardTotal >= 3 {
		return errors.New(helper.MaxThreeBankCard)
	}

	cardNoHash := fmt.Sprintf("%d", bankcardHash)
	ex := g.Ex{
		"bank_card_hash": cardNoHash,
	}
	bcd, err := BankCardFindOne(ex)
	if err != nil {
		return err
	}

	var isDelete bool
	// 存在记录
	if bcd.UID != "" {

		// 正常和冻结的银行卡
		if bcd.State == "1" || bcd.State == "3" {
			return errors.New(helper.BankCardExistErr)
		}

		// 删除状态直接恢复
		if bcd.State == "2" {
			isDelete = true
		}
	}

	ex = g.Ex{
		"uid": mb.UID,
	}
	record := g.Record{
		"bankcard_total": g.L("bankcard_total+1"),
	}
	// 会员未绑定真实姓名，更新第一次绑定银行卡的真实姓名到会员信息
	if mb.RealnameHash == "0" {
		// 第一次新增银行卡判断真实姓名是否为越南文
		if !validator.CheckStringVName(realname) {
			return errors.New(helper.RealNameFMTErr)
		}

		recs := schema.Enc_t{
			Field: "realname",
			Value: realname,
			ID:    mb.UID,
		}
		encRes = append(encRes, recs)

		realNameHash := fmt.Sprintf("%d", MurmurHash(realname, 0))
		// 会员信息更新真实姓名和真实姓名hash
		record["realname_hash"] = realNameHash
	}

	// 会员未绑定手机号，更新手机号和手机号hash
	if phone != "" && mb.PhoneHash == "0" {
		recs := schema.Enc_t{
			Field: "phone",
			Value: phone,
			ID:    mb.UID,
		}
		encRes = append(encRes, recs)
		record["phone_hash"] = fmt.Sprintf("%d", MurmurHash(phone, 0))
	}

	recs := schema.Enc_t{
		Field: "bankcard",
		Value: bankcardNo,
		ID:    data.ID,
	}
	encRes = append(encRes, recs)
	_, err = rpcInsert(encRes)
	if err != nil {
		return errors.New(helper.UpdateRPCErr)
	}

	bc := g.Record{
		"id":               data.ID,
		"prefix":           meta.Prefix,
		"uid":              mb.UID,
		"username":         data.Username,
		"bank_address":     data.BankAddress,
		"bank_id":          data.BankID,
		"bank_branch_name": data.BankAddress,
		"bank_card_hash":   cardNoHash,
		"created_at":       data.CreatedAt,
	}

	// 会员银行卡插入加锁
	key = fmt.Sprintf("bc:%s", mb.Username)
	err = Lock(key)
	if err != nil {
		return err
	}

	defer Unlock(key)

	//开启事务
	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	if isDelete {
		query := fmt.Sprintf("delete from tbl_member_bankcard where bank_card_hash = %s and state = 2", cardNoHash)
		_, err = tx.Exec(query)
		if err != nil {
			_ = tx.Rollback()
			return pushLog(err, helper.DBErr)
		}
	}

	// 更新会员银行卡信息
	queryInsert, _, _ := dialect.Insert("tbl_member_bankcard").Rows(bc).ToSQL()
	_, err = tx.Exec(queryInsert)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	// 更新会员信息
	queryUpdate, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = tx.Exec(queryUpdate)
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

func BankCardFindOne(ex g.Ex) (BankCard, error) {

	data := BankCard{}
	ex["prefix"] = meta.Prefix
	t := dialect.From("tbl_member_bankcard")
	query, _, _ := t.Select(colsBankcard...).Where(ex).Order(g.C("state").Asc()).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&data, query)
	if err != nil && err != sql.ErrNoRows {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func BankcardList(username string) ([]BankcardData, error) {

	var (
		data     []BankcardData
		cardList []BankCard
	)
	mb, err := MemberCache(nil, username)
	if err != nil {
		return data, errors.New(helper.AccessTokenExpires)
	}

	key := "lk:" + mb.UID
	bcs, err := meta.MerchantRedis.Do(ctx, "JSON.GET", key, ".").Text()
	if err != nil {
		return data, pushLog(err, helper.RedisErr)
	}

	mp := map[string]BankCard{}
	err = helper.JsonUnmarshal([]byte(bcs), &mp)
	if err != nil {
		return data, pushLog(err, helper.FormatErr)
	}

	for _, v := range mp {
		if v.State == "1" {
			cardList = append(cardList, v)
		}
	}

	length := len(cardList)
	if length == 0 {
		return data, nil
	}

	var res []schema.Dec_t
	for _, v := range cardList {
		recs := schema.Dec_t{
			Field: "bankcard",
			Hide:  true,
			ID:    v.ID,
		}
		res = append(res, recs)
	}

	recs := schema.Dec_t{
		Field: "realname",
		Hide:  true,
		ID:    mb.UID,
	}
	res = append(res, recs)
	record, err := rpcGet(res)
	if err != nil {
		return data, errors.New(helper.GetRPCErr)
	}

	rpcLen := len(record)
	// 会员段返回解密银行卡真实姓名和银行卡号
	for k, v := range cardList {
		card := record[k].Res
		if rpcLen > k && record[k].Err == "" {
			card = record[k].Res
		}
		realName := ""
		if rpcLen > length && record[length].Err == "" {
			realName = record[length].Res
		}
		val := BankcardData{BankCard: v, RealName: realName, Bankcard: card}
		data = append(data, val)
	}

	return data, nil
}

func BankcardDelete(username, bid string) error {

	// 获取会员真实姓名
	mb, err := MemberCache(nil, username)
	if err != nil {
		return err
	}

	if mb.UID == "" {
		return errors.New(helper.AccessTokenExpires)
	}

	ex := g.Ex{
		"id":    bid,
		"state": []int{1, 3},
	}
	data, err := BankCardFindOne(ex)
	if err != nil {
		return err
	}

	if data.Username != username {
		return errors.New(helper.BankCardNotExist)
	}

	// 如果该银行卡有用于提款并且提款未完成的时候不允许删除
	err = bankcardDeleteCheck(data.UID, data.ID)
	if err != nil {
		return err
	}

	// 删除冻结的银行卡，直接删除
	if data.State == "3" {

		tx, err := meta.MerchantDB.Begin()
		if err != nil {
			return pushLog(err, helper.DBErr)
		}

		t := dialect.Update("tbl_member_bankcard")
		query, _, _ := t.Set(g.Record{"state": 2}).Where(g.Ex{"id": bid}).ToSQL()
		_, err = tx.Exec(query)
		if err != nil {
			_ = tx.Rollback()
			return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
		}

		record := g.Record{
			"bankcard_total": g.L("bankcard_total-1"),
		}
		query, _, _ = dialect.Update("tbl_members").Set(record).Where(g.Ex{"uid": mb.UID}).ToSQL()
		_, err = tx.Exec(query)
		if err != nil {
			_ = tx.Rollback()
			return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
		}

		err = tx.Commit()
		if err != nil {
			return pushLog(err, helper.DBErr)
		}

		return nil
	}

	bankcardHash, _ := strconv.ParseUint(data.BankcardHash, 10, 64)
	//开启事务
	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	t := dialect.Update("tbl_member_bankcard")
	query, _, _ := t.Set(g.Record{"state": 2}).Where(g.Ex{"id": bid}).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	query, _, _ = dialect.Update("tbl_members").
		Set(g.Record{"bankcard_total": g.L("bankcard_total-1")}).Where(g.Ex{"username": username, "prefix": meta.Prefix}).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	value := fmt.Sprintf("%d", bankcardHash)
	record := g.Record{
		"id":           bid,
		"ty":           TyBankcard,
		"value":        value,
		"remark":       "delete",
		"accounts":     username,
		"created_at":   time.Now().Unix(),
		"created_uid":  "0",
		"created_name": "",
	}
	query, _, _ = dialect.Insert("tbl_blacklist").Rows(&record).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	// 会员删除银行卡，加入黑名单
	idx := bankcardHash % 10
	key := fmt.Sprintf("bl:bc%d", idx)
	// 加入values set
	_, err = meta.MerchantRedis.SAdd(ctx, key, value).Result()
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	return nil
}

func bankcardDeleteCheck(uid, bid string) error {

	ex := g.Ex{
		"uid":   uid,
		"bid":   bid,
		"state": g.Op{"notIn": []int64{WithdrawReviewReject, WithdrawSuccess, WithdrawFailed}},
	}
	var count int
	query, _, _ := dialect.From("tbl_withdraw").Select(g.COUNT(1)).Where(ex).ToSQL()
	err := meta.MerchantDB.Get(&count, query)
	if err != nil {
		return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}

	if count != 0 {
		return errors.New(helper.WithDrawProcessing)
	}

	return nil
}
