package model

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"member/contrib/helper"
	"member/contrib/validator"
	"strconv"
	"time"

	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fasthttp"
)

// 更新用户密码
func MemberPasswordUpdate(ty int, sid, code, old, password, ts, phone string, fctx *fasthttp.RequestCtx) error {

	mb, err := MemberCache(fctx, "")
	if err != nil {
		return errors.New(helper.AccessTokenExpires)
	}

	//phoneHash := fmt.Sprintf("%d", MurmurHash(phone, 0))
	//if phoneHash != mb.PhoneHash {
	//	return errors.New(helper.UsernamePhoneMismatch)
	//}
	recs, err := grpc_t.Decrypt(mb.UID, false, []string{"phone"})
	if err != nil {
		return errors.New(helper.GetRPCErr)
	}

	phone = recs["phone"]
	ip := helper.FromRequest(fctx)
	err = CheckSmsCaptcha(ip, sid, phone, code)
	if err != nil {
		return err
	}

	// 邮箱 有绑定
	if ty == 1 && mb.PhoneHash != "0" {
		if !helper.CtypeDigit(sid) {
			return errors.New(helper.ParamErr)
		}

		if !helper.CtypeDigit(ts) {
			return errors.New(helper.ParamErr)
		}

		if !helper.CtypeDigit(code) {
			return errors.New(helper.ParamErr)
		}
		//
		//recs, err := grpc_t.Decrypt(mb.UID, false, []string{"phone"})
		//if err != nil {
		//	return errors.New(helper.GetRPCErr)
		//}
		//address := recs["phone"]

	}

	pwd := fmt.Sprintf("%d", MurmurHash(password, mb.CreatedAt))
	// 登录密码修改
	if ty == 1 {
		oldPassword := fmt.Sprintf("%d", MurmurHash(old, mb.CreatedAt))
		// 原始密码是否一致
		if oldPassword != mb.Password {
			return errors.New(helper.OldPasswordErr)
		}
	} else {
		if mb.WithdrawPwd != 0 {
			return errors.New(helper.WithdrawPwdExist)
		} else if pwd == mb.Password {
			return errors.New(helper.WPwdCanNotSameWithPwd)
		}
	}

	ex := g.Ex{
		"uid": mb.UID,
	}
	record := g.Record{}
	if ty == 1 {
		record["password"] = pwd
	} else {
		record["withdraw_pwd"] = pwd
	}
	// 更新会员信息
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}
	fmt.Println("==== UpdatePWD TD Update ====")

	its, ie := strconv.ParseInt(ts, 10, 64)
	if ie != nil {
		fmt.Println("parse int err:", ie)
	}

	tdInsert("sms_log", g.Record{
		"ts":         its,
		"state":      "2",
		"updated_at": time.Now().Unix(),
	})
	fmt.Println("==== UpdatePWD TD Update End ====")

	MemberUpdateCache(mb.UID, "")

	return nil
}

// 更新用户手机号
func MemberUpdatePhone(phone string, fctx *fasthttp.RequestCtx) error {

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

	mb, err := MemberCache(fctx, "")
	if err != nil {
		return err
	}

	//会员绑定手机号后，不允许更新手机号
	if mb.PhoneHash != "0" {
		return errors.New(helper.PhoneBindAlreadyErr)
	}

	encRes := [][]string{}
	encRes = append(encRes, []string{"phone", phone})
	err = grpc_t.Encrypt(mb.UID, encRes)
	if err != nil {
		return errors.New(helper.UpdateRPCErr)
	}

	record := g.Record{
		"phone_hash": phoneHash,
	}
	ex = g.Ex{
		"uid": mb.UID,
	}
	// 更新会员信息
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	MemberUpdateCache(mb.UID, "")
	return nil
}

// 更新用户zalo号
func MemberUpdateZalo(zalo string, fctx *fasthttp.RequestCtx) error {

	zaloHash := fmt.Sprintf("%d", MurmurHash(zalo, 0))
	ex := g.Ex{
		"zalo_hash": zaloHash,
		"prefix":    meta.Prefix,
	}
	if MemberBindCheck(ex) {
		return errors.New(helper.ZaloExist)
	}

	mb, err := MemberCache(fctx, "")
	if err != nil {
		return err
	}

	//会员绑定zalo号后，不允许更新zalo号
	if mb.ZaloHash != "0" {
		return errors.New(helper.ZaloBindAlreadyErr)
	}

	encRes := [][]string{}
	encRes = append(encRes, []string{"zalo", zalo})
	err = grpc_t.Encrypt(mb.UID, encRes)
	if err != nil {
		return errors.New(helper.UpdateRPCErr)
	}

	record := g.Record{
		"zalo_hash": zaloHash,
	}
	ex = g.Ex{
		"uid": mb.UID,
	}
	// 更新会员信息
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	MemberUpdateCache(mb.UID, "")
	return nil
}

