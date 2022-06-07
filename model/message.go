package model

import (
	"database/sql"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"member/contrib/helper"
	"time"
)

//MessageList  站内信列表
func MessageList(ty, page, pageSize int, username string) (MessageTDData, error) {

	data := MessageTDData{
		S: pageSize,
	}
	ex := g.Ex{
		"prefix":    meta.Prefix,
		"username":  username,
		"is_delete": 0,
	}
	if ty != 0 {
		ex["ty"] = ty
	}
	t := dialect.From("messages")
	if page == 1 {
		query, _, _ := t.Select(g.COUNT("ts")).Where(ex).ToSQL()
		fmt.Println(query)
		err := meta.MerchantTD.Get(&data.T, query)
		if err != nil && err != sql.ErrNoRows {
			return data, pushLog(err, helper.DBErr)
		}

		if data.T == 0 {
			return data, nil
		}
	}

	offset := (page - 1) * pageSize
	query, _, _ := t.Select(colsMessageTD...).Where(ex).Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("ts").Desc()).ToSQL()
	if ty == 0 {
		//query, _, _ = t.Select(colsMessageTD...).Where(ex).Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("is_top").Desc(), g.C("send_at").Desc()).ToSQL()
	}
	fmt.Println(query)
	err := meta.MerchantTD.Select(&data.D, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

func MessageNum(username string) (int64, error) {

	var num int64
	ex := g.Ex{
		"prefix":    meta.Prefix,
		"username":  username,
		"is_read":   0,
		"is_delete": 0,
	}
	query, _, _ := dialect.From("messages").Select(g.COUNT("ts")).Where(ex).ToSQL()
	fmt.Println(query)
	err := meta.MerchantTD.Get(&num, query)
	if err != nil && err != sql.ErrNoRows {
		return num, pushLog(err, helper.DBErr)
	}

	return num, nil
}

//MessageRead  站内信已读
func MessageRead(ts string) error {

	t, err := time.ParseInLocation("2006-01-02T15:04:05.999 07:00", ts, loc)
	if err != nil {
		return pushLog(err, helper.DateTimeErr)
	}
	fmt.Println(t.Date())
	record := g.Record{
		"ts":        t.UnixMilli(),
		"is_delete": 1,
	}
	query, _, _ := dialect.Insert("messages").Rows(record).ToSQL()
	fmt.Println(query)
	_, err = meta.MerchantTD.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}

// 站内信删除已读
func MessageDelete(username string, ids []string, flag int) error {

	if flag == 2 {
		ex := g.Ex{
			"prefix":   meta.Prefix,
			"is_read":  1,
			"username": username,
		}
		query, _, _ := dialect.From("messages").Select("ts").Where(ex).ToSQL()
		fmt.Println(query)
		err := meta.MerchantTD.Select(&ids, query)
		if err != nil {
			return pushLog(err, helper.DBErr)
		}
	}
	var records []g.Record
	for _, v := range ids {
		t, _ := time.ParseInLocation(time.RFC3339, v, loc)
		fmt.Println(t.Date())
		record := g.Record{
			"ts":        t.UnixMilli(),
			"is_delete": 1,
		}
		records = append(records, record)
	}
	query, _, _ := dialect.Insert("messages").Rows(records).ToSQL()
	fmt.Println(query)
	_, err := meta.MerchantTD.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}
