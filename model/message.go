package model

import (
	"database/sql"
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"member/contrib/helper"
	"time"
)

//MessageList  站内信列表
func MessageList(ty, page, pageSize int, isRead, username string) (MessageTDData, error) {

	data := MessageTDData{
		S: pageSize,
	}
	return data, nil

	ex := g.Ex{
		"prefix":    meta.Prefix,
		"username":  username,
		"is_delete": 0,
		"is_read":   isRead,
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
	fmt.Println(query)
	err := meta.MerchantTD.Select(&data.D, query)
	if err != nil {
		return data, pushLog(err, helper.DBErr)
	}

	return data, nil
}

// 紧急站内信
func MessageEmergency(username string) (MessageTD, error) {

	data := MessageTD{}
	return data, nil

	ex := g.Ex{
		"prefix":   meta.Prefix,
		"username": username,
		"is_top":   1,
		"is_read":  0,
	}
	query, _, _ := dialect.From("messages").Select(colsMessageTD...).Where(ex).Order(g.C("ts").Desc()).Limit(1).ToSQL()
	fmt.Println(query)
	err := meta.MerchantTD.Get(&data, query)
	if err != nil && err != sql.ErrNoRows {
		return data, pushLog(err, helper.DBErr)
	}

	if err == sql.ErrNoRows {
		return data, err
	}

	return data, nil
}

func MessageNum(username string) (int64, error) {

	return 0, nil

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

	return nil

	fmt.Println(ts)
	l := len(ts)
	if l < 26 {
		return errors.New(helper.DateTimeErr)
	}

	ts = ts[:l-6] + "+" + ts[l-5:]
	t, err := time.ParseInLocation("2006-01-02T15:04:05.999999+07:00", ts, loc)
	if err != nil {
		return pushLog(err, helper.DateTimeErr)
	}
	fmt.Println(t.Date())
	record := g.Record{
		"ts":      t.UnixMicro(),
		"is_read": 1,
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
func MessageDelete(username string, tss []string, flag int) error {

	return nil

	fmt.Println("MessageDelete", username, tss)
	if flag == 2 {
		var data []string
		ex := g.Ex{
			"prefix":   meta.Prefix,
			"is_read":  1,
			"username": username,
		}
		query, _, _ := dialect.From("messages").Select("ts").Where(ex).ToSQL()
		fmt.Println("MessageDelete", query)
		err := meta.MerchantTD.Select(&data, query)
		if err != nil {
			return pushLog(err, helper.DBErr)
		}

		fmt.Println("MessageDelete", data)
		return messageDelete(data)
	}

	fmt.Println("MessageDelete", tss)

	return messageDelete(tss)
}

func messageDelete(tss []string) error {

	if len(tss) == 0 {
		return nil
	}

	var records []g.Record
	for _, ts := range tss {
		fmt.Println("MessageDelete", ts)
		l := len(ts)
		if l < 26 {
			return errors.New(helper.DateTimeErr)
		}
		ts = ts[:l-6] + "+" + ts[l-5:]
		t, err := time.ParseInLocation("2006-01-02T15:04:05.999999+07:00", ts, loc)
		if err != nil {
			return pushLog(err, helper.DateTimeErr)
		}

		fmt.Println(t.Date())
		record := g.Record{
			"ts":        t.UnixMicro(),
			"is_delete": 1,
		}
		records = append(records, record)
	}
	query, _, _ := dialect.Insert("messages").Rows(records).ToSQL()
	fmt.Println("MessageDelete", query)
	_, err := meta.MerchantTD.Exec(query)
	if err != nil {
		return pushLog(err, helper.DBErr)
	}

	return nil
}
