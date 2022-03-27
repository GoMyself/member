module member2

go 1.17

replace github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

require (
	bitbucket.org/nwf2013/schema v0.0.0-20210601104355-f8329e4edddd
	github.com/buaazp/fasthttprouter v0.1.1
	github.com/doug-martin/goqu/v9 v9.18.0
	github.com/fluent/fluent-logger-golang v1.9.0
	github.com/go-redis/redis/v8 v8.11.4
	github.com/jmoiron/sqlx v1.3.4
	github.com/jordan-wright/email v4.0.1-0.20210109023952-943e75fe5223+incompatible
	github.com/olivere/elastic/v7 v7.0.31
	github.com/shopspring/decimal v1.3.1
	github.com/spaolacci/murmur3 v1.1.0
	github.com/valyala/fasthttp v1.32.0
	github.com/valyala/fastjson v1.6.3
	github.com/valyala/gorpc v0.0.0-20160519171614-908281bef774
	github.com/wI2L/jettison v0.7.3
	go.uber.org/automaxprocs v1.4.0
	lukechampine.com/frand v1.4.2
)

require (
	github.com/aead/chacha20 v0.0.0-20180709150244-8b13a72661da // indirect
	github.com/andybalholm/brotli v1.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/klauspost/compress v1.14.4 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/kr/text v0.1.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/minio/sha256-simd v0.1.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/xid v1.2.1 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce // indirect
	golang.org/x/net v0.0.0-20211209124913-491a49abca63 // indirect
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9 // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/genproto v0.0.0-20200825200019-8632dd797987 // indirect
	google.golang.org/grpc v1.33.2 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/ini.v1 v1.57.0 // indirect
)

require (
	github.com/beanstalkd/go-beanstalk v0.1.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869
	github.com/coreos/bbolt v0.0.0-00010101000000-000000000000 // indirect
	github.com/coreos/etcd v3.3.27+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/go-sql-driver/mysql v1.6.0
	github.com/goccy/go-json v0.9.5
	github.com/google/btree v1.0.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/ipipdotnet/ipdb-go v1.3.1
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/minio/md5-simd v1.1.2
	github.com/minio/minio-go/v7 v7.0.23
	github.com/modern-go/reflect2 v1.0.2
	github.com/nats-io/nats-server/v2 v2.7.3 // indirect
	github.com/nats-io/nats.go v1.13.1-0.20220121202836-972a071d373d
	github.com/panjf2000/ants/v2 v2.4.8
	github.com/pelletier/go-toml v1.9.4
	github.com/prometheus/client_golang v1.12.1 // indirect
	github.com/silenceper/pool v1.0.0
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/tinylib/msgp v1.1.6
	github.com/tmc/grpc-websocket-proxy v0.0.0-20220101234140-673ab2c3ae75 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	github.com/xxtea/xxtea-go v0.0.0-20170828040851-35c4b17eecf6
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/time v0.0.0-20220224211638-0e9765cccd65 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
