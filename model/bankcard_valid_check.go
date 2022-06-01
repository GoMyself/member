package model

import (
	"errors"
	"fmt"
	"log"
	"member2/contrib/helper"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

type bankcard_check_t struct {
	BankCard string `json:"bankCard"`
	Name     string `json:"name"`
	BankCode string `json:"bankCode"`
	Sign     string `json:"sign"`
}

var bankcardCode = map[string]string{
	"1006": "AGR",
	"1007": "OCB",
	"1008": "BIDV",
	"1009": "Vietin",
	"1010": "VCB",
	"1011": "VPB",
	"1012": "MBB",
	"1013": "TCB",
	"1014": "SAC",
	"1015": "VIB",
	"1016": "MSB",
	"1017": "EIB",
	"1018": "BAC",
	"1019": "SGB",
	"1020": "SHI",
	"1021": "HSBC",
	"1022": "WOO",
	"1023": "VRBANK",
	"1024": "UOB",
	"1025": "pbvn",
	"1026": "PG",
	"1027": "GPBANK",
	"1028": "IVB",
	"1029": "HLBVN",
	"1030": "KLB",
	"1031": "NAMA",
	"1032": "NCB",
	"1033": "OCB",
	"1034": "VietA",
	"1035": "BAOVIET",
	"1036": "VietB",
	"1037": "CIMB",
	"1038": "LPB",
	"1039": "ABBANK",
	"1040": "CAP",
	"1041": "HDBANK",
	"1042": "SCB",
	"1043": "DONGA",
	"1044": "SHBank",
	"1045": "SEA",
	"1046": "PV",
	"1048": "VDB",
	"1049": "SCVN",
	"1050": "CBB",
	"1051": "ACB",
}

func BankcardCheck(fctx *fasthttp.RequestCtx, bankCard, bankId, name string) string {

	ts := fmt.Sprintf("%d", fctx.Time().In(loc).UnixMilli())

	data := bankcard_check_t{
		BankCard: bankCard,
		//BankCode: bankCode,
		Name: name,
	}

	if val, ok := bankcardCode[bankId]; ok {
		data.BankCode = val
	} else {
		// 插入记录 银行卡校验失败日志
		MemberCardLogInsert(fctx, name, bankCard, 0)
		return helper.RecordNotExistErr
	}

	id, err := BankcardTaskCreate(ts, data)
	if err != nil {
		// 插入记录 银行卡校验失败日志
		MemberCardLogInsert(fctx, name, bankCard, 0)
		return helper.BankcardValidErr
	}

	for i := 0; i < 5; i++ {
		ts = fmt.Sprintf("%d", fctx.Time().In(loc).UnixMilli())
		valid, err := BankcardTaskQuery(ts, id)
		if err == nil {
			if valid {
				//插入记录 校验成功日志
				MemberCardLogInsert(fctx, name, bankCard, 1)
				return helper.Success
			} else {
				//插入记录 校验失败日志
				MemberCardLogInsert(fctx, name, bankCard, 0)
				return helper.BankcardValidErr
			}
		}
		time.Sleep(2 * time.Second)
	}
	// 插入记录 银行卡校验失败日志
	MemberCardLogInsert(fctx, name, bankCard, 0)
	return helper.BankcardValidErr
}

func BankcardTaskQuery(ts, id string) (bool, error) {

	headers := map[string]string{
		"Timestamp":    ts,
		"Nonce":        helper.GenId(),
		"Content-Type": "application/json",
	}

	str := fmt.Sprintf("orderNo=%s&timestamp=%s&appsecret=%s", id, ts, meta.CardValid.Key)
	uri := fmt.Sprintf("%s/bank/result/query", meta.CardValid.URL)

	sign := helper.GetMD5Hash(helper.GetMD5Hash(helper.GetMD5Hash(str)))

	b := fmt.Sprintf("{\"orderNo\":\"%s\", \"sign\":\"%s\"}", id, sign)

	body, statusCode, err := helper.HttpDoTimeout([]byte(b), "POST", uri, headers, 5*time.Second)
	if err != nil {
		return false, pushLog(err, helper.GetRPCErr)
	}
	if statusCode != 200 {
		return false, pushLog(fmt.Errorf("BankcardTaskQuery %d", statusCode), helper.GetRPCErr)
	}

	fmt.Println("BankcardTaskCreate = ", string(body))

	value := fastjson.MustParseBytes(body)
	//msg := string(value.GetStringBytes("msg"))
	code := string(value.GetStringBytes("code"))

	if code == "0000" {
		data := value.GetBool("data")
		return data, nil
	}

	return false, errors.New(helper.BankcardValidErr)
}

func BankcardTaskCreate(ts string, res bankcard_check_t) (string, error) {

	headers := map[string]string{
		"Timestamp":    ts,
		"Nonce":        helper.GenId(),
		"Content-Type": "application/json",
	}

	str := fmt.Sprintf("bankCode=%s&bankCard=%s&name=%s&timestamp=%s&appsecret=%s", res.BankCode, res.BankCard, res.Name, ts, meta.CardValid.Key)
	uri := fmt.Sprintf("%s/bank/check/create", meta.CardValid.URL)

	res.Sign = helper.GetMD5Hash(helper.GetMD5Hash(helper.GetMD5Hash(str)))

	b, err := helper.JsonMarshal(res)
	if err != nil {
		log.Fatal(err)
	}

	body, statusCode, err := helper.HttpDoTimeout(b, "POST", uri, headers, 5*time.Second)
	if err != nil {
		return "", pushLog(err, helper.GetRPCErr)
	}
	if statusCode != 200 {
		return "", pushLog(fmt.Errorf("BankcardTaskQuery %d", statusCode), helper.GetRPCErr)
	}

	fmt.Println("BankcardTaskCreate = ", string(body))

	value := fastjson.MustParseBytes(body)
	data := string(value.GetStringBytes("data"))
	code := string(value.GetStringBytes("code"))

	if code == "0000" {
		return data, nil
	}

	return "", err
}

/**
 * @Description: MemberCardList // 新增会员 银行卡 记录
 * @Author: starc
 * @Date: 2022/5/31 16:38
 * @LastEditTime: 2022/6/1 20:00
 * @LastEditors: starc
 */
func MemberCardLogInsert(ctx *fasthttp.RequestCtx, BankName, BankNo string, Status int) error {

	var Username, RealName, Ip string

	mb, err := MemberCache(ctx, "")
	if err != nil {
		return err
	}
	Username = mb.Username

	// 获取用户真实信息
	d, err := grpc_t.Decrypt(mb.UID, true, []string{"realname"})

	fmt.Println("grpc_t.Decrypt uids = ", mb.UID)
	fmt.Println("grpc_t.Decrypt d = ", d)

	if err != nil {
		fmt.Println("grpc_t.Decrypt err = ", err)
		return errors.New(helper.GetRPCErr)
	}

	RealName = d["realname"]
	Ip = helper.FromRequest(ctx)
	ts := ctx.Time().In(loc).UnixMilli()
	err2 := MemberCardInsert(Username, RealName, BankName, BankNo, Ip, Status, ts)
	if err2 != nil {
		helper.Print(ctx, false, err2.Error())
		return err2
	}
	helper.Print(ctx, true, true)
	return nil
}
