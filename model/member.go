package model

import (
	"database/sql"
	"errors"
	"fmt"
	"member2/contrib/helper"
	"member2/contrib/session"
	"member2/contrib/tdlog"
	"member2/contrib/validator"
	"strconv"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9/exp"

	g "github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
	"github.com/valyala/fasthttp"
)

type Game struct {
	Id         string `db:"id" json:"id"`                   //
	PlatformId string `db:"platform_id" json:"platform_id"` // 平台id,
	Name       string `db:"name" json:"name"`               // 游戏名称,
	EnName     string `db:"en_name" json:"en_name"`         // 英文名称,
	ClientType string `db:"client_type" json:"client_type"` // 0:all 1:web 2:h5 4:app\n此处值为支持端的数值之和,
	GameType   string `db:"game_type" json:"game_type"`     // 游戏类型:2=真人,3=捕鱼,4=电子,5=彩票,6=体育,7=棋牌,8=电竞,
	GameId     string `db:"game_id" json:"game_id"`         // 游戏id,
	ImgPhone   string `db:"img_phone" json:"img_phone"`     // 手机图片,
	ImgPc      string `db:"img_pc" json:"img_pc"`           // pc图片,
	ImgCover   string `db:"img_cover" json:"img_cover"`     // 游戏封面,
	OnLine     string `db:"on_line" json:"on_line"`         // 0 下线 1上线,
	IsHot      string `db:"is_hot" json:"is_hot"`           // 0正常1热门,
	IsFs       string `db:"is_fs" json:"is_fs"`             // 是否参与反水 0 不参与 1参与,
	Sorting    uint   `db:"sorting" json:"sorting"`         // 排序,
	CreatedAt  uint   `db:"created_at" json:"created_at"`   //
	IsNew      string `db:"is_new" json:"is_new"`           // 是否最新:0=否,1=是,
	GameCode   string `db:"game_code" json:"game_code"`     // 游戏编码,对应投注记录中的游戏,
	VnAlias    string `db:"vn_alias" json:"vn_alias"`       // 越南别名,
}

func MemberAmount(ctx *fasthttp.RequestCtx) (string, error) {

	m, err := MemberCache(ctx, "")
	if err != nil {
		return "", errors.New(helper.AccessTokenExpires)
	}

	mb := MBBalance{
		Balance:    m.Balance,
		Commission: m.Commission,
		LockAmount: m.LockAmount,
	}
	data, err := helper.JsonMarshal(mb)
	if err != nil {
		return "", errors.New(helper.FormatErr)
	}

	return string(data), nil
}

func MemberLogin(vid, code, username, password, ip, device, deviceNo string, lastLoginAt uint32) (string, error) {

	// 检查ip黑名单
	idx := MurmurHash(ip, 0) % 10
	key := fmt.Sprintf("bl:ip%d", idx)
	ok, err := meta.MerchantRedis.SIsMember(ctx, key, ip).Result()
	if err != nil {
		return "", pushLog(err, helper.RedisErr)
	}

	if ok {
		return "", errors.New(fmt.Sprintf("%s,%s", helper.IpBanErr, ip))
	}

	// web/h5不检查设备号黑名单
	if device != "24" && device != "25" {

		// 检查设备号黑名单
		idx = MurmurHash(deviceNo, 0) % 10
		key = fmt.Sprintf("bl:dev%d", idx)
		ok, err = meta.MerchantRedis.SIsMember(ctx, key, deviceNo).Result()
		if err != nil {
			return "", pushLog(err, helper.RedisErr)
		}

		if ok {
			return "", errors.New(helper.DeviceBanErr)
		}
	}

	// 处理会员输入错误密码逻辑
	key = fmt.Sprintf("MPE:%s", username)
	errTimes, err := meta.MerchantRedis.Get(ctx, key).Int64()
	if err != nil && err != redis.Nil {
		return "", errors.New(helper.RedisErr)
	}

	if errTimes > 19 {
		return "", errors.New(helper.Blocked)
	}

	if errTimes > 5 && !MemberVerify(vid, code) {
		errTimes = meta.MerchantRedis.Incr(ctx, key).Val()
		if errTimes > 19 {
			meta.MerchantRedis.ExpireAt(ctx, key, time.Now().Add(60*time.Minute)).Result()
		}
		return "", fmt.Errorf("code|%d", errTimes)
	}

	mb, err := MemberCache(nil, username)
	if err != nil {
		return "", errors.New(helper.UserNotExist)
	}

	if mb.UID == "" {
		return "", errors.New(helper.UserNotExist)
	}

	if mb.State == 2 {
		return "", errors.New(helper.Blocked)
	}

	pwd := fmt.Sprintf("%d", MurmurHash(password, mb.CreatedAt))
	if pwd != mb.Password {
		errTimes = meta.MerchantRedis.Incr(ctx, key).Val()
		if errTimes > 19 {

			_, err = meta.MerchantRedis.ExpireAt(ctx, key, time.Now().Add(60*time.Minute)).Result()
			if err != nil {
				return "", pushLog(err, helper.RedisErr)
			}
		}

		return "", fmt.Errorf("password|%d", errTimes)
	}

	ex := g.Ex{"uid": mb.UID}
	record := g.Record{
		"last_login_ip":     ip,
		"last_login_at":     lastLoginAt,
		"last_login_device": deviceNo,
		"last_login_source": device,
	}

	t := dialect.Update("tbl_members")
	query, _, _ := t.Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return "", err
	}

	log := map[string]string{
		"username":  mb.Username,
		"ip":        ip,
		"device":    device,
		"device_no": deviceNo,
		"parents":   mb.ParentName,
	}
	err = tdlog.WriteLog("member_login_log", log)
	if err != nil {
		fmt.Printf("member write member_login_log error : [%s]/n", err.Error())
	}

	l := MemberLoginLog{
		Username: username,
		IPS:      ip,
		Device:   device,
		DeviceNo: deviceNo,
		Date:     lastLoginAt,
		Parents:  mb.ParentName,
		Prefix:   meta.Prefix,
	}
	err = meta.Zlog.Post(esPrefixIndex("memberlogin"), l)
	if err != nil {
		fmt.Printf("zlog error : %v data : %#v\n", err, l)
	}

	sid, err := session.Set([]byte(username), mb.UID)
	if err != nil {
		return "", errors.New(helper.SessionErr)
	}

	_, err = meta.MerchantRedis.Unlink(ctx, key).Result()
	if err != nil {
		return "", pushLog(err, helper.RedisErr)
	}

	return sid, nil
}

