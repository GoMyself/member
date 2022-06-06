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
	fmt.Println("EsMemberListSearch from index:", index, "sortField:", sortField, "result:", resOrder.Hits.TotalHits.Value, "length hits:", len(resOrder.Hits.Hits), resOrder.Aggregations)
	return resOrder.Hits.TotalHits.Value, resOrder.Hits.Hits, resOrder.Aggregations, nil
}

// EsMemberListSort @Description: starc 会员列表查询执行
// * @Author: starc
// * @Date: 2022/6/4 12:38
// * @LastEditTime: 2022/6/7 19:00
// * @LastEditors: starc
func EsMemberListSort(index, sortField string,
	page, pageSize int,
	fields []string,
	query *elastic.BoolQuery,
	agg map[string]*elastic.SumAggregation,
	isAsc int) (int64, []*elastic.SearchHit, elastic.Aggregations, error) {

	//var data []MemberListCol
	query.Filter(elastic.NewTermQuery("report_type", 2)) //  1投注时间2结算时间3投注时间月报4结算时间月报
	query.Filter(elastic.NewTermQuery("data_type", 1))
	fmt.Println("Warning EsMemberListSort query: \n")
	fmt.Printf("query:%+v\n", query)
	fsc := elastic.NewFetchSourceContext(true).Include(fields...)
	offset := (page - 1) * pageSize
	//打印es查询json
	//升序否?
	orderAsc := false
	if isAsc == 1 {
		orderAsc = true
	}
	esService := meta.ES.Search().FetchSourceContext(fsc).Query(query).From(offset).Size(pageSize).
		TrackTotalHits(true).
		Sort(sortField, orderAsc)
	fmt.Println("Warning meta.ES Sort:")
	fmt.Printf("esService:%+v\n", esService)

	for k, v := range agg {
		esService = esService.Aggregation(k, v)
	}
	resOrder, err := esService.Index(index).Do(ctx)
	fmt.Println("Warning meta.ES Sort: \n", resOrder)

	if err != nil {
		fmt.Println(err)
		return 0, nil, nil, pushLog(err, helper.ESErr)
	}

	if resOrder.Status != 0 || resOrder.Hits.TotalHits.Value <= int64(offset) {
		return resOrder.Hits.TotalHits.Value, nil, nil, nil
	}
	fmt.Println("EsMemberListSort from index:", index, "sortField:", sortField, "result:", resOrder.Hits.TotalHits.Value, "length hits:", len(resOrder.Hits.Hits), resOrder.Aggregations)
	return resOrder.Hits.TotalHits.Value, resOrder.Hits.Hits, resOrder.Aggregations, nil

}
