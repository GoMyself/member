package controller

import (
	"fmt"
	"member/contrib/helper"
	"member/contrib/validator"
	"member/model"

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

	vs, err := model.MemberRebateGetCache(mb.UID)
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
	tyTemp := string(ctx.PostArgs().Peek("ty"))
	zrTemp := string(ctx.PostArgs().Peek("zr"))
	qpTemp := string(ctx.PostArgs().Peek("qp"))
	djTemp := string(ctx.PostArgs().Peek("dj"))
	dzTemp := string(ctx.PostArgs().Peek("dz"))
	cpTemp := string(ctx.PostArgs().Peek("cp"))
	fcTemp := string(ctx.PostArgs().Peek("fc"))
	byTemp := string(ctx.PostArgs().Peek("by"))
	cgHighRebateTemp := string(ctx.PostArgs().Peek("cg_high_rebate"))
	cgOfficialRebateTemp := string(ctx.PostArgs().Peek("cg_official_rebate"))

	//fmt.Println("Update = ", string(ctx.PostBody()))

	if !validator.CheckUName(subName, 5, 14) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	mb, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	// 在推广链接黑名单中，不允许新增
	//ok, err := model.MemberLinkBlacklist(mb.Username)
	//if err != nil {
	//	helper.Print(ctx, false, err.Error())
	//	return
	//}
	//
	//if ok {
	//	helper.Print(ctx, false, helper.NotAllowModifySubRebateErr)
	//	return
	//}

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

	ty, err := decimal.NewFromString(tyTemp) //下级会员体育返水比例
	if err != nil || ty.IsNegative() || ty.GreaterThan(parent.TY) {
		fmt.Println("tyTemp = ", tyTemp)
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	zr, err := decimal.NewFromString(zrTemp) //下级会员真人返水比例
	if err != nil || zr.IsNegative() || zr.GreaterThan(parent.ZR) {

		fmt.Println("zrTemp = ", zrTemp)
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	qp, err := decimal.NewFromString(qpTemp) //下级会员棋牌返水比例
	if err != nil || qp.IsNegative() || qp.GreaterThan(parent.QP) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	dj, err := decimal.NewFromString(djTemp) //下级会员电竞返水比例
	if err != nil || dj.IsNegative() || dj.GreaterThan(parent.DJ) {

		fmt.Println("djTemp = ", djTemp)
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	dz, err := decimal.NewFromString(dzTemp) //下级会员电子返水比例
	if err != nil || dz.IsNegative() || dz.GreaterThan(parent.DZ) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	cp, err := decimal.NewFromString(cpTemp) //下级会员彩票返水比例
	if err != nil || cp.IsNegative() || cp.GreaterThan(parent.CP) {

		fmt.Println("cpTemp = ", cpTemp)
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	fc, err := decimal.NewFromString(fcTemp) //下级会员斗鸡返水比例
	if err != nil || fc.IsNegative() || fc.GreaterThan(parent.FC) {

		fmt.Println("fcTemp = ", fcTemp)
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	by, err := decimal.NewFromString(byTemp) //下级会员捕鱼返水比例
	if err != nil || by.IsNegative() || by.GreaterThan(parent.BY) {

		fmt.Println("byTemp = ", byTemp)
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	cgHighRebate, err := decimal.NewFromString(cgHighRebateTemp)
	if err != nil || fc.IsNegative() || cgHighRebate.GreaterThan(parent.CGHighRebate) {

		fmt.Println("cgHighRebateTemp = ", cgHighRebateTemp)
		helper.Print(ctx, false, helper.RebateOutOfRange)
	}
	cgOfficialRebate, err := decimal.NewFromString(cgOfficialRebateTemp)
	if err != nil || fc.IsNegative() || cgOfficialRebate.GreaterThan(parent.CGOfficialRebate) {

		fmt.Println("cgOfficialRebateTemp = ", cgOfficialRebateTemp)
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