func MemberReg(device int, username, password, ip, deviceNo, regUrl, linkID, phone string, createdAt uint32) (string, error) {

	// 检查ip黑名单
	idx := MurmurHash(ip, 0) % 10
	key := fmt.Sprintf("bl:ip%d", idx)
	topId := "4722355249852325"

	ok, err := meta.MerchantRedis.SIsMember(ctx, key, ip).Result()
	if err != nil {
		return "", pushLog(err, helper.RedisErr)
	}

	if ok {
		return "", errors.New(fmt.Sprintf("%s,%s", helper.IpBanErr, ip))
	}

	phoneHash := fmt.Sprintf("%d", MurmurHash(phone, 0))
	phone_exist := meta.MerchantRedis.Do(ctx, "CF.EXISTS", "phone_exist", phone).Val()
	if phone_exist == "1" {
		return "", errors.New(helper.PhoneExist)
	}
	/*
		//检查手机是否已经存在
		phoneHash := fmt.Sprintf("%d", MurmurHash(phone, 0))
		ex := g.Ex{
			"phone_hash": phoneHash,
		}
		if MemberBindCheck(ex) {
			return "", errors.New(helper.PhoneExist)
		}
	*/
	// web/h5不检查设备号黑名单
	if _, ok = WebDevices[device]; !ok {

		// 检查设备号黑名单
		idx = MurmurHash(deviceNo, 0) % 10
		key = fmt.Sprintf("bl:dev%d", idx)
		ok, err = meta.MerchantRedis.SIsMember(ctx, key, deviceNo).Result()
		if err != nil {
			return "", pushLog(err, helper.RedisErr)
		}

		if ok {
			return "", errors.New(helper.DeviceBanErr)
		}
	}

	userName := strings.ToLower(username)
	if MemberExist(userName) {
		return "", errors.New(helper.UsernameExist)
	}

	sourceID := devices[device]
	lastLoginSource := device
	uid := helper.GenId()
	m := Member{
		UID:                uid,
		Username:           userName,
		Password:           fmt.Sprintf("%d", MurmurHash(password, createdAt)),
		PhoneHash:          phoneHash,
		EmailHash:          "0",
		RealnameHash:       "0",
		ZaloHash:           "0",
		Prefix:             meta.Prefix,
		State:              1,
		Regip:              ip,
		RegDevice:          deviceNo,
		RegUrl:             regUrl,
		CreatedAt:          createdAt,
		LastLoginIp:        ip,
		LastLoginAt:        createdAt,
		SourceId:           sourceID,
		FirstDepositAmount: "0.000",
		FirstBetAmount:     "0.000",
		Balance:            "0.000",
		LockAmount:         "0.000",
		Commission:         "0.000",
		LastLoginDevice:    deviceNo,
		LastLoginSource:    lastLoginSource,
		Level:              1,
	}

	tx, err := meta.MerchantDB.Begin() // 开启事务
	if err != nil {
		return "", pushLog(err, helper.DBErr)
	}

	fmt.Println("regLink id : ", linkID)
	parent := Member{}
	var query string
	// 邀请链接注册，不成功注册在默认代理root下
	parent, query, err = regLink(uid, linkID, createdAt)
	if err != nil {
		parent, query, err = regRoot(uid, topId, createdAt)
		if err != nil {
			_ = tx.Rollback()
			return "", err
		}
	}

	m.ParentUid = parent.UID
	m.ParentName = parent.Username
	m.TopUid = parent.TopUid
	m.TopName = parent.TopName
	m.AgencyType = parent.AgencyType
	m.GroupName = parent.GroupName

	// 插入返水记录
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return "", pushLog(err, helper.DBErr)
	}

	query, _, _ = dialect.Insert("tbl_members").Rows(&m).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return "", pushLog(err, helper.DBErr)
	}

	treeNode := MemberClosureInsert(m.UID, m.ParentUid)
	_, err = tx.Exec(treeNode)
	if err != nil {
		_ = tx.Rollback()
		return "", pushLog(err, helper.DBErr)
	}

	_ = tx.Commit()

	pipe := meta.MerchantRedis.TxPipeline()

	pipe.Unlink(ctx, m.Username)
	pipe.HMSet(ctx, m.Username, memberToMap(m))
	pipe.Persist(ctx, m.Username)

	_, _ = pipe.Exec(ctx)
	_ = pipe.Close()

	id, err := session.Set([]byte(m.Username), m.UID)
	if err != nil {
		return "", errors.New(helper.SessionErr)
	}

	log := map[string]string{
		"username":  username,
		"ip":        ip,
		"device":    fmt.Sprintf("%d", device),
		"device_no": deviceNo,
		"parents":   m.ParentName,
	}
	err = tdlog.WriteLog("member_login_log", log)
	if err != nil {
		fmt.Printf("member write member_login_log error : [%s]/n", err.Error())
	}

	l := MemberLoginLog{
		Username: userName,
		IPS:      ip,
		Device:   strconv.Itoa(device),
		DeviceNo: deviceNo,
		Date:     createdAt,
		Parents:  m.ParentName,
		Prefix:   meta.Prefix,
	}
	err = meta.Zlog.Post(esPrefixIndex("memberlogin"), l)
	if err != nil {
		fmt.Printf("zlog error : %v data : %#v\n", err, l)
	}

	encRes := [][]string{}

	encRes = append(encRes, []string{"phone", phone})
	err = grpc_t.Encrypt(m.UID, encRes)
	if err != nil {
		return "", errors.New(helper.GetRPCErr)
	}

	meta.MerchantRedis.Do(ctx, "CF.ADD", "phone_exist", phone).Err()
	return id, nil
}

