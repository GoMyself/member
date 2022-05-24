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
	"102": "VCB",
	"103": "SAC",
	"104": "ACB",
	"105": "HD",
	"106": "BIDV",
	"107": "SEA",
	"108": "Vietin",
	"109": "VPB",
	"110": "TCB",
	"111": "MBB",
	"112": "SHB",
	"113": "VIB",
	"114": "PV",
	"115": "AGR",
	"116": "SGB",
	"117": "DONGA",
	"118": "EIB",

	"123": "CIMB",
	"124": "TPB",
	"125": "LPB",
	"126": "ABBANK",
	"127": "BAC",
	"128": "CAP",
	"129": "MSB",
	"130": "KLB",
	"131": "NAMA",
	"132": "NCB",
	"133": "OCB",
	"134": "SGB",
	"135": "VietA",
	"136": "BAO",
	"137": "Vietb",
	"138": "PG",
	"139": "OCE",
	"140": "GPB",
	"141": "IVB",
	"142": "HSBC",

	"143": "SHI",
	"144": "HLB",
	"145": "VRB",
	"148": "WOO",
}

func BankcardCheck(fctx *fasthttp.RequestCtx, bankCard, bankId, name string) error {

	ts := fmt.Sprintf("%d", fctx.Time().In(loc).UnixMilli())

	data := bankcard_check_t{
		BankCard: bankCard,
		//BankCode: bankCode,
		Name: name,
	}

	if val, ok := bankcardCode[bankId]; ok {
		data.BankCode = val
	} else {
		return errors.New(helper.RecordNotExistErr)
	}

	id, err := BankcardTaskCreate(ts, data)
	if err != nil {
		return err
	}

	for i := 0; i < 5; i++ {

		ts = fmt.Sprintf("%d", fctx.Time().In(loc).UnixMilli())

		valid, err := BankcardTaskQuery(ts, id)
		if err == nil {
			if valid {
				return errors.New(helper.Success)
			} else {
				return errors.New(helper.Failure)
			}
		}

		time.Sleep(2 * time.Second)
	}

	return nil
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
