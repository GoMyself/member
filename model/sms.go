package model

import (
	"errors"
	"github.com/go-redis/redis/v8"
	"member/contrib/helper"
	"strings"
)

func emailCmp(sid, code, ip, address string) error {

	if sid == "" && code == "" {
		return nil
	}

	if !strings.Contains(address, "@") {
		return errors.New(helper.EmailFMTErr)
	}

	key := address + ip + sid
	val, err := meta.MerchantRedis.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return errors.New(helper.ServerErr)
	}

	if err != nil && err == redis.Nil {
		return errors.New(helper.EmailVerificationErr)
	}

	if val != code {
		return errors.New(helper.EmailVerificationErr)
	}

	meta.MerchantRedis.Unlink(ctx, key)
	return nil
}