func memberToMap(m Member) map[string]string {

	data := map[string]string{
		"uid":                  m.UID,
		"username":             m.Username,
		"password":             m.Password,
		"realname_hash":        m.RealnameHash,                       //真实姓名哈希
		"email_hash":           m.EmailHash,                          //邮件地址哈希
		"phone_hash":           m.PhoneHash,                          //电话号码哈希
		"zalo_hash":            m.ZaloHash,                           //电话号码哈希
		"prefix":               m.Prefix,                             //站点前缀
		"withdraw_pwd":         fmt.Sprintf("%d", m.WithdrawPwd),     //取款密码哈希
		"regip":                m.Regip,                              //注册IP
		"reg_device":           m.RegDevice,                          //注册设备号
		"reg_url":              m.RegUrl,                             //注册链接
		"created_at":           fmt.Sprintf("%d", m.CreatedAt),       //注册时间
		"last_login_ip":        m.LastLoginIp,                        //最后登陆ip
		"last_login_at":        fmt.Sprintf("%d", m.LastLoginAt),     //最后登陆时间
		"source_id":            fmt.Sprintf("%d", m.SourceId),        //注册来源 1 pc 2h5 3 app
		"first_deposit_at":     fmt.Sprintf("%d", m.FirstDepositAt),  //首充时间
		"first_deposit_amount": m.FirstDepositAmount,                 //首充金额
		"first_bet_at":         fmt.Sprintf("%d", m.FirstBetAt),      //首投时间
		"first_bet_amount":     m.FirstBetAmount,                     //首投金额
		"top_uid":              m.TopUid,                             //总代uid
		"top_name":             m.TopName,                            //总代代理
		"parent_uid":           m.ParentUid,                          //上级uid
		"parent_name":          m.ParentName,                         //上级代理
		"bankcard_total":       fmt.Sprintf("%d", m.BankcardTotal),   //用户绑定银行卡的数量
		"last_login_device":    m.LastLoginDevice,                    //最后登陆设备
		"last_login_source":    fmt.Sprintf("%d", m.LastLoginSource), //上次登录设备来源:1=pc,2=h5,3=ios,4=andriod
		"remarks":              m.Remarks,                            //备注
		"state":                fmt.Sprintf("%d", m.State),           //状态 1正常 2禁用
		"level":                fmt.Sprintf("%d", m.Level),           //等级
		"balance":              m.Balance,                            //余额
		"lock_amount":          m.LockAmount,                         //锁定金额
		"commission":           m.Commission,                         //佣金
		"group_name":           m.GroupName,                          //团队名称
		"agency_type":          fmt.Sprintf("%d", m.AgencyType),      //391团队代理 393普通代理
		"address":              m.Address,                            //收货地址
	}

	return data
}

