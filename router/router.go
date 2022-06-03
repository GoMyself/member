package router

import (
	"fmt"
	"member/contrib/helper"
	"member/controller"
	"runtime/debug"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

var (
	ApiTimeoutMsg = `{"status": "false","data":"服务器响应超时，请稍后重试"}`
	ApiTimeout    = time.Second * 16
	router        *fasthttprouter.Router
	buildInfo     BuildInfo
)

type BuildInfo struct {
	Name           string
	GitReversion   string
	BuildTime      string
	BuildGoVersion string
	Now            string
}

func apiServerPanic(ctx *fasthttp.RequestCtx, rcv interface{}) {

	err := rcv.(error)
	fmt.Println(err)
	debug.PrintStack()

	if r := recover(); r != nil {
		fmt.Println("recovered failed", r)
	}

	ctx.SetStatusCode(500)
	return
}

func Version(ctx *fasthttp.RequestCtx) {

	buildInfo.Name = "Member"
	buildInfo.Now = time.Now().Format("2006-01-02 15:04:05")
	helper.Print(ctx, true, buildInfo)
	/*
		ctx.SetContentType("text/html; charset=utf-8")
		fmt.Fprintf(ctx, "{merchant<br />Git reversion = %s<br />Build Time = %s<br />Go version = %s<br />System Time = %s<br />",
			buildInfo.GitReversion, buildInfo.BuildTime, buildInfo.BuildGoVersion, ctx.Time())
	*/
	//ctx.Request.Header.VisitAll(func (key, value []byte) {
	//	fmt.Fprintf(ctx, "%s: %s<br/>", string(key), string(value))
	//})
}

func Ip(ctx *fasthttp.RequestCtx) {

	str := helper.FromRequest(ctx)
	ctx.SetBody([]byte(str))
}

// SetupRouter 设置路由列表
func SetupRouter(b BuildInfo) *fasthttprouter.Router {

	router = fasthttprouter.New()
	router.PanicHandler = apiServerPanic

	// 会员信息
	memberCtl := new(controller.MemberController)
	//银行卡管理
	bankCtl := new(controller.BankcardController)
	// 广告页
	bannerCtl := new(controller.BannerController)
	// app升级
	upgradeCtl := new(controller.UpgradeController)
	// 公共信息
	treeCtl := new(controller.TreeController)
	// 电子游戏
	slotCtl := new(controller.SlotController)
	// 公告
	noticeCtl := new(controller.NoticeController)
	// 咋内信
	msgCtl := new(controller.MessageController)
	// 推广链接
	linkCtl := new(controller.LinkController)
	// 返水
	rebateCtl := new(controller.RebateController)
	// 交易记录
	recordCtl := new(controller.RecordController)
	// 报表
	reportCtl := new(controller.ReportController)
	// vip
	vipCtl := new(controller.VipController)

	get("/member/version", Version)

	post("/member/link/insert", linkCtl.Insert)
	get("/member/link/delete", linkCtl.Delete)
	get("/member/link/list", linkCtl.List)

	// vip配置
	get("/member/vip/config", vipCtl.Config)
	// 会员存款流水
	get("/member/vip/info", vipCtl.Info)

	// 会员信息
	get("/member/token", memberCtl.Token)
	get("/member/info", memberCtl.Info)
	// 检测会员账号是否可用
	get("/member/available", memberCtl.Available)
	// 会员验证码
	get("/member/captcha", memberCtl.Captcha)
	// 会员注册
	post("/member/reg", memberCtl.Reg)
	// 会员登陆
	post("/member/login", memberCtl.Login)
	// 新增下级
	//post("/member/insert", memberCtl.Insert)
	// 会员余额信息 中心钱包和锁定钱包
	get("/member/balance", memberCtl.Balance)
	// 检测提款密码
	get("/member/password/check", memberCtl.CheckPwd)
	// 用户忘记密码
	post("/member/password/forget", memberCtl.ForgetPassword)
	// 用户修改密码
	post("/member/password/update", memberCtl.UpdatePassword)
	// 绑定手机号
	post("/member/bindphone", memberCtl.BindPhone)
	// 绑定邮箱
	post("/member/bindemail", memberCtl.BindEmail)
	// 绑定zalo号
	post("/member/bindzalo", memberCtl.BindZalo)
	// 更新用户信息 （真实姓名/收货地址）
	post("/member/update", memberCtl.Update)
	//场馆列表
	get("/member/platform", memberCtl.Plat)
	//导航栏列表
	get("/member/nav", memberCtl.Nav)
	// 会员列表
	get("/member/list", memberCtl.List)

	get("/member/avatar/update", memberCtl.Avatar)

	//新增银行卡
	post("/member/bankcard/insert", bankCtl.Insert)
	//查询银行卡
	get("/member/bankcard/list", bankCtl.List)

	// 获取广告页
	get("/member/banner", bannerCtl.Images)

	// app升级检测
	get("/member/app/upgrade", upgradeCtl.Info)

	get("/member/tree", treeCtl.List)

	//电子游戏大厅列表
	get("/member/slot/list", slotCtl.List)
	//电子游戏大厅关键字搜索
	post("/member/slot/search", slotCtl.Search)
	//电子游戏奖金池
	get("/member/slot/bonus", slotCtl.BonusPool)

	// 获取游戏投注记录
	get("/member/record/game", recordCtl.Game)
	//转账记录
	get("/member/record/transfer", recordCtl.Transfer)
	// 账变记录列表
	get("/member/record/transaction", recordCtl.Transaction)
	// 交易记录
	get("/member/record/trade", recordCtl.Trade)
	// 交易记录详情
	get("/member/record/tradedetail", recordCtl.TradeDetail)

	// 公告跑马灯
	get("/member/notices", noticeCtl.List)

	// 站内信-列表
	get("/member/message/list", msgCtl.List)
	// 站内信-读取
	get("/member/message/read", msgCtl.Read)
	// 站内信-未读数
	get("/member/message/num", msgCtl.Num)
	// 站内信-删除
	get("/member/message/delete", msgCtl.Delete)

	// 编辑会员密码，返水
	get("/member/rebate/detail", rebateCtl.Detail)
	post("/member/rebate/update", rebateCtl.Update)
	// 获取返水上限
	get("/member/rebate/scale", rebateCtl.Scale)

	// 代理报表
	post("/member/agency/report", reportCtl.Report)
	buildInfo = b

	return router
}

// get is a shortcut for router.GET(path string, handle fasthttp.RequestHandler)
func get(path string, handle fasthttp.RequestHandler) {
	router.GET(path, fasthttp.TimeoutHandler(handle, ApiTimeout, ApiTimeoutMsg))
}

// head is a shortcut for router.HEAD(path string, handle fasthttp.RequestHandler)
func head(path string, handle fasthttp.RequestHandler) {
	router.HEAD(path, fasthttp.TimeoutHandler(handle, ApiTimeout, ApiTimeoutMsg))
}

// options is a shortcut for router.OPTIONS(path string, handle fasthttp.RequestHandler)
func options(path string, handle fasthttp.RequestHandler) {
	router.OPTIONS(path, fasthttp.TimeoutHandler(handle, ApiTimeout, ApiTimeoutMsg))
}

// post is a shortcut for router.POST(path string, handle fasthttp.RequestHandler)
func post(path string, handle fasthttp.RequestHandler) {
	router.POST(path, fasthttp.TimeoutHandler(handle, ApiTimeout, ApiTimeoutMsg))
}

// put is a shortcut for router.PUT(path string, handle fasthttp.RequestHandler)
func put(path string, handle fasthttp.RequestHandler) {
	router.PUT(path, fasthttp.TimeoutHandler(handle, ApiTimeout, ApiTimeoutMsg))
}

// patch is a shortcut for router.PATCH(path string, handle fasthttp.RequestHandler)
func patch(path string, handle fasthttp.RequestHandler) {
	router.PATCH(path, fasthttp.TimeoutHandler(handle, ApiTimeout, ApiTimeoutMsg))
}

// delete is a shortcut for router.DELETE(path string, handle fasthttp.RequestHandler)
func delete(path string, handle fasthttp.RequestHandler) {
	router.DELETE(path, fasthttp.TimeoutHandler(handle, ApiTimeout, ApiTimeoutMsg))
}
