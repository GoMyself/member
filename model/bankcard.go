package model

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"

	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"

	"member/contrib/helper"
)

func BankcardUpdateCache(username string) {

	var data []BankCard

	ex := g.Ex{
		"prefix":   meta.Prefix,
		"username": username,
		//"state":    "1",
	}

	t := dialect.From("tbl_member_bankcard")
	query, _, _ := t.Select(colsBankcard...).Where(ex).Order(g.C("created_at").Desc()).ToSQL()

	err := meta.MerchantDB.Select(&data, query)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("BankcardUpdateCache err = ", err)
		return
	}

	key := fmt.Sprintf("%s:merchant:cbc:%s", meta.Prefix, username)

	pipe := meta.MerchantRedis.Pipeline()
	pipe.Del(ctx, key)
	if len(data) > 0 {

		value, err := helper.JsonMarshal(data)
		if err == nil {
			pipe.Set(ctx, key, string(value), 0).Err()
			//fmt.Println("JSON.SET err = ", err)
		}
	}

	pipe.Exec(ctx)
	pipe.Close()
}

func BankcardInsert(fctx *fasthttp.RequestCtx, phone, realName, bankcardNo string, data BankCard) error {

	encRes := [][]string{}

	mb, err := MemberCache(fctx, "")
	if err != nil {
		return err
	}

	// 判断会员银行卡数目
	if mb.BankcardTotal >= 5 {
		return errors.New(helper.MaxThreeBankCard)
	}

	//判断卡号是否存在
	err = BankCardExistRedis(bankcardNo)
	if err != nil {
		return err
	}

	recs, err := grpc_t.Decrypt(mb.UID, false, []string{"phone", "realname"})
	if recs["phone"] != phone {
		fmt.Println("phone = ", phone)
		fmt.Println("recs phone = ", recs["phone"])
		//return errors.New(helper.PhoneVerificationErr)
	}

	/*
		header := map[string]string{}
		postbody := fmt.Sprintf("{\"bankCard\":\"%s\", \"name\":\"%s\", \"bankCode\":\"%s\"}", bankcardNo, recs["realname"], "VPBank")

		statusBody, statusCode, err := helper.HttpDoTimeout([]byte(postbody), "POST", "http://34.142.195.249:8090/bank/check/create", header, 6*time.Second)
		fmt.Println("statusBody = ", string(statusBody))
		fmt.Println("statusCode = ", statusCode)
		fmt.Println("statusCode err = ", err)
	*/
	memberEx := g.Ex{
		"uid": mb.UID,
	}
	memberRecord := g.Record{
		"bankcard_total": g.L("bankcard_total+1"),
	}

	// 会员未绑定真实姓名，更新第一次绑定银行卡的真实姓名到会员信息
	if mb.RealnameHash == "0" {
		// 第一次新增银行卡判断真实姓名是否为越南语
		/*
			if meta.Lang == "vn" && !validator.CheckStringVName(realName) {
				return errors.New(helper.RealNameFMTErr)
			}
		*/
		encRes = append(encRes, []string{"realname", realName})
		// 会员信息更新真实姓名和真实姓名hash
		memberRecord["realname_hash"] = fmt.Sprintf("%d", MurmurHash(realName, 0))
	} else {
		realName = recs["realname"]
	}

	statusCoce := BankcardCheck(fctx, bankcardNo, data.BankID, realName)
	if statusCoce != helper.Success {
		return errors.New(statusCoce)
	}
	bankcardRecord := g.Record{
		"id":               data.ID,
		"uid":              mb.UID,
		"prefix":           meta.Prefix,
		"username":         data.Username,
		"bank_address":     data.BankAddress,
		"bank_id":          data.BankID,
		"bank_branch_name": data.BankAddress,
		"bank_card_hash":   fmt.Sprintf("%d", MurmurHash(bankcardNo, 0)),
		"created_at":       data.CreatedAt,
	}

	encRes = append(encRes, []string{"bankcard" + data.ID, bankcardNo})

	// 会员银行卡插入加锁
	lKey := fmt.Sprintf("bc:%s", data.Username)
	err = Lock(lKey)
	if err != nil {
		return err
	}

	defer Unlock(lKey)

	//开启事务
	tx, err := meta.MerchantDB.Begin()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	// 更新会员银行卡信息
	queryInsert, _, _ := dialect.Insert("tbl_member_bankcard").Rows(bankcardRecord).ToSQL()
	_, err = tx.Exec(queryInsert)
	if err != nil {
		_ = tx.Rollback()
		fmt.Println("queryInsert = ", queryInsert)
		return pushLog(err, helper.DBErr)
	}

	// 更新会员信息
	queryUpdate, _, _ := dialect.Update("tbl_members").Set(memberRecord).Where(memberEx).ToSQL()
	_, err = tx.Exec(queryUpdate)
	if err != nil {
		_ = tx.Rollback()
		fmt.Println("queryUpdate = ", queryUpdate)
		return pushLog(err, helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println("tx.Commit = ", err.Error())
		return pushLog(err, helper.DBErr)
	}

	err = grpc_t.Encrypt(mb.UID, encRes)
	if err != nil {
		fmt.Println("grpc_t.Encrypt = ", err)
		return errors.New(helper.UpdateRPCErr)
	}

	_ = MemberUpdateCache(mb.UID, "")
	BankcardUpdateCache(mb.Username)

	key := fmt.Sprintf("%s:merchant:bankcard_exist", meta.Prefix)
	_ = meta.MerchantRedis.SAdd(ctx, key, bankcardNo).Err()

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

	if mb.BankcardTotal == 0 {
		return data, nil
	}

	key := fmt.Sprintf("%s:merchant:cbc:%s", meta.Prefix, username)
	cmd := meta.MerchantRedis.Get(ctx, key)
	bcs, err := cmd.Result()
	if err != nil && err != redis.Nil {
		//fmt.Println("BankcardList GET err = ", err.Error())
		return data, pushLog(fmt.Errorf("%s, error : %s", cmd.String(), err.Error()), helper.RedisErr)
	}

	if err == redis.Nil {
		return data, nil
	}

	root, err := fastjson.MustParse(bcs).Array()
	if err != nil {
		return data, nil
	}

	for _, val := range root {
		res := BankCard{}
		res.ID = string(val.GetStringBytes("id"))
		res.UID = string(val.GetStringBytes("uid"))
		res.Username = string(val.GetStringBytes("username"))
		res.BankID = string(val.GetStringBytes("bank_id"))
		res.BankAddress = string(val.GetStringBytes("bank_address"))
		res.BankBranch = string(val.GetStringBytes("bank_branch_name"))
		res.State = val.GetInt("state")

		cardList = append(cardList, res)
	}
	root = nil

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
		_ = pushLog(fmt.Errorf("error : %s, encRes :%v", err, encRes), helper.GetRPCErr)
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
	defer pipe.Close()

	key := fmt.Sprintf("%s:merchant:bankcard_exist", meta.Prefix)
	ex1Temp := pipe.SIsMember(ctx, key, bankcardNo)
	key = fmt.Sprintf("%s:merchant:bankcard_blacklist", meta.Prefix)
	ex2Temp := pipe.SIsMember(ctx, key, bankcardNo)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return pushLog(err, helper.RedisErr)
	}

	if ex1Temp.Val() {
		return errors.New(helper.BankCardExistErr)
	}

	if ex2Temp.Val() {
		return errors.New(helper.BankcardBan)
	}
	return nil
}
