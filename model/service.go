package model

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	g "github.com/doug-martin/goqu/v9"
)

type BuildInfo struct {
	id             int64
	Name           string
	GitReversion   string
	BuildTime      string
	BuildGoVersion string
	Flag           int
	IP             string
	Hostname       string
}

func NewService(gitReversion, buildTime, buildGoVersion string, flag int) BuildInfo {

	ts := time.Now().UnixMicro()
	b := BuildInfo{
		id:             ts,
		Flag:           flag,
		Name:           meta.Program,
		GitReversion:   gitReversion,
		BuildTime:      buildTime,
		BuildGoVersion: buildGoVersion,
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
		return b
	}

	b.Hostname = hostname

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
		return b
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				b.IP = ipnet.IP.String()
				break
			}
		}
	}

	return b
}

func (s BuildInfo) keepAlive() error {

	now := time.Now()
	recs := g.Record{
		"ts":             s.id,
		"name":           s.Name,
		"flag":           s.Flag,
		"ip":             s.IP,
		"hostname":       s.Hostname,
		"buildTime":      s.BuildTime,
		"gitReversion":   s.GitReversion,
		"buildGoVersion": s.BuildGoVersion,
		"created_at":     now.Unix(),
		"prefix":         meta.Prefix,
	}

	query, _, _ := dialect.Insert("services").Rows(recs).ToSQL()
	_, err := meta.MerchantTD.Exec(query)
	if err != nil {
		fmt.Println("insert service failed query ", query)
		fmt.Println("insert service failed error ", err.Error())
		return err
	}

	return nil
}

func (s BuildInfo) Start() error {

	s.keepAlive()

	for {

		time.Sleep(10 * time.Second)
		s.keepAlive()
	}
}
