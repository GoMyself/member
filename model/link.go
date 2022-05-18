package model

import (
	"errors"
	"fmt"
	"member2/contrib/helper"

	"github.com/go-redis/redis/v8"

	g "github.com/doug-martin/goqu/v9"
	"github.com/shopspring/decimal"
	"github.com/valyala/fasthttp"
)

type Link_t struct {
	ID               string `db:"id" json:"id" required:"0"`                                                                                               //
	UID              string `db:"uid" json:"uid" required:"0"`                                                                                             //
	ZR               string `name:"zr" db:"zr" json:"zr" rule:"float" required:"1" min:"3" max:"3" msg:""`                                                 //真人返水
	QP               string `name:"qp" db:"qp" json:"qp" rule:"float" required:"1" min:"3" max:"3" msg:""`                                                 //棋牌返水
	TY               string `name:"ty" db:"ty" json:"ty" rule:"float" required:"1" min:"3" max:"3" msg:""`                                                 //体育返水
	DJ               string `name:"dj" db:"dj" json:"dj" rule:"float" required:"1" min:"3" max:"3" msg:""`                                                 //电竞返水
	DZ               string `name:"dz" db:"dz" json:"dz" rule:"float" required:"1" min:"3" max:"3" msg:""`                                                 //电子返水
	CP               string `name:"cp" db:"cp" json:"cp" rule:"float" required:"1" min:"3" max:"3" msg:""`                                                 //彩票返水
	FC               string `name:"fc" db:"fc" json:"fc" rule:"float" required:"1" min:"3" max:"3" msg:""`                                                 //斗鸡返水
	BY               string `name:"by" db:"by" json:"by" rule:"float" required:"1" min:"3" max:"3" msg:""`                                                 //捕鱼返水
	CGHighRebate     string `name:"cg_high_rebate" db:"cg_high_rebate" json:"cg_high_rebate" rule:"float" required:"1" min:"3" max:"5" msg:""`             //斗鸡返水
	CGOfficialRebate string `name:"cg_official_rebate" db:"cg_official_rebate" json:"cg_official_rebate" rule:"float" required:"1" min:"3" max:"5" msg:""` //斗鸡返水
	CreatedAt        string `db:"created_at" json:"created_at" rule:"none" required:"0"`                                                                   //
}

func LinkInsert(ctx *fasthttp.RequestCtx, data Link_t) error {

	zr, _ := decimal.NewFromString(data.ZR)
	qp, _ := decimal.NewFromString(data.QP)
	ty, _ := decimal.NewFromString(data.TY)
	dj, _ := decimal.NewFromString(data.DJ)
	dz, _ := decimal.NewFromString(data.DZ)
	cp, _ := decimal.NewFromString(data.CP)
	fc, _ := decimal.NewFromString(data.FC)
	by, _ := decimal.NewFromString(data.BY)
	cgHighRebate, _ := decimal.NewFromString(data.CGHighRebate)
	cgOfficialRebate, _ := decimal.NewFromString(data.CGOfficialRebate)

	zr = zr.Truncate(1)
	qp = qp.Truncate(1)
	ty = ty.Truncate(1)
	dj = dj.Truncate(1)
	dz = dz.Truncate(1)
	cp = cp.Truncate(1)
	fc = fc.Truncate(1)
	by = by.Truncate(1)
	cgHighRebate = cgHighRebate.Truncate(1)
	cgOfficialRebate = cgOfficialRebate.Truncate(1)

	sess, err := MemberInfo(ctx)
	if err != nil {
		return err
	}

	own, err := MemberRebateFindOne(sess.UID)
	if err != nil {
		return err
	}

	if qp.GreaterThan(own.QP) || qp.IsNegative() {
		return errors.New(helper.RebateOutOfRange)
	}
	if zr.GreaterThan(own.ZR) || zr.IsNegative() {
		return errors.New(helper.RebateOutOfRange)
	}
	if ty.GreaterThan(own.TY) || ty.IsNegative() {
		return errors.New(helper.RebateOutOfRange)
	}
	if dz.GreaterThan(own.DZ) || dz.IsNegative() {
		return errors.New(helper.RebateOutOfRange)
	}
	if dj.GreaterThan(own.DJ) || dj.IsNegative() {
		return errors.New(helper.RebateOutOfRange)
	}
	if cp.GreaterThan(own.CP) || cp.IsNegative() {
		return errors.New(helper.RebateOutOfRange)
	}
	if fc.GreaterThan(own.FC) || fc.IsNegative() {
		return errors.New(helper.RebateOutOfRange)
	}
	if by.GreaterThan(own.BY) || by.IsNegative() {
		return errors.New(helper.RebateOutOfRange)
	}
	if cgOfficialRebate.GreaterThan(own.CGOfficialRebate) || cgOfficialRebate.IsNegative() {
		return errors.New(helper.RebateOutOfRange)
	}
	if cgHighRebate.GreaterThan(own.CGHighRebate) || cgHighRebate.IsNegative() {
		return errors.New(helper.RebateOutOfRange)
	}

	recs := g.Record{
		"id":                 data.ID,
		"uid":                sess.UID,
		"created_at":         data.CreatedAt,
		"zr":                 zr.StringFixed(1),
		"qp":                 qp.StringFixed(1),
		"ty":                 ty.StringFixed(1),
		"dj":                 dj.StringFixed(1),
		"dz":                 dz.StringFixed(1),
		"cp":                 cp.StringFixed(1),
		"fc":                 fc.StringFixed(1),
		"by":                 by.StringFixed(1),
		"cg_high_rebate":     cgOfficialRebate.StringFixed(1),
		"cg_official_rebate": cgHighRebate.StringFixed(1),
		"prefix":             meta.Prefix,
	}

	query, _, _ := dialect.Insert("tbl_member_link").Rows(recs).ToSQL()
	fmt.Println("query = ", query)
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

func LinkDelete(ctx *fasthttp.RequestCtx, id string) error {

	sess, err := MemberInfo(ctx)
	if err != nil {
		return err
	}

	query, _, _ := dialect.Delete("tbl_member_link").Where(g.Ex{
		"id":     id,
		"uid":    sess.UID,
		"prefix": meta.Prefix,
	}).ToSQL()

	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}
	return nil
}

func LinkList(fCtx *fasthttp.RequestCtx) ([]Link_t, error) {

	var data []Link_t
	sess, err := MemberCache(fCtx, "")
	if err != nil {
		return data, err
	}

	key := "lk:" + sess.UID
	res, err := meta.MerchantRedis.Do(ctx, "JSON.GET", key, ".").Text()
	if err != nil && err != redis.Nil {
		return data, pushLog(err, helper.RedisErr)
	}

	if err == redis.Nil {
		return data, nil
	}

	mp := map[string]Link_t{}
	err = helper.JsonUnmarshal([]byte(res), &mp)
	if err != nil {
		return data, pushLog(err, helper.FormatErr)
	}

	for _, v := range mp {
		data = append(data, v)
	}

	return data, nil
}
