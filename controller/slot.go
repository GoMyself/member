package controller

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"member/contrib/helper"
	"member/contrib/validator"
	"member/model"
)

type SlotController struct{}

type gameSearchParam struct {
	Pid      string `rule:"digit" msg:"pid error" name:"pid"`
	Keyword  string `rule:"none" name:"keyword"`
	Flag     int    `rule:"digit" default:"0" min:"0" max:"2" msg:"flag error" name:"flag"` //0:全部 1:最新 2:热门
	PageSize int    `rule:"digit" default:"10" min:"10" max:"200" msg:"page_size error" name:"page_size"`
	Page     int    `rule:"digit" default:"1" min:"1" msg:"page error" name:"page"`
}

type gameListParam struct {
	Pid      string `rule:"digit" msg:"pid error" name:"pid"`
	Flag     int    `rule:"digit" default:"0" min:"0" max:"2" msg:"flag error" name:"flag"` //0:全部 1:最新 2:热门
	PageSize int    `rule:"digit" default:"10" min:"10" max:"200" msg:"page_size error" name:"page_size"`
	Page     int    `rule:"digit" default:"1" min:"1" msg:"page error" name:"page"`
}

func (that *SlotController) List(ctx *fasthttp.RequestCtx) {

	params := gameListParam{}
	err := validator.Bind(ctx, &params)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	data, err := model.SlotList(params.Pid, params.Flag, params.Page, params.PageSize)
	if err != nil {
		helper.Print(ctx, false, err.Error())
	}

	helper.PrintJson(ctx, true, data)
}

func (that *SlotController) Search(ctx *fasthttp.RequestCtx) {

	params := gameSearchParam{}
	err := validator.Bind(ctx, &params)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if !validator.CheckStringLength(params.Keyword, 1, 50) {
		helper.Print(ctx, false, helper.ContentLengthErr)
		return
	}

	param := map[string]string{
		"pid":     params.Pid,
		"keyword": params.Keyword,
		"flag":    fmt.Sprintf("%d", params.Flag),
	}
	fmt.Println(param)

	data, err := model.SlotSearch(params.Page, params.PageSize, param)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.PrintJson(ctx, true, data)
}

// 电子游戏奖金池
func (that *SlotController) BonusPool(ctx *fasthttp.RequestCtx) {

	bonus, err := model.SlotBonusPool()
	if err != nil {
		helper.Print(ctx, false, err.Error())
	}

	helper.Print(ctx, true, bonus)
}
