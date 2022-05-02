package model

import (
	"context"
	"fmt"
	"member2/contrib/helper"
	"member2/contrib/tdlog"
	"member2/contrib/tracerr"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/hprose/hprose-golang/v3/rpc/core"
	rpchttp "github.com/hprose/hprose-golang/v3/rpc/http"
	. "github.com/hprose/hprose-golang/v3/rpc/http/fasthttp"

	"errors"

	g "github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/elastic/v7"
	"github.com/spaolacci/murmur3"
	"github.com/valyala/gorpc"
)

type log_t struct {
	ID      string `json:"id" msg:"id"`
	Project string `json:"project" msg:"project"`
	Flags   string `json:"flags" msg:"flags"`
	Fn      string `json:"fn" msg:"fn"`
	File    string `json:"file" msg:"file"`
	Content string `json:"content" msg:"content"`
}

var grpc_t struct {
	View       func(uid, field string) ([]string, error)
	Encrypt    func(uid string, data [][]string) error
	Decrypt    func(uid string, hide bool, field []string) (map[string]string, error)
	DecryptAll func(uids []string, hide bool, field []string) (map[string]map[string]string, error)
}

type MetaTable struct {
	Zlog              *fluent.Fluent
	MerchantRedis     *redis.Client
	MerchantDB        *sqlx.DB
	ReportDB          *sqlx.DB
	Grpc              *gorpc.DispatcherClient
	ES                *elastic.Client
	AutoCommission    bool
	Email             Email
	Prefix            string
	EsPrefix          string
	PullPrefix        string
	Lang              string
	MinioUploadUrl    string
	MinioImagesBucket string
	MinioJsonBucket   string
	RpcConn           string
}

var (
	meta *MetaTable
	loc  *time.Location
	ctx  = context.Background()

	dialect          = g.Dialect("mysql")
	colsMember       = helper.EnumFields(Member{})
	colsMemberInfos  = helper.EnumFields(MemberInfos{})
	colsBankcard     = helper.EnumFields(BankCard{})
	colsMemberRebate = helper.EnumFields(MemberRebate{})
	fieldsMember     = helper.EnumRedisFields(Member{})
	fieldsMemberInfo = helper.EnumRedisFields(MemberInfos{})

	colsEsCommissionTransaction = []string{"id", "bill_no", "uid", "username", "cash_type", "amount", "before_amount", "after_amount", "created_at", "platform_id"}
)

type Email struct {
	Name     string
	Account  string
	Password string
}

func Constructor(mt *MetaTable, rpcconn string) {

	meta = mt
	meta.RpcConn = rpcconn
	if meta.Lang == "cn" {
		loc, _ = time.LoadLocation("Asia/Shanghai")
	} else if meta.Lang == "vn" || meta.Lang == "th" {
		loc, _ = time.LoadLocation("Asia/Bangkok")
	}

	rpchttp.RegisterHandler()
	RegisterTransport()

	client := core.NewClient(rpcconn)
	//client.Use(log.Plugin)

	client.UseService(&grpc_t)
}

func MurmurHash(str string, seed uint32) uint64 {

	h64 := murmur3.New64WithSeed(seed)
	h64.Write([]byte(str))
	v := h64.Sum64()
	h64.Reset()

	return v
}

func pushLog(err error, code string) error {

	err = tracerr.Wrap(err)
	fields := map[string]string{
		"filename": tracerr.SprintSource(err, 2, 2),
		"content":  err.Error(),
		"fn":       code,
		"id":       helper.GenId(),
		"project":  "Member",
	}
	l := log_t{
		ID:      helper.GenId(),
		Project: "member",
		Flags:   code,
		Fn:      "",
		File:    tracerr.SprintSource(err, 2, 2),
		Content: err.Error(),
	}
	err = tdlog.Info(fields)
	if err != nil {
		fmt.Printf("write td[%#v] err : %s", fields, err.Error())
	}

	_ = meta.Zlog.Post(esPrefixIndex("member_error"), l)

	switch code {
	case helper.DBErr, helper.RedisErr, helper.ESErr:
		code = helper.ServerErr
	}

	note := fmt.Sprintf("Hệ thống lỗi %s", fields["id"])
	return errors.New(note)
}

func esPrefixIndex(index string) string {
	return meta.EsPrefix + index
}

func pullPrefixIndex(index string) string {
	return meta.PullPrefix + index
}

func Close() {
	_ = meta.MerchantDB.Close()
	_ = meta.MerchantRedis.Close()
}