// 更新用户信息
func MemberUpdateAvatar(avatar string, fctx *fasthttp.RequestCtx) error {

	mb, err := MemberCache(fctx, "")
	if err != nil {
		return err
	}

	record := g.Record{
		"avatar": avatar,
	}
	ex := g.Ex{
		"uid": mb.UID,
	}
	// 更新会员信息
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	MemberUpdateCache(mb.UID, "")
	return nil
}

func CheckEmailCaptcha(ip, sid, email, code string) error {

	key := fmt.Sprintf("%s:mail:%s%s", meta.Prefix, email, sid)
	cmd := meta.MerchantRedis.Get(ctx, key)
	//fmt.Println(cmd.String())
	val, err := cmd.Result()
	if err != nil && err != redis.Nil {
		_ = pushLog(err, helper.RedisErr)
		return errors.New(helper.EmailVerificationErr)
	}

	if code == val {
		return nil
	}

	return errors.New(helper.EmailVerificationErr)
}

// 更新用户信息
func MemberUpdateEmail(email, ts string, fctx *fasthttp.RequestCtx) error {

	emailHash := fmt.Sprintf("%d", MurmurHash(email, 0))
	ex := g.Ex{
		"email_hash": emailHash,
		"prefix":     meta.Prefix,
	}
	if MemberBindCheck(ex) {
		return errors.New(helper.EmailExist)
	}

	mb, err := MemberCache(fctx, "")
	if err != nil {
		return err
	}

	encRes := [][]string{}
	encRes = append(encRes, []string{"email", email})
	err = grpc_t.Encrypt(mb.UID, encRes)
	if err != nil {
		return errors.New(helper.UpdateRPCErr)
	}

	record := g.Record{
		"email_hash": emailHash,
	}
	ex = g.Ex{
		"uid": mb.UID,
	}
	// 更新会员信息
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	MemberUpdateCache(mb.UID, "")
	fmt.Println("==== Bind email TD Update ====")
	its, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		fmt.Println("parse int err:", err)
	}
	tdInsert("mail_log", g.Record{
		"ts":         its,
		"state":      "2",
		"updated_at": time.Now().Unix(),
	})
	fmt.Println("==== Bind email TD Update End ====")
	return nil
}

// 用户信息更新
func MemberUpdateName(fctx *fasthttp.RequestCtx, birth, realName, address string) error {

	mb, err := MemberCache(fctx, "")
	if err != nil {
		return err
	}

	record := g.Record{}
	if birth != "" {
		t, err := time.Parse("2006-01-02", birth)
		if err != nil {
			return errors.New(helper.TimeTypeErr)
		}

		record["birth"] = fmt.Sprintf("%d", t.Unix())
		record["birth_hash"] = fmt.Sprintf("%d", MurmurHash(birth, 0))
	}

	//会员填写真实姓名后，不允许更新真实姓名
	if realName != "" { // 传入用户真实姓名  需要修改

		if mb.RealnameHash != "0" {
			return errors.New(helper.RealNameAlreadyBind)
		}

		if meta.Lang == "vn" && !validator.CheckStringVName(realName) {
			return errors.New(helper.RealNameFMTErr)
		}

		encRes := [][]string{}
		encRes = append(encRes, []string{"realname", realName})
		err = grpc_t.Encrypt(mb.UID, encRes)
		if err != nil {
			return errors.New(helper.UpdateRPCErr)
		}

		record["realname_hash"] = fmt.Sprintf("%d", MurmurHash(realName, 0))
	}

	if address != "" {
		record["address"] = address
	}

	if len(record) == 0 {
		return errors.New(helper.NoDataUpdate)
	}
	ex := g.Ex{
		"uid": mb.UID,
	}
	// 更新会员信息
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(fmt.Errorf("birth : %s ,query : %s ,error : %s", birth, query, err.Error()), helper.DBErr)
	}

	MemberUpdateCache(mb.UID, "")
	return nil
}
