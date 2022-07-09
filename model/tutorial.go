package model

import (
	"fmt"
	"member/contrib/helper"
)

func TutorialRead(uid string) error {

	key := fmt.Sprintf(`%s:tutorial:%s`, meta.Prefix, uid)
	err := meta.MerchantRedis.SAdd(ctx, key, uid).Err()
	if err != nil {
		return pushLog(err, helper.RedisErr)
	} else {
		return nil
	}
}

func TutorialState(uid string) (bool, error) {

	key := fmt.Sprintf(`%s:tutorial:%s`, meta.Prefix, uid)
	if meta.MerchantRedis.SIsMember(ctx, key, uid).Val() {
		return true, nil
	} else {
		return false, nil
	}
}
