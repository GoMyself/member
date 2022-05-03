package model

import (
	"fmt"
	"github.com/valyala/fastjson"
	"strconv"
)

func BannerImages(flags, device string) (string, error) {

	filterKey := ""

	if device == "24" || device == "25" {
		filterKey = "filter:1"
	} else {
		filterKey = "filter:2"
	}

	base := fastjson.MustParse(`{"filter":{},"banner":[]}`)
	bannerKey := fmt.Sprintf("G%s%s", flags, device)

	data, err := meta.MerchantRedis.MGet(ctx, bannerKey, filterKey).Result()
	if err != nil {
		return "{}", nil
	}

	if v, ok := data[1].(string); ok {
		base.Set("filter", fastjson.MustParse(v))
	}

	if v, ok := data[0].(string); ok {
		base.Set("banner", fastjson.MustParse(v))
	}

	return base.String(), nil
}

func BannerPopular() string {

	key := "popular_events"
	data, err := meta.MerchantRedis.Get(ctx, key).Result()
	if err != nil {
		return "{}"
	}
	// 返回三条
	var p fastjson.Parser
	b, err := p.Parse(data)
	if err != nil {
		return "{}"
	}

	rs := b.GetArray()
	n := len(rs)
	if n < 3 {
		return "{}"
	}

	for i := n - 1; i > 2; i-- {
		idx := strconv.Itoa(i)
		b.Del(idx)
	}

	return b.String()
}
