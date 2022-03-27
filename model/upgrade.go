package model

func UpgradeInfo(device string) (string, error) {

	key := "upgrade:" + device
	return meta.MerchantRedis.Get(ctx, key).Result()
}
