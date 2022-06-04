package model

import (
	"fmt"
	//"github.com/doug-martin/goqu/v9/exp"
	"github.com/olivere/elastic/v7"
	"member/contrib/helper"
)

func EsMemberListSearch(index, sortField string,
	page, pageSize int,
	fields []string,
	query *elastic.BoolQuery,
	agg map[string]*elastic.SumAggregation) (int64, []*elastic.SearchHit, elastic.Aggregations, error) {

	fsc := elastic.NewFetchSourceContext(true).Include(fields...)
	offset := (page - 1) * pageSize
	//打印es查询json
	esService := meta.ES.Search().FetchSourceContext(fsc).Query(query).From(offset).Size(pageSize).TrackTotalHits(true).Sort(sortField, false)
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

func EsMemberListSort(index, sortField string,
	page, pageSize int,
	fields []string,
	query *elastic.BoolQuery,
	agg map[string]*elastic.SumAggregation) (int64, []*elastic.SearchHit, elastic.Aggregations, error) {

	//var data []MemberListCol

	query.Filter(elastic.NewTermQuery("report_type", 2)) //  1投注时间2结算时间3投注时间月报4结算时间月报
	query.Filter(elastic.NewTermQuery("data_type", 1))

	fsc := elastic.NewFetchSourceContext(true).Include(fields...)
	offset := (page - 1) * pageSize
	//打印es查询json
	esService := meta.ES.Search().FetchSourceContext(fsc).Query(query).From(offset).Size(pageSize).TrackTotalHits(true).Sort(sortField, false)
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
	//
	//number := 0
	//if page == 1 {
	//
	//	query, _, _ := dialect.From("tbl_report_agency").Select(g.COUNT(g.DISTINCT("uid"))).Where(ex).ToSQL()
	//	err := meta.ReportDB.Get(&number, query)
	//	if err != nil && err != sql.ErrNoRows {
	//		return data, 0, pushLog(err, helper.DBErr)
	//	}
	//
	//	if number == 0 {
	//		return data, 0, nil
	//	}
	//}
	//
	//orderField := g.C("report_time")
	//if sortField != "" {
	//	orderField = g.C(sortField)
	//}
	//
	//orderBy := orderField.Desc()
	//if isAsc == 1 {
	//	orderBy = orderField.Asc()
	//}
	//
	//offset := (page - 1) * pageSize
	//query, _, _ := dialect.From("tbl_report_agency").Select(
	//	"uid",
	//	"username",
	//	g.SUM("deposit_amount").As("deposit"),
	//	g.SUM("withdrawal_amount").As("withdraw"),
	//	g.SUM("dividend_amount").As("dividend"),
	//	g.SUM("rebate_amount").As("rebate"),
	//	g.SUM("company_net_amount").As("net_amount"),
	//).GroupBy("uid").
	//	Where(ex).
	//	Offset(uint(offset)).
	//	Limit(uint(pageSize)).
	//	Order(orderBy).
	//	ToSQL()
	//err := meta.ReportDB.Select(&data, query)
	//if err != nil {
	//	return data, number, pushLog(err, helper.DBErr)
	//}
	//
	//return data, number, nil
}
