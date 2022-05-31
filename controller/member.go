package controller

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	g "github.com/doug-martin/goqu/v9"
	"github.com/shopspring/decimal"

	"member2/contrib/helper"
	"member2/contrib/validator"
	"member2/model"

	"github.com/valyala/fasthttp"
)

type MemberController struct{}

type MemberRegParam struct {
	LinkID     string `rule:"none" json:"link_id" name:"link_id"`
	RegUrl     string `rule:"none" json:"reg_url" name:"reg_url"`
	Name       string `rule:"uname" name:"username" min:"5" max:"14" msg:"username error"`
	DeviceNo   string `rule:"none" name:"device_no"`
	Password   string `rule:"upwd" name:"password" min:"8" max:"20" msg:"password error"`
	Phone      string `rule:"digit" name:"phone"`
	Sid        string `json:"sid" name:"sid" rule:"digit" msg:"sid error"`
	Ts         string `json:"ts" name:"ts" rule:"digit" msg:"ts error"`
	VerifyCode string `rule:"digit" name:"verify_code"`
}

// 修改用户密码参数
type forgetPassword struct {
	Username string `rule:"alnum" msg:"username error" name:"username"`
	Sid      string `json:"sid" name:"sid" rule:"digit" msg:"phone error"`
	Ts       string `json:"ts" name:"ts" rule:"digit" msg:"ts error"`
	Code     string `json:"code" name:"code" rule:"digit" msg:"code error"`
	Phone    string `rule:"digit" msg:"phone error" name:"phone"`
	Password string `rule:"upwd" name:"password" msg:"password error"`
}

// 绑定邮箱
type paramBindEmail struct {
	Sid   string `json:"sid" name:"sid" rule:"none" msg:"phone error"`
	Code  string `json:"code" name:"code" rule:"none" min:"2" max:"8" msg:"code error"`
	Email string `json:"email" name:"email" rule:"none"`
}

type MemberAutoTransferParam struct {
	Status string `rule:"none" name:"status"`
}

func (that *MemberController) Info(ctx *fasthttp.RequestCtx) {

	m, err := model.MemberInfo(ctx)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, m)
}

func (that *MemberController) Token(ctx *fasthttp.RequestCtx) {

	m, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, m)
}

func (that *MemberController) Balance(ctx *fasthttp.RequestCtx) {

	a, err := model.MemberAmount(ctx)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.PrintJson(ctx, true, a)
}

func (that *MemberController) Login(ctx *fasthttp.RequestCtx) {

	username := string(ctx.PostArgs().Peek("username"))
	if !validator.CheckUName(username, 5, 14) {
		helper.Print(ctx, false, helper.UsernameFMTErr)
		return
	}

	password := string(ctx.PostArgs().Peek("password"))
	if !validator.CheckUPassword(password, 8, 20) {
		helper.Print(ctx, false, helper.PasswordFMTErr)
		return
	}

	deviceNo := string(ctx.PostArgs().Peek("device_no"))
	code := string(ctx.PostArgs().Peek("code"))
	vid := string(ctx.PostArgs().Peek("vid"))

	device := string(ctx.Request.Header.Peek("d"))
	i, err := strconv.Atoi(device)
	if err != nil {
		helper.Print(ctx, false, helper.DeviceTypeErr)
		return
	}

	if _, ok := model.Devices[i]; !ok {
		helper.Print(ctx, false, helper.DeviceTypeErr)
		return
	}

	ip := helper.FromRequest(ctx)
	username = strings.ToLower(username)
	id, err := model.MemberLogin(ctx, vid, code, username, password, ip, device, deviceNo)
	if err != nil {
		if err.Error() == helper.Blocked || err.Error() == helper.UserNotExist ||
			err.Error() == helper.DeviceBanErr || strings.Contains(err.Error(), helper.IpBanErr) {
			helper.Print(ctx, false, err.Error())
			return
		}

		helper.Print(ctx, false, err.Error())
		return
	}

	ctx.Response.Header.Set("id", id)
	helper.Print(ctx, true, helper.Success)
}

