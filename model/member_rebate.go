package model

import (
	"database/sql"
	"errors"
	"fmt"
	"member/contrib/helper"

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

func MemberMaxRebateFindOne(uid string) (MemberRebateResult_t, error) {

	data := MemberMaxRebate{}
	res := MemberRebateResult_t{}

	t := dialect.From("tbl_member_rebate_info")
	query, _, _ := t.Select(
		g.MAX("zr").As("zr"),
		g.MAX("qp").As("qp"),
		g.MAX("dz").As("dz"),
		g.MAX("dj").As("dj"),
		g.MAX("ty").As("ty"),
		g.MAX("cp").As("cp"),
		g.MAX("fc").As("fc"),
		g.MAX("by").As("by"),
		g.MAX("cg_high_rebate").As("cg_high_rebate"),
		g.MAX("cg_official_rebate").As("cg_official_rebate"),
	).Where(g.Ex{"parent_uid": uid, "prefix": meta.Prefix}).ToSQL()
	err := meta.MerchantDB.Get(&data, query)
	if err == sql.ErrNoRows {

		res.ZR = decimal.NewFromInt(0).Truncate(1)
		res.QP = decimal.NewFromInt(0).Truncate(1)
		res.TY = decimal.NewFromInt(0).Truncate(1)
		res.DJ = decimal.NewFromInt(0).Truncate(1)
		res.DZ = decimal.NewFromInt(0).Truncate(1)
		res.CP = decimal.NewFromInt(0).Truncate(1)
		res.FC = decimal.NewFromInt(0).Truncate(1)
		res.BY = decimal.NewFromInt(0).Truncate(1)
		res.CGHighRebate = decimal.NewFromFloat(9.00).Truncate(2)
		res.CGOfficialRebate = decimal.NewFromFloat(9.00).Truncate(2)

		return res, nil
	}
	if err != nil {
		return res, pushLog(err, helper.DBErr)
	}

	res.ZR = decimal.NewFromFloat(data.ZR.Float64).Truncate(1)
	res.QP = decimal.NewFromFloat(data.QP.Float64).Truncate(1)
	res.TY = decimal.NewFromFloat(data.TY.Float64).Truncate(1)
	res.DJ = decimal.NewFromFloat(data.DJ.Float64).Truncate(1)
	res.DZ = decimal.NewFromFloat(data.DZ.Float64).Truncate(1)
	res.CP = decimal.NewFromFloat(data.CP.Float64).Truncate(1)
	res.FC = decimal.NewFromFloat(data.FC.Float64).Truncate(1)
	res.BY = decimal.NewFromFloat(data.BY.Float64).Truncate(1)
	res.CGHighRebate = decimal.NewFromFloat(data.CgHighRebate.Float64).Truncate(2)
	res.CGOfficialRebate = decimal.NewFromFloat(data.CgOfficialRebate.Float64).Truncate(2)

	return res, nil
}

func MemberRebateUpdateCache1(uid string, mr MemberRebateResult_t) error {

	key := fmt.Sprintf("%s:m:rebate:%s", meta.Prefix, uid)
	vals := []interface{}{"zr", mr.ZR.StringFixed(1), "qp", mr.QP.StringFixed(1), "ty", mr.TY.StringFixed(1), "dj", mr.DJ.StringFixed(1), "dz", mr.DZ.StringFixed(1), "cp", mr.CP.StringFixed(1), "fc", mr.FC.StringFixed(1), "by", mr.BY.StringFixed(1), "cg_high_rebate", mr.CGHighRebate.StringFixed(2), "cg_official_rebate", mr.CGOfficialRebate.StringFixed(2)}

	pipe := meta.MerchantRedis.Pipeline()
	pipe.Del(ctx, key)
	pipe.HMSet(ctx, key, vals...)
	pipe.Persist(ctx, key)
	_, err := pipe.Exec(ctx)
	pipe.Close()

	return err
}

func MemberRebateUpdateCache2(uid string, mr MemberRebate) error {

	key := fmt.Sprintf("%s:m:rebate:%s", meta.Prefix, uid)
	vals := []interface{}{"zr", mr.ZR, "qp", mr.QP, "ty", mr.TY, "dj", mr.DJ, "dz", mr.DZ, "cp", mr.CP, "fc", mr.FC, "by", mr.BY, "cg_high_rebate", mr.CGHighRebate, "cg_official_rebate", mr.CGOfficialRebate}

	pipe := meta.MerchantRedis.Pipeline()
	pipe.Del(ctx, key)
	pipe.HMSet(ctx, key, vals...)
	pipe.Persist(ctx, key)
	_, err := pipe.Exec(ctx)
	pipe.Close()

	return err
}

func MemberRebateCmp(uid string, own MemberRebateResult_t) bool {

	lower, err := MemberMaxRebateFindOne(uid)
	if err != nil {
		return false
	}

	if own.QP.Cmp(lower.QP) == -1 {
		return false
	}
	if own.ZR.Cmp(lower.ZR) == -1 {
		return false
	}
	if own.TY.Cmp(lower.TY) == -1 {
		return false
	}
	if own.DJ.Cmp(lower.DJ) == -1 {
		return false
	}
	if own.DZ.Cmp(lower.DZ) == -1 {
		return false
	}
	if own.CP.Cmp(lower.CP) == -1 {
		return false
	}
	if own.FC.Cmp(lower.FC) == -1 {
		return false
	}
	if own.BY.Cmp(lower.BY) == -1 {
		return false
	}
	if own.CGHighRebate.Cmp(lower.CGHighRebate) == -1 {
		return false
	}
	if own.CGOfficialRebate.Cmp(lower.CGOfficialRebate) == -1 {
		return false
	}
	return true

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

/*
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
*/

func MemberRebateGetCache(uid string) (MemberRebate, error) {

	m := MemberRebate{}
	key := fmt.Sprintf("%s:m:rebate:%s", meta.Prefix, uid)

	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()

	exist := pipe.Exists(ctx, key)
	rs := pipe.HMGet(ctx, key, "zr", "dj", "ty", "dz", "cp", "fc", "by", "cg_high_rebate", "cg_official_rebate", "qp")

	_, err := pipe.Exec(ctx)
	if err != nil {
		return m, pushLog(err, helper.RedisErr)
	}

	if exist.Val() == 0 {
		return m, errors.New(helper.RecordNotExistErr)
	}

	if err = rs.Scan(&m); err != nil {
		return m, pushLog(rs.Err(), helper.RedisErr)
	}

	return m, nil
}

func MemberRebateFindOne(uid string) (MemberRebateResult_t, error) {

	data := MemberRebate{}
	res := MemberRebateResult_t{}

	t := dialect.From("tbl_member_rebate_info")
	query, _, _ := t.Select(colsMemberRebate...).Where(g.Ex{"uid": uid}).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&data, query)
	if err != nil {
		fmt.Println("MemberRebateFindOne query = ", query)
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
	fmt.Printf("反水查询 tbl_member_rebate_info sql:%+v\n", query)
	if err != nil {
		return nil, pushLog(err, helper.DBErr)
	}

	data := make(map[string]MemberRebate)
	for _, v := range own {
		data[v.UID] = v
	}
	return data, nil
}
