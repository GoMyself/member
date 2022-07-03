package main

type conf struct {
	Lang           string `json:"lang"`
	Prefix         string `json:"prefix"`
	EsPrefix       string `json:"es_prefix"`
	PullPrefix     string `json:"pull_prefix"`
	IsDev          bool   `json:"is_dev"`
	AutoCommission bool   `json:"auto_commission"`
	Sock5          string `json:"sock5"`
	RPC            string `json:"rpc"`
	Fcallback      string `json:"fcallback"`
	AutoPayLimit   string `json:"autoPayLimit"`
	Nats           struct {
		Servers  []string `json:"servers"`
		Username string   `json:"username"`
		Password string   `json:"password"`
	} `json:"nats"`
	Beanstalkd struct {
		Addr    string `json:"addr"`
		MaxIdle int    `json:"maxIdle"`
		MaxCap  int    `json:"maxCap"`
	} `json:"beanstalkd"`
	Db struct {
		Master struct {
			Addr        string `json:"addr"`
			MaxIdleConn int    `json:"max_idle_conn"`
			MaxOpenConn int    `json:"max_open_conn"`
		} `json:"master"`
		Report struct {
			Addr        string `json:"addr"`
			MaxIdleConn int    `json:"max_idle_conn"`
			MaxOpenConn int    `json:"max_open_conn"`
		} `json:"report"`
		Bet struct {
			Addr        string `json:"addr"`
			MaxIdleConn int    `json:"max_idle_conn"`
			MaxOpenConn int    `json:"max_open_conn"`
		} `json:"bet"`
		Tidb struct {
			Addr        string `json:"addr"`
			MaxIdleConn int    `json:"max_idle_conn"`
			MaxOpenConn int    `json:"max_open_conn"`
		} `json:"tidb"`
	} `json:"db"`
	Td struct {
		Addr        string `json:"addr"`
		MaxIdleConn int    `json:"max_idle_conn"`
		MaxOpenConn int    `json:"max_open_conn"`
	} `json:"td"`
	BankcardValidAPI struct {
		URL string `json:"url"`
		Key string `json:"key"`
	} `json:"bankcard_valid_api"`
	Redis struct {
		Addr     []string `json:"addr"`
		Password string   `json:"password"`
	} `json:"redis"`
	Es struct {
		Host     []string `json:"host"`
		Username string   `json:"username"`
		Password string   `json:"password"`
	} `json:"es"`
	Email struct {
		Name     string `json:"name"`
		Account  string `json:"account"`
		Password string `json:"password"`
	} `json:"email"`
	Port struct {
		Game     string `json:"game"`
		Member   string `json:"member"`
		Promo    string `json:"promo"`
		Merchant string `json:"merchant"`
		Finance  string `json:"finance"`
	} `json:"port"`
}
