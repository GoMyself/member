package model

import (
	"fmt"
	"github.com/olivere/elastic/v7"
	"member/contrib/helper"
)

var (
	transferFields    = []string{"id", "uid", "bill_no", "platform_id", "username", "transfer_type", "amount", "before_amount", "after_amount", "created_at", "state", "automatic", "confirm_name"}
	transactionFields = []string{"id", "bill_no", "uid", "username", "cash_type", "amount", "before_amount", "after_amount", "created_at"}
	gameRecordFields  = []string{"agency_uid", "agency_name", "agency_gid", "settle_time", "start_time", "resettle", "presettle", "ball_type", "odds", "handicap", "handicap_type", "game_name", "flow_quota", "is_use", "main_bill_no", "api_bill_no", "api_name", "updated_at", "created_at", "result", "prefix", "file_path", "copy_flag", "play_type", "flag", "valid_bet_amount", "bet_amount", "game_type", "bet_time", "net_amount", "uid", "name", "player_name", "api_type", "bill_no", "row_id", "id"}
	//rebateFields      = []string{"id", "agency_uid", "agency_name", "agency_gid", "level", "uid", "agency_type", "username", "rebate_at", "ration_at", "should_amount", "rebate_amount", "check_at", "state", "check_note", "ration_flag", "check_uid", "check_name", "create_at"}
	dividendFields = []string{"id", "uid", "top_uid", "top_name", "parent_uid", "parent_name", "wallet", "ty", "agency_type", "water_limit", "platform_id", "username", "amount", "hand_out_amount", "water_flow", "notify", "state", "hand_out_state", "remark", "review_remark", "apply_at", "apply_uid", "apply_name", "review_at", "review_uid", "review_name"}
	adjustFields   = []string{"id", "uid", "username", "top_uid", "top_name", "parent_uid", "parent_name", "amount", "adjust_type", "adjust_mode", "is_turnover", "turnover_multi", "pid", "apply_remark", "review_remark", "state", "hand_out_state", "images", "apply_at", "apply_uid", "apply_name", "review_at", "review_uid", "review_name"}
	depositFields  = []string{"id", "prefix", "oid", "channel_id", "finance_type", "uid", "top_uid", "top_name", "parent_uid", "parent_name", "username", "cid", "pid", "amount", "state", "automatic", "created_at", "created_uid", "created_name", "confirm_at", "confirm_uid", "confirm_name", "review_remark", "rate", "usdt_apply_amount", "protocol_type", "address", "hash_id", "usdt_final_amount", "flag", "bankcard_id", "manual_remark", "bank_no", "bank_code"}
	withdrawFields = []string{"id", "prefix", "bid", "flag", "finance_type", "oid", "uid", "top_uid", "top_name", "parent_uid", "parent_name", "username", "pid", "amount", "state", "automatic", "created_at", "confirm_at", "confirm_uid", "review_remark", "withdraw_at", "confirm_name", "withdraw_uid", "withdraw_name", "withdraw_remark", "bank_name", "card_name", "card_no", "receive_at"}
)

