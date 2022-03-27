package controller

import (
	"github.com/valyala/fasthttp"
	"member2/contrib/helper"
	"member2/contrib/validator"
	"member2/model"
)

type BankcardController struct{}

// 绑定银行卡
func (that *BankcardController) Insert(ctx *fasthttp.RequestCtx) {

	code := string(ctx.PostArgs().Peek("code"))
	sid := string(ctx.PostArgs().Peek("sid"))
	bankID := string(ctx.PostArgs().Peek("bank_id"))
	bankcardNo := string(ctx.PostArgs().Peek("bankcard_no"))
	bankAddress := string(ctx.PostArgs().Peek("bank_addr"))
	phone := string(ctx.PostArgs().Peek("phone"))
	realname := string(ctx.PostArgs().Peek("realname"))

	data := model.BankCard{
		ID:          helper.GenId(),
		BankID:      bankID,
		Username:    string(ctx.UserValue("token").([]byte)),
		BankBranch:  bankAddress,
		BankAddress: bankAddress,
		CreatedAt:   uint64(ctx.Time().Unix()),
	}

	// 用户绑定银行卡
	err := model.BankcardInsert(ctx, phone, sid, code, realname, bankcardNo, data)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *BankcardController) List(ctx *fasthttp.RequestCtx) {

	username := string(ctx.UserValue("token").([]byte))
	if username == "" {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	// 更新权限信息
	res, err := model.BankcardList(username)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, res)
}

func (that *BankcardController) Delete(ctx *fasthttp.RequestCtx) {

	username := string(ctx.UserValue("token").([]byte))
	if username == "" {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	bid := string(ctx.QueryArgs().Peek("bid"))
	if !validator.CheckStringDigit(bid) {
		helper.Print(ctx, false, helper.IDErr)
		return
	}

	// 删除会员银行卡
	err := model.BankcardDelete(username, bid)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
