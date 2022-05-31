package model

import (
	"fmt"
	"github.com/valyala/fastjson"
)

func BannerImages(flags, device string) (string, error) {

	bannerKey := fmt.Sprintf("%s:banner:G%s%s", meta.Prefix, flags, device)
	data, err := meta.MerchantRedis.Get(ctx, bannerKey).Result()
	if err != nil {
		return "{}", nil
	}

	base := fastjson.MustParse(`{"banner":[]}`)
	base.Set("banner", fastjson.MustParse(data))

	return base.String(), nil
}
