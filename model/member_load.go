package model

import (
	"fmt"
)

func MemberFlushAll() {

	var m []Member

	t := dialect.From("tbl_members")
	query, _, _ := t.Select(colsMember...).ToSQL()
	err := meta.MerchantDB.Select(&m, query)
	if err != nil {
		fmt.Println("MemberFlushAll meta.MerchantDB.Select = ", err.Error())
		return
	}

	pipe := meta.MerchantRedis.TxPipeline()

	for _, dst := range m {

		key := meta.Prefix + ":member:" + dst.Username
		fields := []interface{}{"uid", dst.UID, "username", dst.Username, "password", dst.Password, "birth", dst.Birth, "birth_hash", dst.BirthHash, "realname_hash", dst.RealnameHash, "email_hash", dst.EmailHash, "phone_hash", dst.PhoneHash, "zalo_hash", dst.ZaloHash, "prefix", dst.Prefix, "tester", dst.Tester, "withdraw_pwd", dst.WithdrawPwd, "regip", dst.Regip, "reg_device", dst.RegDevice, "reg_url", dst.RegUrl, "created_at", dst.CreatedAt, "last_login_ip", dst.LastLoginIp, "last_login_at", dst.LastLoginAt, "source_id", dst.SourceId, "first_deposit_at", dst.FirstDepositAt, "first_deposit_amount", dst.FirstDepositAmount, "first_bet_at", dst.FirstBetAt, "first_bet_amount", dst.FirstBetAmount, "", dst.SecondDepositAt, "", dst.SecondDepositAmount, "top_uid", dst.TopUid, "top_name", dst.TopName, "parent_uid", dst.ParentUid, "parent_name", dst.ParentName, "bankcard_total", dst.BankcardTotal, "last_login_device", dst.LastLoginDevice, "last_login_source", dst.LastLoginSource, "remarks", dst.Remarks, "state", dst.State, "level", dst.Level, "balance", dst.Balance, "lock_amount", dst.LockAmount, "commission", dst.Commission, "group_name", dst.GroupName, "agency_type", dst.AgencyType, "address", dst.Address, "avatar", dst.Avatar}

		pipe.Del(ctx, key)
		pipe.HMSet(ctx, key, fields...)
		pipe.Persist(ctx, key)
	}
	pipe.Exec(ctx)
	pipe.Close()
}