func regLink(uid, linkID string, createdAt uint32) (Member, string, error) {

	m := Member{}
	var query string
	p := strings.Split(linkID, ":")
	if len(p) != 2 {
		return m, query, errors.New(helper.IDErr)
	}

	lkKey := "lk:" + p[0]
	lkRes, err := meta.MerchantRedis.Do(ctx, "JSON.GET", lkKey, ".$"+p[1]).Text()
	if err == nil {
		return m, query, pushLog(err, helper.RedisErr)
	}

	lk := Link_t{}
	err = helper.JsonUnmarshal([]byte(lkRes), &lk)
	if err != nil {
		return m, query, pushLog(err, helper.FormatErr)
	}

	fmt.Println("regLink :", lk)
	m, err = MemberFindByUid(lk.UID)
	if err != nil {
		return m, query, err
	}

	query = fmt.Sprintf("INSERT INTO `tbl_member_rebate_info` (`uid`, `zr`, `qp`, `ty`, `dj`, `dz`,`created_at`,`prefix`)VALUES(%s, '%s', '%s', '%s', '%s', '%s', '%d','%s');",
		uid, lk.ZR, lk.QP, lk.TY, lk.DJ, lk.DZ, createdAt, meta.Prefix)

	return m, query, nil
}

func regRoot(uid, topId string, createdAt uint32) (Member, string, error) {

	m := Member{}
	var query string
	rootRebate, err := RebateScale(topId)
	if err != nil {
		return m, query, pushLog(err, helper.DBErr)
	}

	m, err = MemberFindByUid(topId)
	if err != nil {
		return m, query, err
	}

	query = fmt.Sprintf("INSERT INTO `tbl_member_rebate_info` (`uid`, `zr`, `qp`, `ty`, `dj`, `dz`, `created_at`,`prefix`)VALUES(%s, '%s', '%s', '%s', '%s', '%s', '%d','%s');",
		uid, rootRebate.ZR, rootRebate.QP, rootRebate.TY, rootRebate.DJ, rootRebate.DZ, createdAt, meta.Prefix)

	return m, query, nil
}

func MemberVerify(id, str string) bool {

	val, err := meta.MerchantRedis.Get(ctx, id).Result()
	if err != nil || err == redis.Nil {
		return false
	}

	str = strings.ToLower(str)
	if val != str {
		return false
	}

	_, _ = meta.MerchantRedis.Unlink(ctx, id).Result()

	return true
}

func MemberInsert(parent Member, username, password, remark string, createdAt uint32, mr MemberRebate) error {

	userName := strings.ToLower(username)
	if MemberExist(userName) {
		return errors.New(helper.UsernameExist)
	}

	uid := helper.GenId()

	mr.UID = uid
	mr.ParentUID = parent.UID
	mr.CreatedAt = createdAt
	mr.Prefix = meta.Prefix
	m := Member{
		UID:                uid,
		Username:           userName,
		Password:           fmt.Sprintf("%d", MurmurHash(password, createdAt)),
		Prefix:             meta.Prefix,
		State:              1,
		CreatedAt:          createdAt,
		LastLoginIp:        "",
		LastLoginAt:        createdAt,
		LastLoginDevice:    "",
		LastLoginSource:    0,
		ParentUid:          parent.UID,
		ParentName:         parent.Username,
		TopUid:             parent.TopUid,
		TopName:            parent.TopName,
		FirstDepositAmount: "0.000",
		FirstBetAmount:     "0.000",
		Balance:            "0.000",
		LockAmount:         "0.000",
		Commission:         "0.000",
		Remarks:            remark,
	}

	tx, err := meta.MerchantDB.Begin() // 开启事务
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	query, _, _ := dialect.Insert("tbl_members").Rows(&m).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	query, _, _ = dialect.Insert("tbl_member_rebate_info").Rows(&mr).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	treeNode := MemberClosureInsert(uid, parent.UID)
	_, err = tx.Exec(treeNode)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	_, err = session.Set([]byte(m.Username), m.UID)
	if err != nil {
		return errors.New(helper.SessionErr)
	}

	return nil
}

func MemberCaptcha() ([]byte, string, error) {

	id := helper.GenId()
	text := meta.MerchantRedis.RPopLPush(ctx, "captcha", "captcha").Val()

	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()

	val := pipe.Get(ctx, text)
	pipe.SetNX(ctx, id, text, 120*time.Second)

	_, err := pipe.Exec(ctx)

	if err != nil {
		return nil, id, errors.New(helper.RedisErr)
	}

	img, err := val.Bytes()

	fmt.Println("pipe.Exec(ctx) 2 = ", err)
	if err != nil {
		return nil, id, errors.New(helper.RedisErr)
	}
	return img, id, nil
}

