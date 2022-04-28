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
		"prefix":           meta.Prefix,
		"ty":               ty,
		"username.keyword": username,
	}
	total, esData, _, err := esSearch(meta.EsPrefix+"messages", "send_at", page, pageSize, fields, param, map[string][]interface{}{}, map[string]string{})
	if err != nil {
		return `{"t":0,"d":[]}`, pushLog(err, helper.ESErr)
	}

	data := MessageEsData{}
	data.S = pageSize
	data.T = total
	for _, v := range esData {
		fmt.Println(v.Source)
		msg := MessageEs{}
		msg.ID = v.Id
		err = helper.JsonUnmarshal(v.Source, &msg)
		if err != nil {
			fmt.Printf("json : %s \n, error : %v \n", string(v.Source), err)
		}
		data.D = append(data.D, msg)
	}

	b, err := jettison.Marshal(data)
	if err != nil {
		return "", errors.New(helper.FormatErr)
	}

	fmt.Println("MessageList", username, total, esData, string(b))
	return string(b), nil
}

//MessageRead  站内信已读
func MessageRead(id, username string) error {

	handle := meta.ES.UpdateByQuery(meta.EsPrefix + "messages")

	handle.Query(elastic.NewTermQuery("id", id))
	handle.Query(elastic.NewTermQuery("username.keyword", username))
	handle.Query(elastic.NewTermQuery("prefix.keyword", meta.Prefix))

	_, err := handle.Script(elastic.NewScript("ctx._source['is_read']=1;")).ProceedOnVersionConflict().Do(ctx)
	if err != nil {
		return pushLog(err, helper.ESErr)
	}

	return nil
}

// 站内信删除已读
func MessageDelete(ids []interface{}, username string, flag int) error {

	query := elastic.NewBoolQuery().Filter(
		elastic.NewTermsQuery("id.keyword", ids...),
		elastic.NewTermQuery("username.keyword", username),
		elastic.NewTermQuery("prefix.keyword", meta.Prefix))
	if flag == 2 {
		query = elastic.NewBoolQuery().Filter(
			elastic.NewTermQuery("is_read", 1),
			elastic.NewTermQuery("username.keyword", username),
			elastic.NewTermQuery("prefix.keyword", meta.Prefix))
	}

	_, err := meta.ES.DeleteByQuery(meta.EsPrefix + "messages").
		Query(query).ProceedOnVersionConflict().Do(ctx)
	if err != nil {
		return pushLog(err, helper.ESErr)
	}

	return nil
}
