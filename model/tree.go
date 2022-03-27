package model

func TreeList(level string) (string, error) {

	data, err := meta.MerchantRedis.Get(ctx, "T:"+level).Result()
	if err != nil {
		return "", err
	}

	return data, nil
}