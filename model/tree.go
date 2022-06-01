package model

import "fmt"

func TreeList(level string) (string, error) {

	key := fmt.Sprintf("%s:T:%s", meta.Prefix, level)
	data, err := meta.MerchantRedis.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return data, nil
}
