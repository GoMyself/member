package model

import (
	"errors"
	"fmt"
	"log"
	"member/contrib/helper"
	"strconv"
	"time"

	g "github.com/doug-martin/goqu/v9"
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
	"1005": "TPB",
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
	"1023": "VRB",
	"1024": "UOB",
	"1025": "PUBLIC",
	"1026": "PG",
	"1027": "GPB",
	"1028": "IVB",
	"1029": "HLB",
	"1030": "KLB",
	"1031": "NAMA",
	"1032": "NCB",
	"1033": "OCB",
	"1034": "VietA",
	"1035": "BAO",
	"1036": "VietB",
	"1037": "CIMB",
	"1038": "LPB",
	"1039": "ABBANK",
	"1040": "CAP",
	"1041": "HD",
	"1042": "SCB",
	"1043": "DONGA",
	"1044": "SHB",
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
		BankcardTaskLogInsert(fctx, name, "NotSupport", bankCard, "1122 .RecordNotExistErr", 0)
		return helper.RecordNotExistErr
	}

	id, err := BankcardTaskCreate(ts, data)
	fmt.Printf("WARNING bankcark check:ts:%+v, card data:%+v", ts, data)
	if err != nil {
		// 插入记录 银行卡校验失败日志
		errmsg := fmt.Sprintf("BankcardTaskCreate BankcardValidErr %v %v %v ", helper.BankcardValidErr, id, err.Error())
		BankcardTaskLogInsert(fctx, name, bankcardCode[bankId], bankCard, errmsg[:59], 0)
		return helper.BankcardValidErr
	}

	for i := 0; i < 5; i++ {

		ts = fmt.Sprintf("%d", fctx.Time().In(loc).UnixMilli())
		valid, err := BankcardTaskQuery(ts, id)

		if err == nil {
			if valid {
				//插入记录 校验成功日志
				BankcardTaskLogInsert(fctx, name, bankcardCode[bankId], bankCard, fmt.Sprintf("try at %d Success %v", i, helper.Success), 1)
				return helper.Success
			} else {
				//插入记录 校验失败日志
				BankcardTaskLogInsert(fctx, name, bankcardCode[bankId], bankCard, fmt.Sprintf("try %d BankcardValidErr %v", i, helper.BankcardValidErr), 0)
				return helper.BankcardValidErr
			}
		}
		time.Sleep(2 * time.Second)
	}
	// 插入记录 银行卡校验失败日志
	BankcardTaskLogInsert(fctx, name, bankcardCode[bankId], bankCard, fmt.Sprintf("Final BankcardValidErr %v", helper.BankcardValidErr), 0)
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
	fmt.Printf("WARNING bank check:str %+v uri: %+v, sign: %+v,  body:%+v, status:%+v, err:%+v", str, uri, sign, body, statusCode, err)
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

// 新增会员 银行卡 校验记录
func BankcardTaskLogInsert(fctx *fasthttp.RequestCtx, realname, BankName, BankNo, msg string, Status int) error {

	var Username, Ip string
	mb, err := MemberCache(fctx, "")
	if err != nil {
		return err
	}
	Username = mb.Username

	if realname == "" {
		// 获取用户真实信息

		d, err := grpc_t.Decrypt(mb.UID, true, []string{"realname"})
		fmt.Println("grpc_t.Decrypt uids = ", mb.UID)
		fmt.Println("grpc_t.Decrypt d = ", d)
		fmt.Printf("and mb:%+v\n", mb)

		if err != nil {
			fmt.Println("grpc_t.Decrypt err = ", err)
			return errors.New(helper.GetRPCErr)
		}
		realname = d["realname"]
	}

	Ip = helper.FromRequest(fctx)
	ts := fctx.Time()
	/// 获取设备名称并保存
	DeviceName := string(fctx.Request.Header.Peek("d"))
	device_i, err := strconv.Atoi(DeviceName)
	if err != nil {
		return errors.New(helper.DeviceTypeErr)
	}

	if _, ok := Devices[device_i]; !ok {
		return errors.New(helper.DeviceTypeErr)
	}

	record := g.Record{
		"ts":          ts.In(loc).UnixMicro(),
		"username":    Username,
		"uid":         mb.UID,
		"realname":    realname,
		"bank_name":   BankName,
		"bankcard_no": BankNo,
		"ip":          Ip,
		"msg":         msg,
		"status":      Status,
		"device":      device_i,
		"created_at":  ts.In(loc).Unix(),
		"prefix":      meta.Prefix,
	}

	//写库
	err = BankcardTaskCheckInsertLog(record)
	if err != nil {
		return err
	}
	return nil
}

// 会员管理-会员银行卡 新增校验记录日志 写TD库
func BankcardTaskCheckInsertLog(record g.Record) error {

	query, param, errs := dialect.Insert("bankcard_log").Rows(record).ToSQL()
	if errs != nil {
		return pushLog(fmt.Errorf("errorr:%s, To insert Sql:, %s, param:[%s] ", errs.Error(), query, param), helper.DBErr)
	}
	_, err := meta.MerchantTD.Exec(query)
	if err != nil {
		fmt.Println("insert td = ", err.Error(), query)
		return pushLog(fmt.Errorf("%s,[%s]", err.Error(), query), helper.DBErr)
	}
	return nil
}
