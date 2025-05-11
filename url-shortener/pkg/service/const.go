package service

import "errors"

const (
	PostgresHost     = "POSTGRES_HOST"
	PostgresPort     = "POSTGRES_PORT"
	PostgresUser     = "POSTGRES_USER"
	PostgresPassword = "POSTGRES_PASSWORD"
	PostgresDBName   = "POSTGRES_DBNAME"

	RedisHost = "REDIS_HOST"
	RedisPort = "REDIS_PORT"

	ZookeeperHost = "ZOOKEEPER_HOST"
	ZookeeperPort = "ZOOKEEPER_PORT"

	GrpcPort = "GRPC_PORT"
	HttpPort = "HTTP_PORT"
)

var (
	ErrMissingApiKey = errors.New("missing API key")
	ErrInvalidApiKey = errors.New("invalid API key")
)
