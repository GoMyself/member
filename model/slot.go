package model

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/valyala/fastjson"
	"member/contrib/helper"
	"strings"
)

func SlotList(pid string, flag, page, pageSize int) (string, error) {

	var (
		err   error
		total int64
	)

	buf := `{"d":[], "t":0}`
	key := fmt.Sprintf("%s:game:%s", meta.Prefix, pid)
	if flag == 1 {
		key = fmt.Sprintf("%s:game:n:%s", meta.Prefix, pid)
	} else if flag == 2 {
		key = fmt.Sprintf("%s:game:h:%s", meta.Prefix, pid)
	}
	if page == 1 {
		total, err = meta.MerchantRedis.LLen(ctx, key).Result()
		if err != nil {
			return buf, errors.New(helper.RedisErr)
		}

		if total < 1 {
			return buf, nil
		}
	}

	offset := (page - 1) * pageSize
	end := offset + pageSize - 1
	recs, err := meta.MerchantRedis.LRange(ctx, key, int64(offset), int64(end)).Result()
	if err != nil && err == redis.Nil {
		return buf, errors.New(helper.RedisErr)
	}

	arr := new(fastjson.Arena)
	aa := arr.NewArray()
	for i, val := range recs {
		aa.SetArrayItem(i, fastjson.MustParse(val))
	}

	buf = fmt.Sprintf(`{"d":%s, "t":%d}`, aa.String(), total)
	arr = nil
	return buf, nil
}

func SlotSearch(page, pageSize int, params map[string]string) (string, error) {

	//fields := []string{
	//	"id",
	//	"platform_id",
	//	"name",
	//	"en_name",
	//	"client_type",
	//	"game_type",
	//	"game_id",
	//	"img_phone",
	//	"img_pc",
	//	"img_cover",
	//	"is_hot",
	//	"sorting",
	//	"is_new",
	//	"vn_alias",
	//}
	//
	//fsc := elastic.NewFetchSourceContext(true).Include(fields...)
	//offset := (page - 1) * pageSize
	//
	//boolQuery := elastic.NewBoolQuery()
	//shouldQuery := elastic.NewBoolQuery()
	//keyword := fmt.Sprintf("*%s*", params["keyword"])
	//boolQuery.Should(
	//	elastic.NewWildcardQuery("vn_alias", keyword),
	//)
	//
	//boolQuery.Filter(shouldQuery)
	//
	//boolQuery.Must(elastic.NewTermQuery("on_line", "1"),
	//	elastic.NewTermQuery("platform_id", params["pid"]))
	//
	//if params["flag"] == "1" {
	//	boolQuery.Must(elastic.NewTermQuery("is_new", "1"))
	//} else if params["flag"] == "2" {
	//	boolQuery.Must(elastic.NewTermQuery("is_hot", "1"))
	//}
	//
	//esService := meta.ES.Search().
	//	Index(esPrefixIndex("tbl_game_lists")).
	//	FetchSourceContext(fsc).
	//	Query(boolQuery).
	//	From(offset).
	//	Size(pageSize).
	//	Sort("sorting", true).
	//	TrackTotalHits(true)
	//
	//resOrder, err := esService.Do(ctx)
	//if err != nil {
	//	return "", errors.New(helper.ESErr)
	//}
	//
	//arr := new(fastjson.Arena)
	//aa := arr.NewArray()
	//for i, v := range resOrder.Hits.Hits {
	//
	//	fmt.Println(i)
	//	buf := fastjson.MustParseBytes(v.Source)
	//	buf.Set("id", fastjson.MustParse(fmt.Sprintf(`"%s"`, v.Id)))
	//	aa.SetArrayItem(i, fastjson.MustParse(buf.String()))
	//	/*
	//		record := Game{}
	//		record.Id = v.Id
	//		_ = helper.JsonUnmarshal(v.Source, &record)
	//		data.D = append(data.D, record)
	//	*/
	//}
	//buf := fmt.Sprintf(`{"d":%s,"t":%d}`, aa.String(), resOrder.Hits.TotalHits.Value)
	//arr = nil
	//
	//return buf, nil

	var (
		err   error
		total int64
		num   int
	)

	buf := `{"d":[], "t":0}`
	key := fmt.Sprintf("%s:game:%s", meta.Prefix, params["pid"])
	if params["flag"] == "1" {
		key = fmt.Sprintf("%s:game:n:%s", meta.Prefix, params["pid"])
	} else if params["flag"] == "2" {
		key = fmt.Sprintf("%s:game:h:%s", meta.Prefix, params["pid"])
	}
	if page == 1 {
		total, err = meta.MerchantRedis.LLen(ctx, key).Result()
		if err != nil {
			return buf, errors.New(helper.RedisErr)
		}

		if total < 1 {
			return buf, nil
		}
	}

	offset := (page - 1) * pageSize
	end := offset + pageSize - 1
	recs, err := meta.MerchantRedis.LRange(ctx, key, int64(offset), int64(end)).Result()
	if err != nil && err == redis.Nil {
		return buf, errors.New(helper.RedisErr)
	}

	arr := new(fastjson.Arena)
	aa := arr.NewArray()

	for _, val := range recs {
		fmt.Println(val)
		if strings.Contains(val, params["keyword"]) {
			aa.SetArrayItem(num, fastjson.MustParse(val))
			num++
		}
	}

	buf = fmt.Sprintf(`{"d":%s, "t":%d}`, aa.String(), num)
	arr = nil
	return buf, nil
}

// ?????????????????????
func SlotBonusPool() (string, error) {

	bonus, err := meta.MerchantRedis.Get(ctx, "bonusPool").Result()
	if err != nil {
		return "64287325.41", nil
	}

	return bonus, nil
}
