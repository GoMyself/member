package model

import (
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"

	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"

	"member2/contrib/helper"
	"member2/contrib/validator"
)

func BankcardInsert(fctx *fasthttp.RequestCtx, phone, realName, bankcardNo string, data BankCard) error {

	encRes := [][]string{}

	mb, err := MemberCache(fctx, "")
	if err != nil {
		return err
	}

	// 判断卡号是否存在
	//bankcardHash := MurmurHash(bankcardNo, 0)

	//判断卡号是否存在
	err = BankCardExistRedis(bankcardNo)
	if err != nil {
		return err
	}

	// 判断会员银行卡数目
	if mb.BankcardTotal >= 3 {
		return errors.New(helper.MaxThreeBankCard)
	}

	//判断卡号是否存在
	err = BankCardExistRedis(bankcardNo)
	if err != nil {
		return err
	}

	recs, err := grpc_t.Decrypt(mb.UID, true, []string{"phone"})
	if recs["phone"] != phone {
		return errors.New(helper.PhoneVerificationErr)
	}

	member_ex := g.Ex{
		"uid": mb.UID,
	}
	member_record := g.Record{
		"bankcard_total": g.L("bankcard_total+1"),
	}
	// 会员未绑定真实姓名，更新第一次绑定银行卡的真实姓名到会员信息
	if mb.RealnameHash == "0" {
		// 第一次新增银行卡判断真实姓名是否为越南语
		if meta.Lang == "vn" && !validator.CheckStringVName(realName) {
			return errors.New(helper.RealNameFMTErr)
		}

		encRes = append(encRes, []string{"realname", realName})
		// 会员信息更新真实姓名和真实姓名hash
		member_record["realname_hash"] = fmt.Sprintf("%d", MurmurHash(realName, 0))
	}

	bankcard_record := g.Record{
		"id":               data.ID,
		"uid":              mb.UID,
		"prefix":           meta.Prefix,
		"username":         data.Username,
		"bank_address":     data.BankAddress,
		"bank_id":          data.BankID,
		"bank_branch_name": data.BankAddress,
		"bank_card_hash":   fmt.Sprintf("%d", MurmurHash(bankcardNo, 0)),
		"created_at":       fmt.Sprintf("%d", data.CreatedAt),
	}

	encRes = append(encRes, []string{"bankcard" + data.ID, bankcardNo})

	// 会员银行卡插入加锁
	lkey := fmt.Sprintf("bc:%s", data.Username)
	err = Lock(lkey)
	if err != nil {
		return err
	}

	defer Unlock(lkey)

	//开启事务
	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	// 更新会员银行卡信息
	queryInsert, _, _ := dialect.Insert("tbl_member_bankcard").Rows(bankcard_record).ToSQL()
	_, err = tx.Exec(queryInsert)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	// 更新会员信息
	queryUpdate, _, _ := dialect.Update("tbl_members").Set(member_record).Where(member_ex).ToSQL()
	_, err = tx.Exec(queryUpdate)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	err = grpc_t.Encrypt(mb.UID, encRes)
	if err != nil {
		fmt.Println("grpc_t.Encrypt = ", err)
		return errors.New(helper.UpdateRPCErr)
	}

	return nil
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
	if err != nil && err != redis.Nil {
		return data, errors.New(helper.RedisErr)
	}

	if err == redis.Nil {
		return data, nil
	}

	res := map[string]BankCard{}
	err = helper.JsonUnmarshal([]byte(bcs), &res)
	if err != nil {
		return data, errors.New(helper.FormatErr)
	}

	for _, val := range res {
		cardList = append(cardList, val)
	}

	length := len(cardList)
	if length == 0 {
		return data, nil
	}

	encField := []string{"realname"}
	for _, v := range cardList {
		encField = append(encField, "bankcard"+v.ID)
	}

	encRes, err := grpc_t.Decrypt(mb.UID, true, encField)
	if err != nil {
		return data, errors.New(helper.GetRPCErr)
	}

	for _, v := range cardList {

		key := "bankcard" + v.ID
		val := BankcardData{BankCard: v, RealName: encRes["realname"], Bankcard: encRes[key]}
		data = append(data, val)
	}

	return data, nil
}

func BankCardExistRedis(bankcardNo string) error {

	pipe := meta.MerchantRedis.Pipeline()
	ex1_temp := pipe.Do(ctx, "CF.EXISTS", "bankcard_exist", bankcardNo)
	ex2_temp := pipe.Do(ctx, "CF.EXISTS", "bankcard_blacklist", bankcardNo)
	_, err := pipe.Exec(ctx)
	pipe.Close()
	if err != nil {
		return errors.New(helper.RedisErr)
	}

	if val, ok := ex1_temp.Val().(string); ok && val == "1" {
		return errors.New(helper.BankCardExistErr)
	}

	if val, ok := ex2_temp.Val().(string); ok && val == "1" {
		return errors.New(helper.BankcardBan)
	}

	return nil
}
