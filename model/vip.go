package model

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"member2/contrib/helper"
)

func VipConfig() string {

	result, err := meta.MerchantRedis.Get(ctx, "plat").Result()
	if err != nil {
		return "{}"
	}

	return result
}

// 返回用户信息，会员端使用
func VipInfo(mb Member) (WaterFlow, error) {

	data := WaterFlow{
		TotalDeposit:        "0.0000", //累计存款
		TotalWaterFlow:      "0.0000", //累计流水
		ReturnDeposit:       "0.0000", //回归流水
		ReturnWaterFlow:     "0.0000", //回归存款
		RelegationWaterFlow: "0.0000", //保级流水
	}

	data.UID = mb.UID
	data.Username = mb.Username

	key := fmt.Sprintf("V:%s", mb.UID)
	rs := meta.MerchantRedis.HMGet(ctx, key, "uid", "username", "is_downgrade",
		"total_deposit", "total_water_flow", "return_deposit", "return_water_flow", "relegation_water_flow")

	if rs.Err() != nil {
		if rs.Err() == redis.Nil {
			return data, nil
		}

		return data, pushLog(rs.Err(), helper.RedisErr)
	}

	if err := rs.Scan(&data); err != nil {
		return data, pushLog(err, helper.RedisErr)
	}

	return data, nil
}
