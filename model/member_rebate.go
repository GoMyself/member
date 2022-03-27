package model

import (
	"member2/contrib/helper"

	g "github.com/doug-martin/goqu/v9"
	"github.com/shopspring/decimal"
)

type MemberRebateResult_t struct {
	ZR decimal.Decimal
	QP decimal.Decimal
	TY decimal.Decimal
	DZ decimal.Decimal
	DJ decimal.Decimal
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

	res.ZR = res.ZR.Truncate(1)
	res.QP = res.QP.Truncate(1)
	res.TY = res.TY.Truncate(1)
	res.DJ = res.DJ.Truncate(1)
	res.DZ = res.DZ.Truncate(1)

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
