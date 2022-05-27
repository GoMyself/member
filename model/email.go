package model

import (
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"member2/contrib/helper"
	"net/smtp"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jordan-wright/email"
	"lukechampine.com/frand"
)

func phoneCmp(sid, code, ip, phone string) error {

	key := phone + ip + sid

	fmt.Println("phoneCmp key = ", key)

	val, err := meta.MerchantRedis.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return errors.New(helper.ServerErr)
	}

	fmt.Println("val = ", val)
	fmt.Println("code = ", code)

	if val != code {
		return errors.New(helper.PhoneVerificationErr)
	}

	//rc := g.Record{
	//	"state": "1",
	//}
	//ex := g.Ex{
	//	"phone":  phone,
	//	"state":  "0",
	//	"code":   code,
	//	"prefix": meta.Prefix,
	//}
	//query, _, _ := dialect.Update("sms_log").Set(rc).Where(ex).Limit(1).ToSQL()
	//fmt.Println(query)
	//_, _ = meta.MerchantTD.Exec(query)

	meta.MerchantRedis.Unlink(ctx, key).Err()
	return nil
}

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

func EmailSend(day, ip, username, address string, flag int) (string, error) {

	id := helper.GenId()
	mb, err := MemberCache(nil, username)
	if err != nil {
		return id, err
	}

	if len(mb.Username) == 0 {
		return id, errors.New(helper.UsernameErr)
	}

	if flag == EmailForgetPassword {
		emailHash := fmt.Sprintf("%d", MurmurHash(address, 0))
		if emailHash != mb.EmailHash {
			return id, errors.New(helper.UsernameEmailMismatch)
		}
	} else if flag == EmailModifyPassword {

		recs, err := grpc_t.Decrypt(mb.UID, false, []string{"email"})
		if err != nil {
			return "", errors.New(helper.GetRPCErr)
		}

		address = recs["email"]
	}

	from := fmt.Sprintf("%s Gaming <%sgm@gmail.com>", meta.Email.Name, meta.Email.Name)
	topic := fmt.Sprintf("【Khoa học công nghệ %s】", meta.Email.Name)
	code := fmt.Sprintf("%d", frand.Uint64n(899999)+100000)
	content := fmt.Sprintf("【Khoa học công nghệ %s】mã xác nhận Email của bạn: %s, vui lòng nhập trong vòng 10 phút. Nếu không phải bạn, vui lòng không tiết lộ.", meta.Email.Name, code)

	err = sendMail(from, address, topic, content)
	if err == nil {
		pipe := meta.MerchantRedis.TxPipeline()
		defer pipe.Close()

		key := fmt.Sprintf("%s%s", address, day)
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, 3600*time.Second)
		pipe.SetNX(ctx, address+ip+id, code, 10*60*time.Second)
		_, err := pipe.Exec(ctx)
		if err != nil {
			return id, errors.New(helper.RedisErr)
		}
	} else {
		err = errors.New("failed")
	}

	return id, err
}

func sendMail(from, address, topic, content string) error {

	e := email.NewEmail()
	e.From = from
	e.To = []string{address} //收件人
	e.Subject = topic
	e.Text = []byte(content)
	return e.Send("smtp.gmail.com:587", smtp.PlainAuth("", meta.Email.Account, meta.Email.Password, "smtp.gmail.com"))
}
