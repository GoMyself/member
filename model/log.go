package model

import (
	"fmt"
	//"github.com/doug-martin/goqu/v9/exp"
	"github.com/olivere/elastic/v7"
	"member/contrib/helper"
)

// starc 会员列表查询执行
func EsMemberListSearch(index, sortField string,
	page, pageSize int,
	fields []string,
	query *elastic.BoolQuery,
	agg map[string]*elastic.SumAggregation) (int64, []*elastic.SearchHit, elastic.Aggregations, error) {

	fsc := elastic.NewFetchSourceContext(true).Include(fields...)
	offset := (page - 1) * pageSize
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

// starc 会员列表按序 查询执行
func EsMemberListSort(index, sortField string, ascending bool,
	page, pageSize int,
	fields []string,
	query *elastic.BoolQuery,
	agg map[string]*elastic.SumAggregation) (int64, []*elastic.SearchHit, elastic.Aggregations, error) {

	query.Filter(elastic.NewTermQuery("report_type", 2)) //  1投注时间2结算时间3投注时间月报4结算时间月报
	query.Filter(elastic.NewTermQuery("data_type", 1))
	fsc := elastic.NewFetchSourceContext(true).Include(fields...)
	offset := (page - 1) * pageSize

	esService := meta.ES.Search().FetchSourceContext(fsc).Query(query).From(offset).Size(pageSize).
		TrackTotalHits(true).
		Sort(sortField, ascending)

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