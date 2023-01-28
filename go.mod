module github.com/minio/minio

go 1.18

require (
	cloud.google.com/go/storage v1.26.0
	github.com/Azure/azure-storage-blob-go v0.15.0
	github.com/Shopify/sarama v1.36.0
	github.com/alecthomas/participle v0.7.1
	github.com/bcicen/jstream v1.0.1
	github.com/beevik/ntp v0.3.0
	github.com/bits-and-blooms/bloom/v3 v3.3.1
	github.com/buger/jsonparser v1.1.1
	github.com/cespare/xxhash/v2 v2.1.2
	github.com/cheggaaa/pb v1.0.29
	github.com/coredns/coredns v1.9.4
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/cosnicolaou/pbzip2 v1.0.1
	github.com/dchest/siphash v1.2.3
	github.com/djherbis/atime v1.1.0
	github.com/dustin/go-humanize v1.0.0
	github.com/eclipse/paho.mqtt.golang v1.4.1
	github.com/elastic/go-elasticsearch/v7 v7.17.7
	github.com/fatih/color v1.13.0
	github.com/felixge/fgprof v0.9.3
	github.com/fraugster/parquet-go v0.12.0
	github.com/go-ldap/ldap/v3 v3.4.4
	github.com/go-openapi/loads v0.21.2
	github.com/go-sql-driver/mysql v1.6.0
	github.com/golang-jwt/jwt/v4 v4.4.2
	github.com/gomodule/redigo v1.8.9
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/inconshreveable/mousetrap v1.0.1
	github.com/json-iterator/go v1.1.12
	github.com/klauspost/compress v1.15.11
	github.com/klauspost/cpuid/v2 v2.1.2
	github.com/klauspost/pgzip v1.2.5
	github.com/klauspost/readahead v1.4.0
	github.com/klauspost/reedsolomon v1.11.0
	github.com/lib/pq v1.10.7
	github.com/lithammer/shortuuid/v4 v4.0.0
	github.com/miekg/dns v1.1.50
	github.com/minio/cli v1.24.0
	github.com/minio/console v0.21.3
	github.com/minio/csvparser v1.0.0
	github.com/minio/dperf v0.4.2
	github.com/minio/highwayhash v1.0.2
	github.com/minio/kes v0.22.0
	github.com/minio/madmin-go v1.7.5
	github.com/minio/minio-go/v7 v7.0.43
	github.com/minio/pkg v1.5.4
	github.com/minio/selfupdate v0.5.0
	github.com/minio/sha256-simd v1.0.0
	github.com/minio/simdjson-go v0.4.2
	github.com/minio/sio v0.3.0
	github.com/minio/xxml v0.0.3
	github.com/minio/zipindex v0.3.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nats-io/nats-server/v2 v2.7.4
	github.com/nats-io/nats.go v1.17.0
	github.com/nats-io/stan.go v0.10.3
	github.com/ncw/directio v1.0.5
	github.com/nsqio/go-nsq v1.1.0
	github.com/philhofer/fwd v1.1.2-0.20210722190033-5c56ac6d0bb9
	github.com/pierrec/lz4 v2.6.1+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.13.0
	github.com/prometheus/client_model v0.3.0
	github.com/prometheus/common v0.37.0
	github.com/prometheus/procfs v0.8.0
	github.com/rs/cors v1.8.2
	github.com/rs/dnscache v0.0.0-20211102005908-e0241e321417
	github.com/secure-io/sio-go v0.3.1
	github.com/shirou/gopsutil/v3 v3.22.9
	github.com/streadway/amqp v1.0.0
	github.com/tinylib/msgp v1.1.7-0.20220719154719-f3635b96e483
	github.com/valyala/bytebufferpool v1.0.0
	github.com/xdg/scram v1.0.5
	github.com/yargevad/filepathx v1.0.0
	github.com/zeebo/xxh3 v1.0.2
	go.etcd.io/etcd/api/v3 v3.5.5
	go.etcd.io/etcd/client/v3 v3.5.5
	go.uber.org/atomic v1.10.0
	go.uber.org/zap v1.23.0
	golang.org/x/crypto v0.2.0
	golang.org/x/oauth2 v0.1.0
	golang.org/x/sys v0.2.0
	golang.org/x/time v0.0.0-20220722155302-e5dcc9cfc0b9
	google.golang.org/api v0.98.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	cloud.google.com/go/compute v1.10.0 // indirect
	cloud.google.com/go/iam v0.4.0 // indirect
	github.com/bits-and-blooms/bitset v1.3.3 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/frankban/quicktest v1.14.0 // indirect
	github.com/google/pprof v0.0.0-20220829040838-70bd9ae97f40 // indirect
	github.com/minio/mc v0.0.0-20221103000258-583d449e38cd // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/nats-io/nats-streaming-server v0.24.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.16 // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/xdg/stringprep v1.0.3 // indirect
	go.mongodb.org/mongo-driver v1.10.3 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
)
