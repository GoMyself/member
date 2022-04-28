package model

import (
	"errors"
	"fmt"
	"github.com/olivere/elastic/v7"
	"github.com/wI2L/jettison"
	"member2/contrib/helper"
)

//MessageList  站内信列表
func MessageList(ty, page, pageSize int, username string) (string, error) {

	fields := []string{"msg_id", "username", "title", "sub_title", "content", "is_top", "is_vip", "ty", "is_read", "send_name", "send_at", "prefix"}
	param := map[string]interface{}{
		"prefix":   meta.Prefix,
		"ty":       ty,
		"username": username,
	}
	total, esData, _, err := esSearch(meta.EsPrefix+"messages", "send_at", page, pageSize, fields, param, map[string][]interface{}{}, map[string]string{})
	if err != nil {
		return `{"t":0,"d":[]}`, pushLog(err, helper.ESErr)
	}

	data := MessageEsData{}
	data.S = pageSize
	data.T = total
	for _, v := range esData {
		msg := MessageEs{}
		msg.ID = v.Id
		err = helper.JsonUnmarshal(v.Source, &msg)
		if err != nil {
			fmt.Printf("json : %s \n, error : %v \n", string(v.Source), err)
			continue
		}
		data.D = append(data.D, msg)
	}

	b, err := jettison.Marshal(data)
	if err != nil {
		return "", errors.New(helper.FormatErr)
	}

	return string(b), nil
}

//MessageRead  站内信已读
func MessageRead(id, username string) error {

	boolQuery := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("id", id),
		elastic.NewTermQuery("username", username),
		elastic.NewTermQuery("prefix", meta.Prefix))

	_, err := meta.ES.UpdateByQuery(meta.EsPrefix + "messages").Query(boolQuery).
		Script(elastic.NewScript("ctx._source['is_read']=1;")).ProceedOnVersionConflict().Do(ctx)
	if err != nil {
		return pushLog(err, helper.ESErr)
	}

	return nil
}

// 站内信删除已读
func MessageDelete(ids []interface{}, username string, flag int) error {

	query := elastic.NewBoolQuery().Filter(
		elastic.NewTermsQuery("id", ids...),
		elastic.NewTermQuery("username", username),
		elastic.NewTermQuery("prefix", meta.Prefix))
	if flag == 2 {
		query = elastic.NewBoolQuery().Must(
			elastic.NewTermQuery("is_read", 1),
			elastic.NewTermQuery("username", username),
			elastic.NewTermQuery("prefix", meta.Prefix))
	}

	_, err := meta.ES.DeleteByQuery(meta.EsPrefix + "messages").Query(query).ProceedOnVersionConflict().Do(ctx)
	if err != nil {
		return pushLog(err, helper.ESErr)
	}

	return nil
}
