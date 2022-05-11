package middleware

import (
	"errors"
	"fmt"
	"member2/contrib/session"

	"github.com/valyala/fasthttp"
)

func CheckTokenMiddleware(ctx *fasthttp.RequestCtx) error {

	path := string(ctx.Path())
	fmt.Println("request path:" + path)
	allows := map[string]bool{
		"/member/tree":               true,
		"/member/ip":                 true,
		"/member/notices":            true,
		"/member/app/upgrade":        true,
		"/member/slot/search":        true,
		"/member/banner":             true,
		"/member/platform":           true,
		"/member/nav":                true,
		"/member/slot/list":          true,
		"/member/password/forget":    true,
		"/member/available":          true,
		"/member/captcha":            true,
		"/member/reg":                true,
		"/member/login":              true,
		"/member/version":            true,
		"/member/popularevents":      true,
		"/member/slot/bonus":         true,
		"/member/pprof/":             true,
		"/member/pprof/block":        true,
		"/member/pprof/allocs":       true,
		"/member/pprof/cmdline":      true,
		"/member/pprof/goroutine":    true,
		"/member/pprof/heap":         true,
		"/member/pprof/profile":      true,
		"/member/pprof/trace":        true,
		"/member/pprof/threadcreate": true,
		"/member/agency":             true,
	}

	if _, ok := allows[path]; ok {
		return nil
	}

	data, err := session.Get(ctx)
	if err != nil {
		if path == "/member/email" {
			ctx.SetUserValue("token", data)
			return nil
		}
		//fmt.Printf("%s get token from ctx failed:%s\n",path, err.Error())
		//fmt.Println("err = ", err)
		return errors.New(`{"status":false,"data":"token"}`)
	}

	// 退出登陆
	if path == "/member/logout" {
		session.Destroy(ctx)
		return errors.New(`{"status":true,"data":"success"}`)
	}

	//fmt.Println("b = ", string(data))
	ctx.SetUserValue("token", data)
	return nil
}