//ES查询转账记录
func esSearch(index string, sortFields map[string]bool, page, pageSize int, fields []string,
	param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (int64, []*elastic.SearchHit, elastic.Aggregations, error) {

	boolQuery := elastic.NewBoolQuery()
	terms := make([]elastic.Query, 0)
	filters := make([]elastic.Query, 0)

	if len(rangeParam) > 0 {
		for k, v := range rangeParam {
			if v == nil {
				continue
			}

			if len(v) == 2 {

				if v[0] == nil && v[1] == nil {
					continue
				}
				if val, ok := v[0].(string); ok {
					switch val {
					case "gt":
						rg := elastic.NewRangeQuery(k).Gt(v[1])
						filters = append(filters, rg)
					case "gte":
						rg := elastic.NewRangeQuery(k).Gte(v[1])
						filters = append(filters, rg)
					case "lt":
						rg := elastic.NewRangeQuery(k).Lt(v[1])
						filters = append(filters, rg)
					case "lte":
						rg := elastic.NewRangeQuery(k).Lte(v[1])
						filters = append(filters, rg)
					}
					continue
				}

				rg := elastic.NewRangeQuery(k).Gte(v[0]).Lte(v[1])
				if v[0] == nil {
					rg.IncludeLower(false)
				}

				if v[1] == nil {
					rg.IncludeUpper(false)
				}

				filters = append(filters, rg)
			}
		}
	}

	if len(param) > 0 {
		for k, v := range param {
			if v == nil {
				continue
			}

			if vv, ok := v.([]interface{}); ok {
				filters = append(filters, elastic.NewTermsQuery(k, vv...))
				continue
			}

			terms = append(terms, elastic.NewTermQuery(k, v))
		}
	}

	boolQuery.Filter(filters...)
	boolQuery.Must(terms...)
	fsc := elastic.NewFetchSourceContext(true)
	if len(fields) > 0 {
		fsc = fsc.Include(fields...)
	}

	//打印es查询json
	esService := meta.ES.Search().FetchSourceContext(fsc).Query(boolQuery).TrackTotalHits(true)
	offset := (page - 1) * pageSize
	if page > 0 && pageSize > 0 {
		esService = esService.From(offset).Size(pageSize)
	}

	if len(sortFields) > 0 {
		for k, v := range sortFields {
			esService = esService.Sort(k, v)
		}
	}
	// 聚合条件
	if len(aggField) > 0 {
		for k, v := range aggField {
			esService = esService.Aggregation(k, elastic.NewSumAggregation().Field(v))
		}
	}

	resOrder, err := esService.Index(index).Do(ctx)
	if err != nil {
		fmt.Println(err)
		return 0, nil, nil, pushLog(err, helper.ESErr)
	}

	if resOrder.Status != 0 || resOrder.Hits.TotalHits.Value <= int64(offset) {
		return resOrder.Hits.TotalHits.Value, nil, nil, nil
	}

	return resOrder.Hits.TotalHits.Value, resOrder.Hits.Hits, resOrder.Aggregations, nil
}

func esQuerySearch(index, sortField string, page, pageSize int,
	fields []string, boolQuery *elastic.BoolQuery, agg map[string]*elastic.SumAggregation) (int64, []*elastic.SearchHit, elastic.Aggregations, error) {

	fsc := elastic.NewFetchSourceContext(true).Include(fields...)
	offset := (page - 1) * pageSize
	//打印es查询json
	esService := meta.ES.Search().FetchSourceContext(fsc).Query(boolQuery).From(offset).Size(pageSize).TrackTotalHits(true).Sort(sortField, false)
	for k, v := range agg {
		esService = esService.Aggregation(k, v)
	}
	resOrder, err := esService.Index(index).Do(ctx)
	if err != nil {
		fmt.Println(err)
		return 0, nil, nil, pushLog(err, helper.ESErr)
	}

	if resOrder.Status != 0 || resOrder.Hits.TotalHits.Value <= int64(offset) {
		return resOrder.Hits.TotalHits.Value, nil, nil, nil
	}

	return resOrder.Hits.TotalHits.Value, resOrder.Hits.Hits, resOrder.Aggregations, nil
}

//ES查询转账记录
func esTransferQuery(index string, sortFields map[string]bool, page, pageSize int,
	param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (TransferData, error) {

	data := TransferData{Agg: map[string]string{}}
	total, esData, aggData, err := esSearch(index, sortFields, page, pageSize, transferFields, param, rangeParam, aggField)
	if err != nil {
		return data, err
	}

	for k, v := range aggField {
		amount, _ := aggData.Sum(k)
		if amount != nil {
			data.Agg[v] = fmt.Sprintf("%.4f", *amount.Value)
		}
	}

	data.T = total
	for _, v := range esData {

		trans := Transfer{}
		trans.ID = v.Id
		_ = helper.JsonUnmarshal(v.Source, &trans)
		data.D = append(data.D, trans)
	}

	return data, nil
}

func esGameRecordQuery(index string, sortFields map[string]bool, page, pageSize int,
	param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (GameRecordData, error) {

	data := GameRecordData{Agg: map[string]string{}}
	total, esData, aggData, err := esSearch(index, sortFields, page, pageSize, gameRecordFields, param, rangeParam, aggField)
	if err != nil {
		return data, err
	}

	for k, v := range aggField {
		amount, _ := aggData.Sum(k)
		if amount != nil {
			data.Agg[v] = fmt.Sprintf("%.4f", *amount.Value)
		}
	}

	data.T = total
	for _, v := range esData {
		record := GameRecord{}
		record.ID = v.Id
		_ = helper.JsonUnmarshal(v.Source, &record)
		record.ApiTypes = fmt.Sprintf("%d", record.ApiType)
		data.D = append(data.D, record)
	}

	return data, nil
}

//ES查询账变记录
func esTransactionQuery(index string, sortFields map[string]bool, page, pageSize int,
	param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (TransactionData, error) {

	data := TransactionData{Agg: map[string]string{}}
	total, esData, aggData, err := esSearch(index, sortFields, page, pageSize, transactionFields, param, rangeParam, aggField)
	if err != nil {
		return data, err
	}

	for k, v := range aggField {
		amount, _ := aggData.Sum(k)
		if amount != nil {
			data.Agg[v] = fmt.Sprintf("%.4f", *amount.Value)
		}
	}

	data.T = total
	for _, v := range esData {

		trans := Transaction{}
		trans.ID = v.Id
		_ = helper.JsonUnmarshal(v.Source, &trans)
		data.D = append(data.D, trans)
	}

	return data, nil
}

//ES查询取款记录
func esWithdrawQuery(index string, sortFields map[string]bool, page, pageSize int,
	param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (WithdrawData, error) {

	data := WithdrawData{Agg: map[string]string{}}
	total, esData, aggData, err := esSearch(index, sortFields, page, pageSize, withdrawFields, param, rangeParam, aggField)
	if err != nil {
		return data, err
	}

	for k, v := range aggField {
		amount, _ := aggData.Sum(k)
		if amount != nil {
			data.Agg[v] = fmt.Sprintf("%.4f", *amount.Value)
		}
	}

	data.T = total
	for _, v := range esData {
		deposit := Withdraw{}
		deposit.ID = v.Id
		_ = helper.JsonUnmarshal(v.Source, &deposit)
		data.D = append(data.D, deposit)
	}

	return data, nil
}

//ES查询存款记录
func esDepositQuery(index string, sortFields map[string]bool, page, pageSize int,
	param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (DepositData, error) {

	data := DepositData{Agg: map[string]string{}}
	total, esData, aggData, err := esSearch(index, sortFields, page, pageSize, depositFields, param, rangeParam, aggField)
	if err != nil {
		return data, err
	}

	for k, v := range aggField {
		amount, _ := aggData.Sum(k)
		if amount != nil {
			data.Agg[v] = fmt.Sprintf("%.4f", *amount.Value)
		}
	}

	data.T = total
	for _, v := range esData {

		deposit := Deposit{}
		deposit.ID = v.Id
		_ = helper.JsonUnmarshal(v.Source, &deposit)
		data.D = append(data.D, deposit)
	}

	return data, nil
}

//ES查询红利记录
func esDividendQuery(index string, sortFields map[string]bool, page, pageSize int,
	param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (DividendData, error) {

	data := DividendData{Agg: map[string]string{}}
	total, esData, aggData, err := esSearch(index, sortFields, page, pageSize, dividendFields, param, rangeParam, aggField)
	if err != nil {
		return data, err
	}

	for k, v := range aggField {
		amount, _ := aggData.Sum(k)
		if amount != nil {
			data.Agg[v] = fmt.Sprintf("%.4f", *amount.Value)
		}
	}

	data.T = total
	for _, v := range esData {

		dividend := Dividend{}
		dividend.ID = v.Id
		_ = helper.JsonUnmarshal(v.Source, &dividend)
		data.D = append(data.D, dividend)
	}

	return data, nil
}

//ES查询返水记录
//func esRebateQuery(index, sortField string, page, pageSize int,
//	param map[string]interface{}, rangeParam map[string][]interface{}, aggField map[string]string) (RebateData, error) {
//
//	data := RebateData{Agg: map[string]string{}}
//	total, esData, aggData, err := esSearch(index, sortField, page, pageSize, rebateFields, param, rangeParam, aggField)
//	if err != nil {
//		return data, err
//	}
//
//	for k, v := range aggField {
//		amount, _ := aggData.Sum(k)
//		if amount != nil {
//			data.Agg[v] = fmt.Sprintf("%.4f", *amount.Value)
//		}
//	}
//
//	data.T = total
//	for _, v := range esData {
//
//		rebate := Rebate{}
//		rebate.ID = v.Id
//		_ = helper.JsonUnmarshal(v.Source, &rebate)
//		data.D = append(data.D, rebate)
//	}
//
//	return data, nil
//}
