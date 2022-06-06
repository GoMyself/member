package controller

import (
	"fmt"
	"member/contrib/helper"
	"member/model"

	"github.com/valyala/fasthttp"
)

type BankcardController struct{}

// 绑定银行卡
func (that *BankcardController) Insert(ctx *fasthttp.RequestCtx) {

	bankID := string(ctx.PostArgs().Peek("bank_id"))
	bankcardNo := string(ctx.PostArgs().Peek("bankcard_no"))
	bankAddress := string(ctx.PostArgs().Peek("bank_addr"))
	realname := string(ctx.PostArgs().Peek("realname"))
	phone := string(ctx.PostArgs().Peek("phone"))

	data := model.BankCard{
		ID:          helper.GenId(),
		BankID:      bankID,
		Username:    string(ctx.UserValue("token").([]byte)),
		BankBranch:  bankAddress,
		BankAddress: bankAddress,
		CreatedAt:   fmt.Sprintf("%d", ctx.Time().Unix()),
	}

	// 用户绑定银行卡
	err := model.BankcardInsert(ctx, phone, realname, bankcardNo, data)
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
