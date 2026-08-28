module github.com/cps-enterprise/dcs/master-agent

go 1.23

require (
	github.com/cps-enterprise/dcs/proto v0.0.0
	github.com/hashicorp/go-hclog v1.5.0
	github.com/hashicorp/raft v1.6.0
	github.com/hashicorp/raft-boltdb/v2 v2.3.1
	github.com/jackc/pgx/v5 v5.7.4
	github.com/segmentio/kafka-go v0.4.47
	go.uber.org/zap v1.26.0
	google.golang.org/grpc v1.82.1
	google.golang.org/protobuf v1.36.11
)