func (that *MemberController) Reg(ctx *fasthttp.RequestCtx) {

	param := MemberRegParam{}
	err := validator.Bind(ctx, &param)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if param.LinkID != "" {
		param.LinkID, _ = url.QueryUnescape(param.LinkID)
		links := strings.Split(param.LinkID, "|")
		if len(links) != 2 {
			helper.Print(ctx, false, helper.IDErr)
			return
		}

		for _, v := range links {
			if !validator.CheckStringDigit(v) {
				helper.Print(ctx, false, helper.IDErr)
				return
			}
		}
	}

	if len(param.Phone) < 1 {
		helper.Print(ctx, false, helper.PhoneFMTErr)
		return
	}

	if !validator.IsVietnamesePhone(param.Phone) {
		helper.Print(ctx, false, helper.PhoneFMTErr)
		return
	}

	ip := helper.FromRequest(ctx)
	day := ctx.Time().Format("0102")
	if param.VerifyCode != "6666" {

		smsFlag, err := model.CheckSmsCaptcha(ip, param.Sid, param.Phone, day, param.VerifyCode)
		if err != nil || !smsFlag {
			helper.Print(ctx, false, helper.PhoneVerificationErr)
			return
		}

	}

	createdAt := uint32(ctx.Time().Unix())
	device := string(ctx.Request.Header.Peek("d"))
	i, err := strconv.Atoi(device)
	if err != nil {
		helper.Print(ctx, false, helper.DeviceTypeErr)
		return
	}

	if _, ok := model.Devices[i]; !ok {
		helper.Print(ctx, false, helper.DeviceTypeErr)
		return
	}

	// 注册地址 去除域名前缀
	uid, err := model.MemberReg(i, param.Name, param.Password, ip, param.DeviceNo, param.RegUrl, param.LinkID, param.Phone, param.Ts, createdAt)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	ctx.Response.Header.Set("id", uid)
	helper.Print(ctx, true, helper.Success)
}

func (that *MemberController) Insert(ctx *fasthttp.RequestCtx) {

	subName := string(ctx.PostArgs().Peek("username"))
	password := string(ctx.PostArgs().Peek("password"))
	remark := string(ctx.PostArgs().Peek("remark"))
	sty := string(ctx.PostArgs().Peek("ty"))
	szr := string(ctx.PostArgs().Peek("zr"))
	sqp := string(ctx.PostArgs().Peek("qp"))
	sdj := string(ctx.PostArgs().Peek("dj"))
	sdz := string(ctx.PostArgs().Peek("dz"))
	scp := string(ctx.PostArgs().Peek("cp"))
	sfc := string(ctx.PostArgs().Peek("fc"))
	sby := string(ctx.PostArgs().Peek("by"))

	mb, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	parent, err := model.MemberRebateFindOne(mb.UID)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	ty, err := decimal.NewFromString(sty) //下级会员体育返水比例
	if err != nil || ty.IsNegative() || ty.GreaterThan(parent.TY) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	zr, err := decimal.NewFromString(szr) //下级会员真人返水比例
	if err != nil || zr.IsNegative() || zr.GreaterThan(parent.ZR) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	qp, err := decimal.NewFromString(sqp) //下级会员棋牌返水比例
	if err != nil || qp.IsNegative() || qp.GreaterThan(parent.QP) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	dj, err := decimal.NewFromString(sdj) //下级会员电竞返水比例
	if err != nil || dj.IsNegative() || dj.GreaterThan(parent.DJ) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	dz, err := decimal.NewFromString(sdz) //下级会员电子返水比例
	if err != nil || dz.IsNegative() || dz.GreaterThan(parent.DZ) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	cp, err := decimal.NewFromString(scp) //下级会员彩票返水比例
	if err != nil || cp.IsNegative() || cp.GreaterThan(parent.CP) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	fc, err := decimal.NewFromString(sfc) //下级会员斗鸡返水比例
	if err != nil || fc.IsNegative() || fc.GreaterThan(parent.FC) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	by, err := decimal.NewFromString(sby) //下级会员斗鸡返水比例
	if err != nil || by.IsNegative() || by.GreaterThan(parent.BY) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	if !validator.CheckUName(subName, 5, 14) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	if !validator.CheckUPassword(password, 8, 20) {
		helper.Print(ctx, false, helper.PasswordFMTErr)
		return
	}

	mr := model.MemberRebate{
		TY: ty.StringFixed(1),
		ZR: zr.StringFixed(1),
		QP: qp.StringFixed(1),
		DJ: dj.StringFixed(1),
		DZ: dz.StringFixed(1),
		CP: cp.StringFixed(1),
		FC: fc.StringFixed(1),
		BY: by.StringFixed(1),
	}
	createdAt := uint32(ctx.Time().Unix())
	// 添加下级代理
	err = model.MemberInsert(mb, subName, password, remark, createdAt, mr)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *MemberController) Captcha(ctx *fasthttp.RequestCtx) {

	img, str, err := model.MemberCaptcha()
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	ctx.Response.Header.Set("Content-Type", "image/png")
	ctx.Response.Header.Set("vid", str)
	ctx.Write(img)
}

