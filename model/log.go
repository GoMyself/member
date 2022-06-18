package model

import (
	"database/sql"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
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

	query.Filter(elastic.NewTermQuery("prefix", "w88"))
	fsc := elastic.NewFetchSourceContext(true).Include(fields...)
	offset := (page - 1) * pageSize
	esService := meta.ES.Search().FetchSourceContext(fsc).Query(query).From(offset).Size(pageSize).
		TrackTotalHits(true).Sort(sortField, false)
	if len(agg) > 0 {
		for k, v := range agg {
			esService = esService.Aggregation(k, v)
		}
	}

	resOrder, err := esService.Index(index).Do(ctx)
	if err != nil {
		fmt.Println(err)
		return 0, nil, nil, pushLog(err, helper.ESErr)
	}
	fmt.Printf("receive offset %+v total hits value:%d ,\n", offset, resOrder.Hits.TotalHits.Value)

	fmt.Printf("search receive agg %+v total value:%d ,\n", agg, resOrder.Hits.TotalHits.Value)

	if resOrder.Status != 0 || resOrder.Hits.TotalHits.Value <= int64(offset) {
		return resOrder.Hits.TotalHits.Value, nil, nil, nil
	}

	return resOrder.Hits.TotalHits.Value, resOrder.Hits.Hits, resOrder.Aggregations, nil
}

// 从report 报表补全数据
func MemberAggList(ex g.Ex, isAsc, page, pageSize int) ([]MemberListCol, int, error) {

	var data []MemberListCol

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

	orderBy := orderField.Desc()
	if isAsc == 1 {
		orderBy = orderField.Asc()
	}

	offset := (page - 1) * pageSize
	query, _, _ := dialect.From("tbl_report_agency").Select(
		"uid",
		"username",
		g.SUM("deposit_amount").As("deposit_amount"),
		g.SUM("withdrawal_amount").As("withdrawal_amount"),
		g.SUM("dividend_amount").As("dividend_amount"),
		g.SUM("rebate_amount").As("rebate_amount"),
		g.SUM("company_net_amount").As("company_net_amount"),
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
