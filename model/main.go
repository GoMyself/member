package model

import (
	"context"
	"fmt"
	"member/contrib/helper"
	"runtime"
	"strings"
	"time"

	"github.com/shopspring/decimal"

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
)

type bankcardValidAPI_t struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

var grpc_t struct {
	View        func(uid, field string) ([]string, error)
	Encrypt     func(uid string, data [][]string) error
	Decrypt     func(uid string, hide bool, field []string) (map[string]string, error)
	DecryptAll  func(uids []string, hide bool, field []string) (map[string]map[string]string, error)
	ShortURLGen func(rCtx context.Context, uri string) (string, error)
}

type MetaTable struct {
	MerchantRedis  *redis.ClusterClient
	MerchantDB     *sqlx.DB
	ReportDB       *sqlx.DB
	MerchantTD     *sqlx.DB
	CardValid      bankcardValidAPI_t
	ES             *elastic.Client
	AutoCommission bool
	Prefix         string
	EsPrefix       string
	PullPrefix     string
	Program        string
	Lang           string
}

var (
	meta             *MetaTable
	loc              *time.Location
	ctx              = context.Background()
	nine             = decimal.NewFromFloat(9.00)
	dialect          = g.Dialect("mysql")
	colsMember       = helper.EnumFields(Member{})
	colsBankcard     = helper.EnumFields(BankCard{})
	colsMemberRebate = helper.EnumFields(MemberRebate{})
	colsMessageTD    = helper.EnumFields(MessageTD{})

	colsEsCommissionTransaction = []string{"id", "bill_no", "uid", "username", "cash_type", "amount", "before_amount", "after_amount", "created_at"}
)

func Constructor(mt *MetaTable, rpcconn string) {

	meta = mt
	loc, _ = time.LoadLocation("Asia/Bangkok")

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

	_, file, line, _ := runtime.Caller(1)
	paths := strings.Split(file, "/")
	l := len(paths)
	if l > 2 {
		file = paths[l-2] + "/" + paths[l-1]
	}
	path := fmt.Sprintf("%s:%d", file, line)

	ts := time.Now()
	id := helper.GenId()

	fields := g.Record{
		"id":       id,
		"content":  err.Error(),
		"project":  meta.Program,
		"flags":    code,
		"filename": path,
		"ts":       ts.In(loc).UnixMicro(),
	}

	query, _, _ := dialect.Insert("goerror").Rows(fields).ToSQL()
	fmt.Println(query)
	_, err1 := meta.MerchantTD.Exec(query)
	if err1 != nil {
		fmt.Println("insert SMS = ", err1.Error(), query)
	}

	note := fmt.Sprintf("Hệ thống lỗi %s", id)
	return errors.New(note)
}

func tdInsert(tbl string, record g.Record) {

	query, _, _ := dialect.Insert(tbl).Rows(record).ToSQL()
	fmt.Println(query)
	_, err := meta.MerchantTD.Exec(query)
	if err != nil {
		fmt.Println("update td = ", err.Error(), record)
	}
}

func esPrefixIndex(index string) string {
	return meta.EsPrefix + index
}

func pullPrefixIndex(index string) string {
	return meta.PullPrefix + index
}

func Close() {
	meta.MerchantTD.Close()
	_ = meta.MerchantDB.Close()
	_ = meta.MerchantRedis.Close()
}