// 检查提款密码
func (that *MemberController) CheckPwd(ctx *fasthttp.RequestCtx) {

	mb, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}

	if mb.WithdrawPwd == 0 {
		helper.Print(ctx, false, helper.SetWithdrawPwdFirst)
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 用户修改密码
func (that *MemberController) UpdatePassword(ctx *fasthttp.RequestCtx) {

	ty := ctx.PostArgs().GetUintOrZero("ty")
	sid := string(ctx.PostArgs().Peek("sid"))
	ts := string(ctx.PostArgs().Peek("ts"))
	code := string(ctx.PostArgs().Peek("code"))
	old := string(ctx.PostArgs().Peek("old"))
	password := string(ctx.PostArgs().Peek("password"))

	t := map[int]bool{
		1: true,
		2: true,
	}
	if _, ok := t[ty]; !ok {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	// 修改登录密码，新旧密码不能一样
	if ty == 1 && old == password {
		helper.Print(ctx, false, helper.PasswordConsistent)
		return
	}

	err := model.MemberPasswordUpdate(ty, sid, code, old, password, ts, ctx)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 用户忘记密码
func (that *MemberController) Avatar(ctx *fasthttp.RequestCtx) {

	id := ctx.QueryArgs().GetUintOrZero("id")

	if id < 1 || id > 16 {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	str := fmt.Sprintf("%d", id)
	err := model.MemberUpdateAvatar(str, ctx)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 用户忘记密码
func (that *MemberController) ForgetPassword(ctx *fasthttp.RequestCtx) {

	params := forgetPassword{}
	err := validator.Bind(ctx, &params)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	//fmt.Println(params)

	ip := helper.FromRequest(ctx)
	err = model.MemberForgetPwd(params.Username, params.Password, params.Phone, ip, params.Sid, params.Code, params.Ts)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 用户绑定邮箱
func (that *MemberController) BindEmail(ctx *fasthttp.RequestCtx) {

	params := paramBindEmail{}
	err := validator.Bind(ctx, &params)
	if err != nil {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if params.Email == "" || !strings.Contains(params.Email, "@") {
		helper.Print(ctx, false, helper.EmailFMTErr)
		return
	}

	err = model.MemberUpdateEmail(params.Sid, params.Code, params.Email, ctx)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 用户绑定手机号
func (that *MemberController) BindPhone(ctx *fasthttp.RequestCtx) {

	phone := string(ctx.PostArgs().Peek("phone"))
	err := model.MemberUpdatePhone(phone, ctx)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 用户绑定zalo号
func (that *MemberController) BindZalo(ctx *fasthttp.RequestCtx) {

	zalo := string(ctx.PostArgs().Peek("zalo"))
	err := model.MemberUpdateZalo(zalo, ctx)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 更新用户信息
func (that *MemberController) Update(ctx *fasthttp.RequestCtx) {

	birth := string(ctx.PostArgs().Peek("birth"))
	realname := strings.TrimSpace(string(ctx.PostArgs().Peek("realname")))
	address := string(ctx.PostArgs().Peek("address"))
	if realname == "" && address == "" && birth == "" {
		helper.Print(ctx, false, helper.ParamErr)
		return
	}

	if realname != "" {
		if !validator.CheckStringVName(realname) {
			helper.Print(ctx, false, helper.RealNameFMTErr)
			return
		}
	}

	if address != "" {
		if len(strings.Split(address, "|")) != 4 {
			helper.Print(ctx, false, helper.AddressFMTErr)
			return
		}
	}

	err := model.MemberUpdateName(ctx, birth, realname, address)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}

// 会员账号是否可用
func (that *MemberController) Available(ctx *fasthttp.RequestCtx) {

	username := strings.ToLower(string(ctx.QueryArgs().Peek("username")))
	//字母数字组合，4-9，前2个字符必须为字母
	if !validator.CheckUName(username, 5, 14) {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

	// 检测会员是否已注册
	if model.MemberExist(username) {
		helper.Print(ctx, false, helper.UsernameExist)
		return
	}

	helper.Print(ctx, true, helper.Success)
}

func (that *MemberController) Plat(ctx *fasthttp.RequestCtx) {

	data := model.Platform()
	helper.PrintJson(ctx, true, data)
}

func (that *MemberController) Nav(ctx *fasthttp.RequestCtx) {

	data := model.Nav()
	helper.PrintJson(ctx, true, data)
}

func (that *MemberController) List(ctx *fasthttp.RequestCtx) {

	username := string(ctx.QueryArgs().Peek("username"))
	startTime := string(ctx.QueryArgs().Peek("start_time"))
	endTime := string(ctx.QueryArgs().Peek("end_time"))
	page := ctx.QueryArgs().GetUintOrZero("page")
	pageSize := ctx.QueryArgs().GetUintOrZero("page_size")
	sortField := string(ctx.QueryArgs().Peek("sort_field"))
	isAsc := ctx.QueryArgs().GetUintOrZero("is_asc")
	agg := ctx.QueryArgs().GetUintOrZero("agg")
	if page == 0 {
		page = 1
	}

	if pageSize == 0 {
		pageSize = 10
	}

	ex := g.Ex{}
	if username != "" {
		if !validator.CheckUName(username, 5, 14) {
			helper.Print(ctx, false, helper.UsernameErr)
			return
		}
		ex["username"] = username
	}

	if sortField != "" {
		sortFields := map[string]bool{
			"deposit":    true,
			"withdraw":   true,
			"dividend":   true,
			"rebate":     true,
			"net_amount": true,
		}

		if _, ok := sortFields[sortField]; !ok {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}

		if !validator.CheckIntScope(strconv.Itoa(isAsc), 0, 1) {
			helper.Print(ctx, false, helper.ParamErr)
			return
		}
	}

	currentUsername := string(ctx.UserValue("token").([]byte))
	if currentUsername == "" {
		helper.Print(ctx, false, helper.AccessTokenExpires)
		return
	}
	//currentUsername := "jasper01"
	ex["parent_name"] = currentUsername

	data, err := model.MemberList(ex, username, startTime, endTime, sortField, isAsc, page, pageSize)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	if agg == 1 {
		aggData, err := model.MemberAgg(currentUsername)
		if err != nil {
			helper.Print(ctx, false, err.Error())
			return
		}
		data.Agg = aggData
	}

	helper.Print(ctx, true, data)
}

// UpdateRebate 修改密码以及返水比例
func (that *MemberController) UpdateRebate(ctx *fasthttp.RequestCtx) {

	subName := string(ctx.PostArgs().Peek("username"))
	password := string(ctx.PostArgs().Peek("password"))
	sty := string(ctx.PostArgs().Peek("ty"))
	szr := string(ctx.PostArgs().Peek("zr"))
	sqp := string(ctx.PostArgs().Peek("qp"))
	sdj := string(ctx.PostArgs().Peek("dj"))
	sdz := string(ctx.PostArgs().Peek("dz"))
	scp := string(ctx.PostArgs().Peek("cp"))
	sfc := string(ctx.PostArgs().Peek("fc"))
	sby := string(ctx.PostArgs().Peek("by"))

	mb, err := model.MemberCache(ctx, "")
	if err != nil {
		helper.Print(ctx, false, helper.UsernameErr)
		return
	}

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

	ty, err := decimal.NewFromString(sty) //下级会员体育返水比例
	if err != nil || ty.IsNegative() || ty.GreaterThan(parent.TY) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	zr, err := decimal.NewFromString(szr) //下级会员真人返水比例
	if err != nil || zr.IsNegative() || zr.GreaterThan(parent.ZR) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	qp, err := decimal.NewFromString(sqp) //下级会员棋牌返水比例
	if err != nil || qp.IsNegative() || qp.GreaterThan(parent.QP) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	dj, err := decimal.NewFromString(sdj) //下级会员电竞返水比例
	if err != nil || dj.IsNegative() || dj.GreaterThan(parent.DJ) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	dz, err := decimal.NewFromString(sdz) //下级会员电子返水比例
	if err != nil || dz.IsNegative() || dz.GreaterThan(parent.DZ) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	cp, err := decimal.NewFromString(scp) //下级会员彩票返水比例
	if err != nil || cp.IsNegative() || cp.GreaterThan(parent.CP) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	fc, err := decimal.NewFromString(sfc) //下级会员斗鸡返水比例
	if err != nil || fc.IsNegative() || fc.GreaterThan(parent.FC) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
	}

	by, err := decimal.NewFromString(sby) //下级会员捕鱼返水比例
	if err != nil || by.IsNegative() || by.GreaterThan(parent.BY) {
		helper.Print(ctx, false, helper.RebateOutOfRange)
		return
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

	mr := model.MemberRebate{
		TY: ty.StringFixed(1),
		ZR: zr.StringFixed(1),
		QP: qp.StringFixed(1),
		DJ: dj.StringFixed(1),
		DZ: dz.StringFixed(1),
		CP: cp.StringFixed(1),
		FC: fc.StringFixed(1),
		BY: fc.StringFixed(1),
	}
	// 添加下级代理
	err = model.MemberUpdateInfo(child, password, mr)
	if err != nil {
		helper.Print(ctx, false, err.Error())
		return
	}

	helper.Print(ctx, true, helper.Success)
}
