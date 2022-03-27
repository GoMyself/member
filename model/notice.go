package model

import (
	"fmt"
	"github.com/go-redis/redis/v8"
)

func Notices() (string, error) {

	data, err := meta.MerchantRedis.Get(ctx, "notices").Result()
	if err == redis.Nil {
		return "{}", nil
	}

	return fmt.Sprintf(`{"d":%s}`, data), nil
}
