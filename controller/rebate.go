package controller

import (
	"fmt"
	"member2/contrib/helper"
	"member2/contrib/validator"
	"member2/model"

	"github.com/shopspring/decimal"
	"github.com/valyala/fasthttp"
)

type RebateController struct{}

func (that *RebateController) Scale(ctx *fasthttp.RequestCtx) {

	mb, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	vs, err := model.RebateScale(mb.UID)
	if err != nil {
		helper.Print(ctx, false, err.Error())
	}

	helper.Print(ctx, true, vs)
}

func (that *RebateController) Detail(ctx *fasthttp.RequestCtx) {

	uid := string(ctx.QueryArgs().Peek("uid"))

	if !helper.CtypeDigit(uid) {
		helper.Print(ctx, false, helper.PLatNameErr)
		return
	}
	res, err := model.MemberRebateGetCache(uid)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}
	helper.Print(ctx, true, res)
}

// UpdateRebate 修改密码以及返水比例
func (that *RebateController) Update(ctx *fasthttp.RequestCtx) {

	subName := string(ctx.PostArgs().Peek("username"))
	password := string(ctx.PostArgs().Peek("password"))
	ty_temp := string(ctx.PostArgs().Peek("ty"))
	zr_temp := string(ctx.PostArgs().Peek("zr"))
	qp_temp := string(ctx.PostArgs().Peek("qp"))
	dj_temp := string(ctx.PostArgs().Peek("dj"))
	dz_temp := string(ctx.PostArgs().Peek("dz"))
	cp_temp := string(ctx.PostArgs().Peek("cp"))
	fc_temp := string(ctx.PostArgs().Peek("fc"))
	by_temp := string(ctx.PostArgs().Peek("by"))
	cg_high_rebate_temp := string(ctx.PostArgs().Peek("cg_high_rebate"))
	cg_official_rebate_temp := string(ctx.PostArgs().Peek("cg_official_rebate"))

	fmt.Println("Update = ", string(ctx.PostBody()))

	if !helper.CtypeAlnum(subName) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	mb, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	//fmt.Println("mb = ", mb)

	child, err := model.MemberCache(nil, subName)
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	if child.ParentName != mb.Username {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	parent, err := model.MemberRebateFindOne(mb.UID)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	ty, err := decimal.NewFromString(ty_temp) //下级会员体育返水比例
	if err != nil || ty.IsNegative() || ty.GreaterThan(parent.TY) {
		fmt.Println("ty_temp = ", ty_temp)
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	zr, err := decimal.NewFromString(zr_temp) //下级会员真人返水比例
	if err != nil || zr.IsNegative() || zr.GreaterThan(parent.ZR) {

		fmt.Println("zr_temp = ", zr_temp)
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	qp, err := decimal.NewFromString(qp_temp) //下级会员棋牌返水比例
	if err != nil || qp.IsNegative() || qp.GreaterThan(parent.QP) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	dj, err := decimal.NewFromString(dj_temp) //下级会员电竞返水比例
	if err != nil || dj.IsNegative() || dj.GreaterThan(parent.DJ) {

		fmt.Println("dj_temp = ", dj_temp)
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	dz, err := decimal.NewFromString(dz_temp) //下级会员电子返水比例
	if err != nil || dz.IsNegative() || dz.GreaterThan(parent.DZ) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	cp, err := decimal.NewFromString(cp_temp) //下级会员彩票返水比例
	if err != nil || cp.IsNegative() || cp.GreaterThan(parent.CP) {

		fmt.Println("cp_temp = ", cp_temp)
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	fc, err := decimal.NewFromString(fc_temp) //下级会员斗鸡返水比例
	if err != nil || fc.IsNegative() || fc.GreaterThan(parent.FC) {

		fmt.Println("fc_temp = ", fc_temp)
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	by, err := decimal.NewFromString(by_temp) //下级会员捕鱼返水比例
	if err != nil || by.IsNegative() || by.GreaterThan(parent.BY) {

		fmt.Println("by_temp = ", by_temp)
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	cgHighRebate, err := decimal.NewFromString(cg_high_rebate_temp)
	if err != nil || fc.IsNegative() || cgHighRebate.GreaterThan(parent.CGHighRebate) {

		fmt.Println("cgHighRebateTemp = ", cg_high_rebate_temp)
		helper.Print(ctx, false, helper.RebateOutOfRange)
	}
	cgOfficialRebate, err := decimal.NewFromString(cg_official_rebate_temp)
	if err != nil || fc.IsNegative() || cgOfficialRebate.GreaterThan(parent.CGOfficialRebate) {

		fmt.Println("cgOfficialRebateTemp = ", cg_official_rebate_temp)
		helper.Print(ctx, false, helper.RebateOutOfRange)
	}

	if !validator.CheckUName(subName, 5, 14) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	if password != "" {
		if !validator.CheckUPassword(password, 8, 20) {
			helper.Print(ctx, false, helper.PasswordFMTErr)
			return
		}
	}

	mr := model.MemberRebateResult_t{
		TY:               ty,
		ZR:               zr,
		QP:               qp,
		DJ:               dj,
		DZ:               dz,
		CP:               cp,
		FC:               fc,
		BY:               by,
		CGHighRebate:     cgHighRebate,
		CGOfficialRebate: cgOfficialRebate,
	}

	if ok := model.MemberRebateCmp(child.UID, mr); !ok {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	// 添加下级代理
	err = model.MemberUpdateInfo(child, password, mr)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