// 返回用户信息，会员端使用
func MemberInfo(fctx *fasthttp.RequestCtx) (MemberInfosData, error) {

	var err error
	res := MemberInfosData{}
	res.MemberInfos, err = memberInfoCache(fctx)
	if err != nil {
		return res, errors.New(helper.AccessTokenExpires)
	}

	encRes := []string{}
	if res.MemberInfos.RealnameHash != "0" {

		encRes = append(encRes, "realname")
	}

	if res.MemberInfos.PhoneHash != "0" {

		encRes = append(encRes, "phone")
	}

	if res.MemberInfos.EmailHash != "0" {

		encRes = append(encRes, "email")
	}

	if res.MemberInfos.ZaloHash != "0" {

		encRes = append(encRes, "zalo")
	}

	if len(encRes) > 0 {
		recs, err := grpc_t.Decrypt(res.MemberInfos.UID, true, encRes)
		if err != nil {

			//fmt.Println("MemberInfo res.MemberInfos.UID = ", res.MemberInfos.UID)
			//fmt.Println("MemberInfo grpc_t.Decrypt err = ", err.Error())
			return res, errors.New(helper.UpdateRPCErr)
		}

		res.Zalo = recs["zalo"]
		res.RealName = recs["realname"]
		res.Phone = recs["phone"]
		res.Email = recs["email"]
	}

	return res, nil
}

// 返回用户信息，会员端使用
func memberInfoCache(fCtx *fasthttp.RequestCtx) (MemberInfos, error) {

	m := MemberInfos{}
	name := string(fCtx.UserValue("token").([]byte))
	if name == "" {
		return m, errors.New(helper.UsernameErr)
	}

	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()

	exist := pipe.Exists(ctx, name)
	rs := pipe.HMGet(ctx, name, fieldsMemberInfo...)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return m, pushLog(err, helper.RedisErr)
	}

	num, err := exist.Result()
	if num == 0 {
		return m, errors.New(helper.UsernameErr)
	}

	if err = rs.Scan(&m); err != nil {
		return m, pushLog(err, helper.RedisErr)
	}

	return m, nil
}

// 通过用户名获取用户在redis中的数据
func MemberCache(fCtx *fasthttp.RequestCtx, name string) (Member, error) {

	m := Member{}
	if name == "" {
		name = string(fCtx.UserValue("token").([]byte))
		if name == "" {
			return m, errors.New(helper.UsernameErr)
		}
	}

	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()

	exist := pipe.Exists(ctx, name)
	rs := pipe.HMGet(ctx, name, fieldsMember...)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return m, pushLog(err, helper.RedisErr)
	}

	if exist.Val() == 0 {
		return m, errors.New(helper.UsernameErr)
	}

	if err = rs.Scan(&m); err != nil {
		return m, pushLog(rs.Err(), helper.RedisErr)
	}

	return m, nil
}

// 查询用户单条数据
func MemberFindOne(name string) (Member, error) {

	m := Member{}

	t := dialect.From("tbl_members")
	query, _, _ := t.Select(colsMember...).Where(g.Ex{"username": name, "prefix": meta.Prefix}).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&m, query)
	if err != nil && err != sql.ErrNoRows {
		return m, pushLog(err, helper.DBErr)
	}

	if err == sql.ErrNoRows {
		return m, errors.New(helper.UsernameErr)
	}

	return m, nil
}

func MemberFindByUid(uid string) (Member, error) {

	m := Member{}

	t := dialect.From("tbl_members")
	query, _, _ := t.Select(colsMember...).Where(g.Ex{"uid": uid}).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&m, query)
	if err != nil && err != sql.ErrNoRows {
		return m, pushLog(err, helper.DBErr)
	}

	if err == sql.ErrNoRows {
		return m, errors.New(helper.UsernameErr)
	}

	return m, nil
}

