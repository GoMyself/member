package model

import (
	"fmt"
	"member2/contrib/helper"

	g "github.com/doug-martin/goqu/v9"
	"github.com/shopspring/decimal"
)

type MemberRebateResult_t struct {
	ZR               decimal.Decimal
	QP               decimal.Decimal
	TY               decimal.Decimal
	DZ               decimal.Decimal
	DJ               decimal.Decimal
	CP               decimal.Decimal
	FC               decimal.Decimal
	BY               decimal.Decimal
	CGOfficialRebate decimal.Decimal
	CGHighRebate     decimal.Decimal
}

func MemberRebateUpdateALL() error {

	var data []MemberRebate

	t := dialect.From("tbl_member_rebate_info")
	query, _, _ := t.Select(colsMemberRebate...).ToSQL()
	err := meta.MerchantDB.Select(&data, query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	pipe := meta.MerchantRedis.Pipeline()

	for _, mr := range data {

		key := fmt.Sprintf("%s:m:rebate:%s", meta.Prefix, mr.UID)
		vals := []interface{}{"zr", mr.ZR, "qp", mr.QP, "ty", mr.TY, "dj", mr.DJ, "dz", mr.DZ, "cp", mr.CP, "fc", mr.FC, "by", mr.BY, "cg_high_rebate", mr.CGHighRebate, "cg_official_rebate", mr.CGOfficialRebate}

		pipe.Unlink(ctx, key)
		pipe.HMSet(ctx, key, vals...)
		pipe.Persist(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	pipe.Close()

	return nil
}

func MemberRebateUpdateCache(mr MemberRebate) error {

	key := fmt.Sprintf("%s:m:rebate:%s", meta.Prefix, mr.UID)
	vals := []interface{}{"zr", mr.ZR, "qp", mr.QP, "ty", mr.TY, "dj", mr.DJ, "dz", mr.DZ, "cp", mr.CP, "fc", mr.FC, "by", mr.BY, "cg_high_rebate", mr.CGHighRebate, "cg_official_rebate", mr.CGOfficialRebate}

	pipe := meta.MerchantRedis.Pipeline()
	pipe.Unlink(ctx, key)
	pipe.HMSet(ctx, key, vals...)
	pipe.Persist(ctx, key)
	_, err := pipe.Exec(ctx)
	pipe.Close()

	return err
}

func RebateScale(uid string) (MemberRebate, error) {

	data := MemberRebate{}

	t := dialect.From("tbl_member_rebate_info")
	query, _, _ := t.Select(colsMemberRebate...).Where(g.Ex{"uid": uid}).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&data, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func MemberRebateFindOne(uid string) (MemberRebateResult_t, error) {

	data := MemberRebate{}
	res := MemberRebateResult_t{}

	t := dialect.From("tbl_member_rebate_info")
	query, _, _ := t.Select(colsMemberRebate...).Where(g.Ex{"uid": uid}).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&data, query)
	if err != nil {
		return res, pushLog(err, helper.DBErr)
	}

	res.ZR, _ = decimal.NewFromString(data.ZR)
	res.QP, _ = decimal.NewFromString(data.QP)
	res.TY, _ = decimal.NewFromString(data.TY)
	res.DJ, _ = decimal.NewFromString(data.DJ)
	res.DZ, _ = decimal.NewFromString(data.DZ)
	res.CP, _ = decimal.NewFromString(data.CP)
	res.FC, _ = decimal.NewFromString(data.FC)
	res.BY, _ = decimal.NewFromString(data.BY)
	res.CGHighRebate, _ = decimal.NewFromString(data.CGHighRebate)
	res.CGOfficialRebate, _ = decimal.NewFromString(data.CGOfficialRebate)

	res.ZR = res.ZR.Truncate(1)
	res.QP = res.QP.Truncate(1)
	res.TY = res.TY.Truncate(1)
	res.DJ = res.DJ.Truncate(1)
	res.DZ = res.DZ.Truncate(1)
	res.CP = res.CP.Truncate(1)
	res.FC = res.FC.Truncate(1)
	res.BY = res.BY.Truncate(1)
	res.CGHighRebate = res.CGHighRebate.Truncate(2)
	res.CGOfficialRebate = res.CGOfficialRebate.Truncate(2)

	return res, nil
}

func MemberRebateSelect(ids []string) (map[string]MemberRebate, error) {

	var own []MemberRebate
	query, _, _ := dialect.From("tbl_member_rebate_info").Select(colsMemberRebate...).Where(g.Ex{"uid": ids}).ToSQL()
	err := meta.MerchantDB.Select(&own, query)
	if err != nil {
		return nil, pushLog(err, helper.DBErr)
	}

	data := make(map[string]MemberRebate)
	for _, v := range own {
		data[v.UID] = v
	}
	return data, nil
}
