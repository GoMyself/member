package model

import (
	"database/sql"
	"errors"
	"fmt"
	"member/contrib/helper"
	"member/contrib/session"
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

func MemberAmount(fctx *fasthttp.RequestCtx) (string, error) {

	username := string(fctx.UserValue("token").([]byte))
	if username == "" {
		return "", errors.New(helper.AccessTokenExpires)
	}

	mb := MBBalance{}
	t := dialect.From("tbl_members")
	query, _, _ := t.Select("balance", "lock_amount").Where(g.Ex{"username": username}).Limit(1).ToSQL()
	err := meta.MerchantDB.Get(&mb, query)
	if err != nil && err == sql.ErrNoRows {
		return "", pushLog(err, helper.DBErr)
	}

	if err == sql.ErrNoRows {
		return "", pushLog(err, helper.AccessTokenExpires)
	}

	data, err := helper.JsonMarshal(mb)
	if err != nil {
		return "", errors.New(helper.FormatErr)
	}

	return string(data), nil
}

func MemberLogin(fctx *fasthttp.RequestCtx, vid, code, username, password, ip, device, deviceNo string) (string, error) {

	ts := fctx.Time()
	key := fmt.Sprintf("%s:merchant:ip_blacklist", meta.Prefix)
	ipBlackListEx := meta.MerchantRedis.Do(ctx, "CF.EXISTS", key, ip).Val()
	if v, ok := ipBlackListEx.(int64); ok && v == 1 {
		return "", errors.New(helper.IpBanErr)
	}

	// web/h5不检查设备号黑名单
	if device != "24" && device != "25" {

		// 检查设备号黑名单
		key := fmt.Sprintf("%s:merchant:device_blacklist", meta.Prefix)
		deviceBlackListEx := meta.MerchantRedis.Do(ctx, "CF.EXISTS", key, deviceNo).Val()
		if v, ok := deviceBlackListEx.(int64); ok && v == 1 {
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
		"last_login_at":     ts.In(loc).Unix(),
		"last_login_device": deviceNo,
		"last_login_source": device,
	}

	t := dialect.Update("tbl_members")
	query, _, _ := t.Set(record).Where(ex).ToSQL()
	_, err = meta.MerchantDB.Exec(query)
	if err != nil {
		return "", err
	}

	data := g.Record{
		"prefix":      meta.Prefix,
		"username":    username,
		"ip":          ip,
		"device":      device,
		"device_no":   deviceNo,
		"top_uid":     mb.TopUid,
		"top_name":    mb.TopName,
		"parent_uid":  mb.ParentUid,
		"parent_name": mb.ParentName,
		"ts":          ts.In(loc).UnixMilli(),
		"create_at":   ts.In(loc).Unix(),
	}

	query, _, _ = dialect.Insert("member_login_log").Rows(&data).ToSQL()
	//fmt.Println(query)
	_, err = meta.MerchantTD.Exec(query)
	if err != nil {
		fmt.Println("insert SMS = ", err.Error())
	}

	MemberUpdateCache(mb.UID, "")
	/*
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
	*/
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

func MemberReg(device int, username, password, ip, deviceNo, regUrl, linkID, phone, ts string, createdAt uint32) (string, error) {

	// 检查ip黑名单

	topId := "4722355249852325"
	key := fmt.Sprintf("%s:merchant:ip_blacklist", meta.Prefix)
	ipBlacklistEx := meta.MerchantRedis.Do(ctx, "CF.EXISTS", key, ip).Val()
	if v, ok := ipBlacklistEx.(int64); ok && v == 1 {
		return "", errors.New(helper.IpBanErr)
	}

	phoneHash := fmt.Sprintf("%d", MurmurHash(phone, 0))
	phoneExist := meta.MerchantRedis.Do(ctx, "CF.EXISTS", "phoneExist", phone).Val()
	if v, ok := phoneExist.(int64); ok && v == 1 {
		return "", errors.New(helper.PhoneExist)
	}

	if !helper.CtypeDigit(ts) {
		return "", errors.New(helper.ParamErr)
	}

	// web/h5不检查设备号黑名单
	if _, ok := WebDevices[device]; !ok {

		// 检查设备号黑名单
		key := fmt.Sprintf("%s:merchant:device_blacklist", meta.Prefix)
		deviceBlacklistEx := meta.MerchantRedis.Do(ctx, "CF.EXISTS", key, deviceNo).Val()
		if v, ok := deviceBlacklistEx.(int64); ok && v == 1 {
			return "", errors.New(helper.DeviceBanErr)
		}

		maxKey := fmt.Sprintf("%s:risk:maxregnum", meta.Prefix)
		num, err := meta.MerchantRedis.Get(ctx, maxKey).Int()
		if err != nil && err != redis.Nil {
			return "", pushLog(err, helper.RedisErr)
		}

		if err == nil && num > 0 {
			ex := g.Ex{
				"prefix":     meta.Prefix,
				"reg_device": deviceNo,
			}
			regNum, err := MemberCount(ex)
			if err != nil {
				return "", err
			}

			if regNum > num {
				return "", errors.New(helper.RegLimitExceed)
			}
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
		UID:                 uid,
		Username:            userName,
		Password:            fmt.Sprintf("%d", MurmurHash(password, createdAt)),
		Birth:               "0",
		BirthHash:           "0",
		PhoneHash:           phoneHash,
		EmailHash:           "0",
		RealnameHash:        "0",
		ZaloHash:            "0",
		Prefix:              meta.Prefix,
		State:               1,
		Regip:               ip,
		RegDevice:           deviceNo,
		RegUrl:              regUrl,
		CreatedAt:           createdAt,
		LastLoginIp:         ip,
		LastLoginAt:         createdAt,
		SourceId:            sourceID,
		FirstDepositAmount:  "0.000",
		FirstBetAmount:      "0.000",
		SecondDepositAmount: "0.000",
		Balance:             "0.000",
		LockAmount:          "0.000",
		Commission:          "0.000",
		LastLoginDevice:     deviceNo,
		LastLoginSource:     lastLoginSource,
		Level:               1,
		//Tester: ,
		Avatar: "0",
	}

	tx, err := meta.MerchantDB.Begin() // 开启事务
	if err != nil {
		return "", pushLog(err, helper.DBErr)
	}

	fmt.Println("regLink id : ", linkID)
	parent := Member{}
	mr := MemberRebate{}
	var query string
	// 邀请链接注册，不成功注册在默认代理root下
	if linkID != "" {
		parent, mr, err = regLink(uid, linkID, createdAt)
		if err != nil {
			parent, mr, err = regRoot(uid, topId, createdAt)
			if err != nil {
				_ = tx.Rollback()
				return "", err
			}
		}
	} else {
		parent, mr, err = regRoot(uid, topId, createdAt)
		if err != nil {
			_ = tx.Rollback()
			return "", err
		}
	}

	m.Tester = parent.Tester
	m.ParentUid = parent.UID
	m.ParentName = parent.Username
	m.TopUid = parent.TopUid
	m.TopName = parent.TopName
	m.AgencyType = parent.AgencyType
	m.GroupName = parent.GroupName

	query, _, _ = dialect.Insert("tbl_member_rebate_info").Rows(&mr).ToSQL()
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

	id, err := session.Set([]byte(m.Username), m.UID)
	if err != nil {
		return "", errors.New(helper.SessionErr)
	}

	var encRes [][]string
	encRes = append(encRes, []string{"phone", phone})
	err = grpc_t.Encrypt(m.UID, encRes)
	if err != nil {
		return "", errors.New(helper.GetRPCErr)
	}

	_ = meta.MerchantRedis.Do(ctx, "CF.ADD", "phoneExist", phone).Err()
	_ = MemberRebateUpdateCache2(m.UID, mr)
	MemberUpdateCache(uid, "")

	fmt.Println("==== Reg TD Update ====")

	its, ie := strconv.ParseInt(ts, 10, 64)
	if ie != nil {
		fmt.Println("parse int err:", ie)
	}

	tdInsert("sms_log", g.Record{
		"ts":         its,
		"state":      "1",
		"updated_at": createdAt,
	})
	fmt.Println("==== Reg TD Update End ====")
	return id, nil
}

func regLink(uid, linkID string, createdAt uint32) (Member, MemberRebate, error) {

	m := Member{}
	mr := MemberRebate{}

	//var query string
	p := strings.Split(linkID, "|")
	if len(p) != 2 {
		return m, mr, errors.New(helper.IDErr)
	}

	lkKey := fmt.Sprintf("%s:lk:%s", meta.Prefix, p[0])
	lkRes, err := meta.MerchantRedis.Do(ctx, "JSON.GET", lkKey, ".$"+p[1]).Text()
	if err != nil {
		return m, mr, pushLog(err, helper.RedisErr)
	}

	lk := Link_t{}
	err = helper.JsonUnmarshal([]byte(lkRes), &lk)
	if err != nil {
		return m, mr, pushLog(err, helper.FormatErr)
	}

	//fmt.Println("regLink :", lk)
	m, err = MemberFindByUid(lk.UID)
	if err != nil {
		return m, mr, err
	}

	mr = MemberRebate{
		UID:              uid,
		ParentUID:        m.ParentUid,
		ZR:               lk.ZR,
		QP:               lk.QP,
		TY:               lk.TY,
		DJ:               lk.DJ,
		DZ:               lk.DZ,
		CP:               lk.CP,
		FC:               lk.FC,
		BY:               lk.BY,
		CGHighRebate:     lk.CGHighRebate,
		CGOfficialRebate: lk.CGOfficialRebate,
		CreatedAt:        createdAt,
		Prefix:           meta.Prefix,
	}

	/*
		query = fmt.Sprintf("INSERT INTO `tbl_member_rebate_info` (`uid`, `zr`, `qp`, `ty`, `dj`, `dz`, `cp`, `fc`, `by`, `created_at`, `prefix`, `cg_official_rebate`, `cg_high_rebate`)VALUES(%s, '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%d','%s','%s','%s');",
			uid, lk.ZR, lk.QP, lk.TY, lk.DJ, lk.DZ, lk.CP, lk.FC, lk.BY, createdAt, meta.Prefix, lk.CGOfficialRebate, lk.CGHighRebate)
	*/

	return m, mr, nil
}

func regRoot(uid, topId string, createdAt uint32) (Member, MemberRebate, error) {

	m := Member{}
	mr := MemberRebate{}

	rootRebate, err := MemberRebateGetCache(topId)
	if err != nil {
		return m, mr, pushLog(err, helper.DBErr)
	}

	m, err = MemberFindByUid(topId)
	if err != nil {
		return m, mr, err
	}

	mr = MemberRebate{
		UID:              uid,
		ParentUID:        m.ParentUid,
		ZR:               rootRebate.ZR,
		QP:               rootRebate.QP,
		TY:               rootRebate.TY,
		DJ:               rootRebate.DJ,
		DZ:               rootRebate.DZ,
		CP:               rootRebate.CP,
		FC:               rootRebate.FC,
		BY:               rootRebate.BY,
		CGHighRebate:     rootRebate.CGHighRebate,
		CGOfficialRebate: rootRebate.CGOfficialRebate,
		CreatedAt:        createdAt,
		Prefix:           meta.Prefix,
	}

	/*
		query = fmt.Sprintf("INSERT INTO `tbl_member_rebate_info` (`uid`, `zr`, `qp`, `ty`, `dj`, `dz`, `cp`, `fc`, `by`, `created_at`,`prefix`,`cg_official_rebate`,`cg_high_rebate`)VALUES(%s, '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%d', '%s', '%s', '%s');",
			uid, rootRebate.ZR, rootRebate.QP, rootRebate.TY, rootRebate.DJ, rootRebate.DZ, rootRebate.CP, rootRebate.FC, rootRebate.BY, createdAt, meta.Prefix, rootRebate.CGOfficialRebate, rootRebate.CGHighRebate)
	*/

	return m, mr, nil
}

func MemberVerify(id, code string) bool {

	code = fmt.Sprintf("%s:cap:code:%s", meta.Prefix, code)
	id = fmt.Sprintf("%s:cap:id:%s", meta.Prefix, id)
	cmd := meta.MerchantRedis.Get(ctx, id)
	fmt.Println(cmd.String())
	val, err := cmd.Result()
	if err != nil || err == redis.Nil {
		return false
	}

	fmt.Println(id, code, val)
	code = strings.ToLower(code)
	if val != code {
		return false
	}

	_, _ = meta.MerchantRedis.Unlink(ctx, id).Result()

	return true
}

func MemberCaptcha() ([]byte, string, error) {

	id := helper.GenId()
	key := fmt.Sprintf("%s:captcha", meta.Prefix)
	code := meta.MerchantRedis.RPopLPush(ctx, key, key).Val()

	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()

	code = fmt.Sprintf("%s:cap:code:%s", meta.Prefix, code)
	idKey := fmt.Sprintf("%s:cap:id:%s", meta.Prefix, id)
	val := pipe.Get(ctx, code)
	pipe.SetNX(ctx, idKey, code, 120*time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, id, pushLog(err, helper.RedisErr)
	}

	img, err := val.Bytes()
	if err != nil {
		return nil, id, pushLog(err, helper.RedisErr)
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

	key := meta.Prefix + ":member:" + name

	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()

	exist := pipe.Exists(ctx, key)
	rs := pipe.HMGet(ctx, key, "uid", "username", "password", "birth", "birth_hash", "realname_hash", "email_hash", "phone_hash", "zalo_hash", "prefix", "tester", "withdraw_pwd", "regip", "reg_device", "reg_url", "created_at", "last_login_ip", "last_login_at", "source_id", "first_deposit_at", "first_deposit_amount", "first_bet_at", "first_bet_amount", "", "", "top_uid", "top_name", "parent_uid", "parent_name", "bankcard_total", "last_login_device", "last_login_source", "remarks", "state", "level", "balance", "lock_amount", "commission", "group_name", "agency_type", "address", "avatar")

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

// 通过用户名获取用户在redis中的数据
func MemberCache(fctx *fasthttp.RequestCtx, name string) (Member, error) {

	m := Member{}
	if name == "" {
		name = string(fctx.UserValue("token").([]byte))
		if name == "" {
			return m, errors.New(helper.UsernameErr)
		}
	}

	field := []string{"uid", "username", "password", "birth", "birth_hash", "realname_hash", "email_hash", "phone_hash", "zalo_hash", "prefix", "tester", "withdraw_pwd", "regip", "reg_device", "reg_url", "created_at", "last_login_ip", "last_login_at", "source_id", "first_deposit_at", "first_deposit_amount", "first_bet_at", "first_bet_amount", "top_uid", "top_name", "parent_uid", "parent_name", "bankcard_total", "last_login_device", "last_login_source", "remarks", "state", "level", "balance", "lock_amount", "commission", "group_name", "agency_type", "address", "avatar"}
	key := meta.Prefix + ":member:" + name

	pipe := meta.MerchantRedis.TxPipeline()
	defer pipe.Close()

	exist := pipe.Exists(ctx, key)
	rs := pipe.HMGet(ctx, key, field...)

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
	/*
		recs := rs.Val()

		for k, v := range field {
			fmt.Println(v, " = ", recs[k])
		}
	*/
	/*
		t := dialect.From("tbl_members")
		query, _, _ := t.Select(colsMember...).Where(g.Ex{"username": name, "prefix": meta.Prefix}).Limit(1).ToSQL()
		err := meta.MerchantDB.Get(&m, query)
		if err != nil && err != sql.ErrNoRows {
			return m, pushLog(err, helper.DBErr)
		}

		if err == sql.ErrNoRows {
			return m, errors.New(helper.UsernameErr)
		}
	*/
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

func MemberUpdateCache(uid, username string) error {

	var (
		err error
		dst Member
	)

	if helper.CtypeDigit(uid) {
		dst, err = MemberFindByUid(uid)
		if err != nil {
			return err
		}
	} else {
		dst, err = MemberFindOne(username)
		if err != nil {
			return err
		}
	}

	key := meta.Prefix + ":member:" + dst.Username
	fields := []interface{}{"uid", dst.UID, "username", dst.Username, "password", dst.Password, "birth", dst.Birth, "birth_hash", dst.BirthHash, "realname_hash", dst.RealnameHash, "email_hash", dst.EmailHash, "phone_hash", dst.PhoneHash, "zalo_hash", dst.ZaloHash, "prefix", dst.Prefix, "tester", dst.Tester, "withdraw_pwd", dst.WithdrawPwd, "regip", dst.Regip, "reg_device", dst.RegDevice, "reg_url", dst.RegUrl, "created_at", dst.CreatedAt, "last_login_ip", dst.LastLoginIp, "last_login_at", dst.LastLoginAt, "source_id", dst.SourceId, "first_deposit_at", dst.FirstDepositAt, "first_deposit_amount", dst.FirstDepositAmount, "first_bet_at", dst.FirstBetAt, "first_bet_amount", dst.FirstBetAmount, "", dst.SecondDepositAt, "", dst.SecondDepositAmount, "top_uid", dst.TopUid, "top_name", dst.TopName, "parent_uid", dst.ParentUid, "parent_name", dst.ParentName, "bankcard_total", dst.BankcardTotal, "last_login_device", dst.LastLoginDevice, "last_login_source", dst.LastLoginSource, "remarks", dst.Remarks, "state", dst.State, "level", dst.Level, "balance", dst.Balance, "lock_amount", dst.LockAmount, "commission", dst.Commission, "group_name", dst.GroupName, "agency_type", dst.AgencyType, "address", dst.Address, "avatar", dst.Avatar}

	pipe := meta.MerchantRedis.TxPipeline()
	pipe.Del(ctx, key)
	pipe.HMSet(ctx, key, fields...)
	pipe.Persist(ctx, key)
	pipe.Exec(ctx)
	pipe.Close()
	return nil
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

// 检测会员账号是否已存在
func MemberExist(username string) bool {

	var uid uint64
	t := dialect.From("tbl_members")
	query, _, _ := t.Select("uid").Where(g.Ex{"username": username, "prefix": meta.Prefix}).ToSQL()
	err := meta.MerchantDB.Get(&uid, query)
	if err == sql.ErrNoRows {
		return false
	}
	return true
}

//会员忘记密码
func MemberForgetPwd(username, pwd, phone, ip, sid, code, ts string) error {

	mb, err := MemberCache(nil, username)
	if err != nil {
		return err
	}

	phoneHash := fmt.Sprintf("%d", MurmurHash(phone, 0))
	if phoneHash != mb.PhoneHash {
		return errors.New(helper.UsernamePhoneMismatch)
	}

	err = CheckSmsCaptcha(ip, sid, phone, code)
	if err != nil {
		return err
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

	fmt.Println("==== ForgotPWD TD Update ====")

	its, ie := strconv.ParseInt(ts, 10, 64)
	if ie != nil {
		fmt.Println("parse int err:", ie)
	}

	tdInsert("sms_log", g.Record{
		"ts":         its,
		"state":      "1",
		"updated_at": time.Now().Unix(),
	})
	fmt.Println("==== ForgotPWD TD Update End ====")

	MemberUpdateCache(mb.UID, "")
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

func MemberCount(ex g.Ex) (int, error) {

	var num int

	t := dialect.From("tbl_members")
	query, _, _ := t.Select(g.COUNT("uid")).Where(ex).Limit(1).ToSQL()
	fmt.Println(query)
	err := meta.MerchantDB.Get(&num, query)
	if err != nil && err != sql.ErrNoRows {
		return 0, pushLog(err, helper.DBErr)
	}

	return num, nil
}

func Platform() string {

	key := fmt.Sprintf("%s:plat", meta.Prefix)
	res, err := meta.MerchantRedis.Get(ctx, key).Result()
	if err == redis.Nil || err != nil {
		return "[]"
	}

	return res
}

func Nav() string {

	key := fmt.Sprintf("%s:nav", meta.Prefix)
	res, err := meta.MerchantRedis.Get(ctx, key).Result()
	if err == redis.Nil || err != nil {
		fmt.Println(key, err)
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
			data.D[i].BY = rb.BY
			data.D[i].CGHighRebate = rb.CGHighRebate
			data.D[i].CGOfficialRebate = rb.CGOfficialRebate
		}
	}

	return data, nil
}

func memberListSort(ex g.Ex, sortField string, startAt, endAt int64, isAsc, page, pageSize int) ([]MemberListCol, int, error) {

	var data []MemberListCol

	ex["report_time"] = g.Op{"between": exp.NewRangeVal(startAt, endAt)}
	ex["report_type"] = 2 //  1投注时间2结算时间3投注时间月报4结算时间月报
	ex["data_type"] = 1
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
	query, _, _ := dialect.From("tbl_report_agency").Select(
		"uid",
		"username",
		g.SUM("deposit_amount").As("deposit"),
		g.SUM("withdrawal_amount").As("withdraw"),
		g.SUM("dividend_amount").As("dividend"),
		g.SUM("rebate_amount").As("rebate"),
		g.SUM("company_net_amount").As("net_amount"),
	).GroupBy("uid").
		Where(ex).
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
		"data_type":   1,
	}
	query, _, _ = dialect.From("tbl_report_agency").Where(ex).
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

//  获取 活跃人数 直属下级人数 新增注册人数
func MemberAgg(username string) (MemberAggData, error) {

	data := MemberAggData{}

	reportTime := helper.MonthTST(0, loc)

	ex := g.Ex{
		"username":    username,
		"report_time": reportTime.Unix(),
		"report_type": 4,
		"data_type":   1,
	}

	query, _, _ := dialect.From("tbl_report_agency").Select("mem_count", "regist_count", "active_count").Where(ex).Limit(1).ToSQL()
	err := meta.ReportDB.Get(&data, query)
	if err != nil && err != sql.ErrNoRows {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func MemberUpdateInfo(user Member, password string, mr MemberRebateResult_t) error {

	key := fmt.Sprintf("%s:rebate:enablemod", meta.Prefix)
	if meta.MerchantRedis.Exists(ctx, key).Val() == 0 {
		return errors.New(helper.MemberRebateModDisable)
	}

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

		MemberUpdateCache(user.UID, "")
	}

	recd := g.Record{

		"by":                 mr.BY,
		"ty":                 mr.TY,
		"zr":                 mr.ZR,
		"qp":                 mr.QP,
		"dj":                 mr.DJ,
		"dz":                 mr.DZ,
		"cp":                 mr.CP,
		"fc":                 mr.FC,
		"cg_official_rebate": mr.CGOfficialRebate,
		"cg_high_rebate":     mr.CGHighRebate,
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

	MemberRebateUpdateCache1(user.UID, mr)
	return nil
}