// 更新用户密码
func MemberPasswordUpdate(ty int, sid, code, old, password string, fctx *fasthttp.RequestCtx) error {

	mb, err := MemberCache(fctx, "")
	if err != nil {
		return err
	}

	// 邮箱 有绑定
	if ty == 1 && mb.PhoneHash != "0" {
		if !helper.CtypeDigit(sid) {
			return errors.New(helper.ParamErr)
		}

		if !helper.CtypeDigit(code) {
			return errors.New(helper.ParamErr)
		}

		ip := helper.FromRequest(fctx)

		recs, err := grpc_t.Decrypt(mb.UID, false, []string{"phone"})
		if err != nil {
			return errors.New(helper.GetRPCErr)
		}
		address := recs["phone"]

		err = phoneCmp(sid, code, ip, address)
		if err != nil {
			return err
		}
	}

	pwd := fmt.Sprintf("%d", MurmurHash(password, mb.CreatedAt))
	// 登录密码修改
	if ty == 1 {
		oldPassword := fmt.Sprintf("%d", MurmurHash(old, mb.CreatedAt))
		// 原始密码是否一致
		if oldPassword != mb.Password {
			return errors.New(helper.OldPasswordErr)
		}
	} else {
		if mb.WithdrawPwd != 0 {
			return errors.New(helper.WithdrawPwdExist)
		} else if pwd == mb.Password {
			return errors.New(helper.WPwdCanNotSameWithPwd)
		}
	}

	ex := g.Ex{
		"uid": mb.UID,
	}
	record := g.Record{}
	if ty == 1 {
		record["password"] = pwd
	} else {
		record["withdraw_pwd"] = pwd
	}
	// 更新会员信息
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

// 更新用户手机号
func MemberUpdatePhone(phone string, ctx *fasthttp.RequestCtx) error {

	if !validator.IsVietnamesePhone(phone) {
		return errors.New(helper.PhoneFMTErr)
	}

	phoneHash := fmt.Sprintf("%d", MurmurHash(phone, 0))
	ex := g.Ex{
		"phone_hash": phoneHash,
	}
	if MemberBindCheck(ex) {
		return errors.New(helper.PhoneExist)
	}

	mb, err := MemberCache(ctx, "")
	if err != nil {
		return err
	}

	//会员绑定手机号后，不允许更新手机号
	if mb.PhoneHash != "0" {
		return errors.New(helper.PhoneBindAlreadyErr)
	}

	encRes := [][]string{}
	encRes = append(encRes, []string{"phone", phone})
	err = grpc_t.Encrypt(mb.UID, encRes)
	if err != nil {
		return errors.New(helper.UpdateRPCErr)
	}

	record := g.Record{
		"phone_hash": phoneHash,
	}
	ex = g.Ex{
		"uid": mb.UID,
	}
	// 更新会员信息
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

// 更新用户zalo号
func MemberUpdateZalo(zalo string, ctx *fasthttp.RequestCtx) error {

	zaloHash := fmt.Sprintf("%d", MurmurHash(zalo, 0))
	ex := g.Ex{
		"zalo_hash": zaloHash,
	}
	if MemberBindCheck(ex) {
		return errors.New(helper.ZaloExist)
	}

	mb, err := MemberCache(ctx, "")
	if err != nil {
		return err
	}

	//会员绑定zalo号后，不允许更新zalo号
	if mb.ZaloHash != "0" {
		return errors.New(helper.ZaloBindAlreadyErr)
	}

	encRes := [][]string{}
	encRes = append(encRes, []string{"zalo", zalo})
	err = grpc_t.Encrypt(mb.UID, encRes)
	if err != nil {
		return errors.New(helper.UpdateRPCErr)
	}

	record := g.Record{
		"zalo_hash": zaloHash,
	}
	ex = g.Ex{
		"uid": mb.UID,
	}
	// 更新会员信息
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

// 更新用户信息
func MemberUpdateEmail(sid, code, email string, ctx *fasthttp.RequestCtx) error {

	emailHash := fmt.Sprintf("%d", MurmurHash(email, 0))
	ex := g.Ex{
		"email_hash": emailHash,
	}
	if MemberBindCheck(ex) {
		return errors.New(helper.EmailExist)
	}

	mb, err := MemberCache(ctx, "")
	if err != nil {
		return err
	}

	ip := helper.FromRequest(ctx)
	err = emailCmp(sid, code, ip, email)
	if err != nil {
		return err
	}

	encRes := [][]string{}
	encRes = append(encRes, []string{"email", email})
	err = grpc_t.Encrypt(mb.UID, encRes)
	if err != nil {
		return errors.New(helper.UpdateRPCErr)
	}

	record := g.Record{
		"email_hash": emailHash,
	}
	ex = g.Ex{
		"uid": mb.UID,
	}
	// 更新会员信息
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

// 用户信息更新
func MemberUpdateName(ctx *fasthttp.RequestCtx, realName, address string) error {

	mb, err := MemberCache(ctx, "")
	if err != nil {
		return err
	}

	record := g.Record{}

	//会员填写真实姓名后，不允许更新真实姓名
	if realName != "" { // 传入用户真实姓名  需要修改

		if mb.RealnameHash != "0" {
			return errors.New(helper.RealNameAlreadyBind)
		}

		if meta.Lang == "vn" && !validator.CheckStringVName(realName) {
			return errors.New(helper.RealNameFMTErr)
		}

		encRes := [][]string{}
		encRes = append(encRes, []string{"realname", realName})
		err = grpc_t.Encrypt(mb.UID, encRes)
		if err != nil {
			return errors.New(helper.UpdateRPCErr)
		}

		record["realname_hash"] = fmt.Sprintf("%d", MurmurHash(realName, 0))
	}

	if address != "" {
		record["address"] = address
	}

	if len(record) == 0 {
		return errors.New(helper.NoDataUpdate)
	}
	ex := g.Ex{
		"uid": mb.UID,
	}
	// 更新会员信息
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

// 检测会员账号是否已存在
func MemberExist(username string) bool {

	ex := meta.MerchantRedis.Exists(ctx, username).Val()

	if ex == 0 {
		return false
	}
	/*
		var uid uint64
		t := dialect.From("tbl_members")
		query, _, _ := t.Select("uid").Where(g.Ex{"username": username, "prefix": meta.Prefix}).ToSQL()
		err := meta.MerchantDB.Get(&uid, query)
		if err == sql.ErrNoRows {
			return false
		}
	*/
	return true
}

//会员忘记密码
func MemberForgetPwd(username, pwd, phone, ip, sid, code string) error {

	err := phoneCmp(sid, code, ip, phone)
	if err != nil {
		return err
	}

	mb, err := MemberCache(nil, username)
	if err != nil {
		return err
	}

	if len(mb.Username) == 0 {
		return errors.New(helper.UsernameErr)
	}

	phoneHash := fmt.Sprintf("%d", MurmurHash(phone, 0))
	if phoneHash != mb.PhoneHash {
		return errors.New(helper.UsernamePhoneMismatch)
	}

	record := g.Record{
		"password": fmt.Sprintf("%d", MurmurHash(pwd, mb.CreatedAt)),
	}
	ex := g.Ex{
		"uid": mb.UID,
	}
	// 更新会员信息
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

func MemberBindCheck(ex g.Ex) bool {

	var id string

	t := dialect.From("tbl_members")
	query, _, _ := t.Select("uid").Where(ex).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&id, query)
	if err == sql.ErrNoRows {
		return false
	}

	return true
}

func Platform() string {

	res, err := meta.MerchantRedis.Get(ctx, "plat").Result()
	if err == redis.Nil || err != nil {
		return "[]"
	}

	return res
}

func Nav() string {

	res, err := meta.MerchantRedis.Get(ctx, "nav").Result()
	if err == redis.Nil || err != nil {
		return "[]"
	}

	return res
}

func MemberList(ex g.Ex, username, startTime, endTime, sortField string, isAsc, page, pageSize int) (MemberListData, error) {

	data := MemberListData{}
	startAt, err := helper.TimeToLoc(startTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	endAt, err := helper.TimeToLoc(endTime, loc)
	if err != nil {
		return data, errors.New(helper.DateTimeErr)
	}

	if startAt >= endAt {
		return data, errors.New(helper.QueryTimeRangeErr)
	}

	data.S = pageSize

	if sortField != "" && username == "" {
		data.D, data.T, err = memberListSort(ex, sortField, startAt, endAt, isAsc, page, pageSize)
		if err != nil {
			return data, err
		}
	} else {
		data.D, data.T, err = memberList(ex, startAt, endAt, page, pageSize)
		if err != nil {
			return data, err
		}
	}

	if len(data.D) == 0 {
		return data, nil
	}

	var ids []string
	for _, v := range data.D {
		ids = append(ids, v.UID)
	}

	// 获取用户的反水比例
	rebate, err := MemberRebateSelect(ids)
	if err != nil {
		return data, err
	}

	for i, v := range data.D {
		if rb, ok := rebate[v.UID]; ok {
			data.D[i].DJ = rb.DJ
			data.D[i].TY = rb.TY
			data.D[i].ZR = rb.ZR
			data.D[i].QP = rb.QP
			data.D[i].DZ = rb.DZ
			data.D[i].CP = rb.CP
			data.D[i].FC = rb.FC
		}
	}

	return data, nil
}

func memberListSort(ex g.Ex, sortField string, startAt, endAt int64, isAsc, page, pageSize int) ([]MemberListCol, int, error) {

	var data []MemberListCol

	ex["report_time"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	ex["report_type"] = 2 //  1投注时间2结算时间3投注时间月报4结算时间月报

	number := 0
	if page == 1 {

		query, _, _ := dialect.From("tbl_report_agency").Select(g.COUNT(g.DISTINCT("uid"))).Where(ex).ToSQL()
		err := meta.ReportDB.Get(&number, query)
		if err != nil && err != sql.ErrNoRows {
			return data, 0, pushLog(err, helper.DBErr)
		}

		if number == 0 {
			return data, 0, nil
		}
	}

	orderField := g.C("report_time")
	if sortField != "" {
		orderField = g.C(sortField)
	}

	orderBy := orderField.Desc()
	if isAsc == 1 {
		orderBy = orderField.Asc()
	}

	offset := (page - 1) * pageSize
	and := g.And(ex, g.C("uid").Neq(g.C("parent_uid")))
	query, _, _ := dialect.From("tbl_report_agency").Select(
		"uid",
		"username",
		g.SUM("deposit_amount").As("deposit"),
		g.SUM("withdrawal_amount").As("withdraw"),
		g.SUM("dividend_amount").As("dividend"),
		g.SUM("rebate_amount").As("rebate"),
		g.SUM("company_net_amount").As("net_amount"),
	).GroupBy("uid").
		Where(and).
		Offset(uint(offset)).
		Limit(uint(pageSize)).
		Order(orderBy).
		ToSQL()
	err := meta.ReportDB.Select(&data, query)
	if err != nil {
		return data, number, pushLog(err, helper.DBErr)
	}

	return data, number, nil
}

func memberList(ex g.Ex, startAt, endAt int64, page, pageSize int) ([]MemberListCol, int, error) {

	var data []MemberListCol
	number := 0
	ex["prefix"] = meta.Prefix
	if page == 1 {
		query, _, _ := dialect.From("tbl_members").Select(g.COUNT(1)).Where(ex).ToSQL()
		err := meta.MerchantDB.Get(&number, query)
		if err != nil && err != sql.ErrNoRows {
			return data, number, pushLog(err, helper.DBErr)
		}

		if number == 0 {
			return data, number, nil
		}
	}

	var members []Member
	offset := (page - 1) * pageSize
	query, _, _ := dialect.From("tbl_members").Select("uid", "username").Where(ex).Offset(uint(offset)).
		Limit(uint(pageSize)).Order(g.L("created_at").Desc()).ToSQL()
	err := meta.MerchantDB.Select(&members, query)
	if err != nil {
		return data, number, pushLog(err, helper.DBErr)
	}

	// 补全数据
	var ids []string
	idMap := make(map[string]string, len(members))
	for _, member := range members {
		ids = append(ids, member.UID)
		idMap[member.UID] = member.Username
	}

	// 获取统计数据
	ex = g.Ex{
		"report_time": g.Op{"between": exp.NewRangeVal(startAt, endAt)},
		"uid":         ids,
		"report_type": 2, // 1投注时间2结算时间3投注时间月报4结算时间月报
	}
	and := g.And(ex, g.C("uid").Neq(g.C("parent_uid")))
	query, _, _ = dialect.From("tbl_report_agency").Where(and).
		Select(
			"uid",
			"username",
			g.SUM("deposit_amount").As("deposit"),
			g.SUM("withdrawal_amount").As("withdraw"),
			g.SUM("dividend_amount").As("dividend"),
			g.SUM("rebate_amount").As("rebate"),
			g.SUM("company_net_amount").As("net_amount"),
		).GroupBy("uid").
		ToSQL()
	err = meta.ReportDB.Select(&data, query)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err.Error())
		return data, number, pushLog(err, helper.DBErr)
	}

	if len(ids) == len(data) {
		return data, number, nil
	}

	// 可能有会员未生成报表数据 这时需要给未生成报表的会员 赋值默认返回值
	//否则会出现total和data length 不一致的问题
	for _, v := range data {
		if _, ok := idMap[v.UID]; ok {
			delete(idMap, v.UID)
		}
	}

	for id, username := range idMap {
		data = append(data, MemberListCol{UID: id, Username: username})
	}

	return data, number, nil
}

// MemberAgg 获取 活跃人数 直属下级人数 新增注册人数
func MemberAgg(username string) (MemberAggData, error) {

	data := MemberAggData{}

	reportTime := helper.MonthTST(0, loc)

	ex := g.Ex{
		"username":    username,
		"report_time": reportTime.Unix(),
		"report_type": 4,
	}

	query, _, _ := dialect.From("tbl_report_agency").Select("mem_count", "regist_count", "active_count").Where(ex).Limit(1).ToSQL()
	err := meta.ReportDB.Get(&data, query)
	if err != nil && err != sql.ErrNoRows {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func MemberUpdateInfo(user Member, password string, mr MemberRebate) error {

	tx, err := meta.MerchantDB.Begin() // 开启事务
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	subEx := g.Ex{
		"uid":    user.UID,
		"prefix": meta.Prefix,
	}

	if password != "" {
		recd := g.Record{
			"password": fmt.Sprintf("%d", MurmurHash(password, user.CreatedAt)),
		}
		query, _, _ := dialect.Update("tbl_members").Set(&recd).Where(subEx).ToSQL()
		_, err = tx.Exec(query)
		if err != nil {
			_ = tx.Rollback()
			return pushLog(err, helper.DBErr)
		}
	}

	recd := g.Record{
		"ty": mr.TY,
		"zr": mr.ZR,
		"qp": mr.QP,
		"dj": mr.DJ,
		"dz": mr.DZ,
		"cp": mr.CP,
		"fc": mr.FC,
	}
	query, _, _ := dialect.Update("tbl_member_rebate_info").Set(&recd).Where(subEx).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return pushLog(err, helper.DBErr)
	}

	err = tx.Commit()
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}
