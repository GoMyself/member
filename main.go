package main

import (
	"fmt"
	"log"
	"member/contrib/apollo"
	"member/contrib/conn"
	"member/contrib/session"
	"member/middleware"
	"member/model"
	"member/router"
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

	mt.MerchantTD = conn.InitTD(cfg.Td.Addr, cfg.Td.MaxIdleConn, cfg.Td.MaxOpenConn)
	mt.MerchantDB = conn.InitDB(cfg.Db.Master.Addr, cfg.Db.Master.MaxIdleConn, cfg.Db.Master.MaxOpenConn)
	mt.ReportDB = conn.InitDB(cfg.Db.Report.Addr, cfg.Db.Report.MaxIdleConn, cfg.Db.Report.MaxOpenConn)
	mt.TiDB = conn.InitDB(cfg.Db.Tidb.Addr, cfg.Db.Tidb.MaxIdleConn, cfg.Db.Tidb.MaxOpenConn)

	mt.MerchantRedis = conn.InitRedisCluster(cfg.Redis.Addr, cfg.Redis.Password)

	bin := strings.Split(os.Args[0], "/")
	mt.Program = bin[len(bin)-1]
	mt.ES = conn.InitES(cfg.Es.Host, cfg.Es.Username, cfg.Es.Password)

	mt.CardValid = cfg.BankcardValidAPI
	model.Constructor(mt, cfg.RPC)
	session.New(mt.MerchantRedis, cfg.Prefix)

	defer func() {
		model.Close()
		mt = nil
	}()

	if os.Args[3] == "load" {
		model.MemberRebateUpdateALL()
		model.MemberFlushAll()
		fmt.Println("MemberRebateUpdateALL done")
		return
	}

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

	// 启动小飞机推送版本信息
	if !cfg.IsDev {
		telegramBotNotice(mt.Program, gitReversion, buildTime, buildGoVersion, "api", cfg.Prefix)
	}

	if err := srv.ListenAndServe(cfg.Port.Member); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}
