package controller

import (
	"github.com/valyala/fasthttp"
	"member2/contrib/helper"
	"member2/contrib/validator"
	"member2/model"
	"strings"
)

type EmailController struct{}

func (that *EmailController) Send(ctx *fasthttp.RequestCtx) {

	flag := ctx.PostArgs().GetUintOrZero("flag")
	username := string(ctx.PostArgs().Peek("username"))
	address := strings.ToLower(string(ctx.PostArgs().Peek("address")))
	vid := string(ctx.PostArgs().Peek("vid"))
	code := string(ctx.PostArgs().Peek("code"))

	if flag != model.EmailForgetPassword && flag != model.EmailModifyPassword {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if flag == model.EmailModifyPassword {
		username = string(ctx.UserValue("token").([]byte))
		if username == "" {
			helper.Print(ctx, false, helper.AccessTokenExpires)
			return
		}
	} else if flag == model.EmailForgetPassword {
		if ok := model.MemberVerify(vid, code); !ok {
			helper.Print(ctx, false, helper.EmailVerificationErr)
			return
		}

		if !validator.CheckUName(username, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}

		if !strings.ContainsAny(address, "@") {
			helper.Print(ctx, false, helper.EmailFMTErr)
			return
		}
	}

	day := ctx.Time().Format("0102")
	ip := helper.FromRequest(ctx)
	id, err := model.EmailSend(day, ip, username, address, flag)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, id)
}
