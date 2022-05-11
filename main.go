package main

import (
	"fmt"
	"log"
	"member2/contrib/apollo"
	"member2/contrib/conn"
	"member2/contrib/session"
	"member2/contrib/tdlog"
	"member2/middleware"
	"member2/model"
	"member2/router"
	"os"
	"strings"

	"github.com/valyala/fasthttp"
	_ "go.uber.org/automaxprocs"
)

var (
	gitReversion   = ""
	buildTime      = ""
	buildGoVersion = ""
)

func main() {

	argc := len(os.Args)
	if argc != 4 {
		fmt.Printf("%s <etcds> <cfgPath> <web|load>\r\n", os.Args[0])
		return
	}

	cfg := conf{}

	endpoints := strings.Split(os.Args[1], ",")
	mt := new(model.MetaTable)
	apollo.New(endpoints)
	apollo.Parse(os.Args[2], &cfg)
	apollo.Close()

	mt.Lang = cfg.Lang
	mt.Prefix = cfg.Prefix
	mt.EsPrefix = cfg.EsPrefix
	mt.PullPrefix = cfg.PullPrefix
	mt.AutoCommission = cfg.AutoCommission
	mt.Zlog = conn.InitFluentd(cfg.Zlog.Host, cfg.Zlog.Port)
	mt.MerchantDB = conn.InitDB(cfg.Db.Master.Addr, cfg.Db.Master.MaxIdleConn, cfg.Db.Master.MaxOpenConn)
	mt.ReportDB = conn.InitDB(cfg.Db.Report.Addr, cfg.Db.Report.MaxIdleConn, cfg.Db.Report.MaxOpenConn)
	mt.MerchantRedis = conn.InitRedisSentinel(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.Sentinel, cfg.Redis.Db)
	//mt.MerchantRedisRead = conn.InitRedisSentinelRead(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.Sentinel, cfg.Redis.Db)
	mt.ES = conn.InitES(cfg.Es.Host, cfg.Es.Username, cfg.Es.Password)
	mt.Email.Name = cfg.Email.Name
	mt.Email.Account = cfg.Email.Account
	mt.Email.Password = cfg.Email.Password

	model.Constructor(mt, cfg.RPC)
	session.New(mt.MerchantRedis, cfg.Prefix)
	tdlog.New(cfg.Td.Servers, cfg.Td.Username, cfg.Td.Password)

	//id := helper.GenId()
	//fmt.Println(id)
	//fields := map[string]string{
	//	"filename": "test",
	//	"content":  "error",
	//	"fn":       "123",
	//	"id":       id,
	//	"project":  "Member",
	//}
	//
	//fmt.Println(tdlog.Info(fields))

	defer func() {
		model.Close()
		mt = nil
	}()

	b := router.BuildInfo{
		GitReversion:   gitReversion,
		BuildTime:      buildTime,
		BuildGoVersion: buildGoVersion,
	}
	app := router.SetupRouter(b)
	srv := &fasthttp.Server{
		Handler:            middleware.Use(app.Handler),
		ReadTimeout:        router.ApiTimeout,
		WriteTimeout:       router.ApiTimeout,
		Name:               "member2",
		MaxRequestBodySize: 51 * 1024 * 1024,
	}
	fmt.Printf("gitReversion = %s\r\nbuildGoVersion = %s\r\nbuildTime = %s\r\n", gitReversion, buildGoVersion, buildTime)
	fmt.Println("member2 running", cfg.Port.Member)
	if err := srv.ListenAndServe(cfg.Port.Member); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}
