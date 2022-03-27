package model

import (
	"bitbucket.org/nwf2013/schema"
	"fmt"
	"github.com/valyala/fastjson"
)

type rpcResult struct {
	Err string `json:"err"`
	Res string `json:"res"`
}

/*
https://github.com/francoispqt/gojay
*/

func rpcSet(data []schema.Enc_t, isUpdate bool) ([]string, error) {

	method := "Encrypt"
	recs     := []string{}
	var p fastjson.Parser

	if isUpdate {
		method = "Update"
	}

	fmt.Println(method, " = ", data)

	res, err := meta.Grpc.Call(method, data)
	if err != nil {
		return recs, err
	}


	v, err := p.ParseBytes(res.([]byte))
	if err != nil {
		return recs, err
	}
	value, err := v.Array()
	if err != nil {
		return recs, err
	}

	for _, val := range value {
		recs = append(recs, val.String())
	}
	return  recs, err
}

func rpcInsert(data []schema.Enc_t) ([]string, error) {

	return rpcSet(data, false)
}

func rpcUpdate(data []schema.Enc_t) ([]string, error) {

	return rpcSet(data, true)
}

func rpcGet(data []schema.Dec_t) ([]rpcResult, error) {

	var p fastjson.Parser
	results := []rpcResult{}

	res, err := meta.Grpc.Call("Decrypt", data)
	if err != nil {
		return results, err
	}

	vv, err := p.ParseBytes(res.([]byte))
	if err != nil {
		return results, err
	}
	value, err := vv.Array()
	if err != nil {
		return results, err
	}
	for _, val := range value {
		r := rpcResult{
			Err : string(val.GetStringBytes("err")),
			Res : string(val.GetStringBytes("res")),
		}
		results = append(results, r)
	}

	return results, nil
}