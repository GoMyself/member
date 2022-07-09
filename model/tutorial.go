package model

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"member/contrib/helper"
)

func TutorialRead(uid string) error {

	key := fmt.Sprintf(`%s:tutorial:%s`, meta.Prefix, uid)
	z := redis.Z{
		Score:  float64(1),
		Member: uid,
	}
	err := meta.MerchantRedis.SAdd(ctx, key, &z).Err()
	if err != nil {
		return pushLog(err, helper.RedisErr)
	} else {
		return nil
	}
}

func TutorialState(uid string) (bool, error) {

	key := fmt.Sprintf(`%s:tutorial:%s`, meta.Prefix, uid)
	bools := meta.MerchantRedis.SMIsMember(ctx, key, uid).Val()

	if len(bools) > 0 {
		return bools[0], nil
	} else {
		return false, nil
	}

}
