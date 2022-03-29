module prepare

go 1.16

replace github.com/coreos/bbolt v1.3.6 => go.etcd.io/bbolt v1.3.6

replace google.golang.org/grpc v1.44.0 => google.golang.org/grpc v1.26.0

replace github.com/BreezeTeam/scheduler/common => ./common

//replace go.mongodb.org/mongo-driver v1.8.3 => github.com/mongodb/mongo-go-driver v1.8.3

require (
	github.com/BreezeTeam/scheduler/common v0.0.0-incompatible
	//github.com/mongodb/mongo-go-driver v1.5.2
	github.com/coreos/bbolt v1.3.6 // indirect
	github.com/coreos/etcd v3.3.27+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/prometheus/client_golang v1.12.1 // indirect
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20220101234140-673ab2c3ae75 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.mongodb.org/mongo-driver v1.8.3
	go.uber.org/zap v1.20.0 // indirect
	golang.org/x/crypto v0.0.0-20220131195533-30dcbda58838 // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11 // indirect
	google.golang.org/genproto v0.0.0-20220204002441-d6cc3cc0770e // indirect
	google.golang.org/grpc v1.44.0 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
