package model

import "fmt"

func UpgradeInfo(device string) (string, error) {
	key := fmt.Sprintf("%s:upgrade:%s", meta.Prefix, device)
	return meta.MerchantRedis.Get(ctx, key).Result()
}
