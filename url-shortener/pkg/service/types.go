package service

import (
	"sync"

	"github.com/alt-coder/url-shortener/url-shortener/pkg/dataModel"
	proto "github.com/alt-coder/url-shortener/url-shortener/proto"
	"context"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/redis/go-redis/v9"
)

// ZkClientInterface defines the methods needed from a Zookeeper client.
type ZkClientInterface interface {
	Create(path string, data []byte, flags int32, acl []zk.ACL) (string, error)
	Get(path string) ([]byte, *zk.Stat, error)
	Set(path string, data []byte, version int32) (*zk.Stat, error)
	Exists(path string) (bool, *zk.Stat, error)
	Close()
}

// RedisClientInterface defines the methods needed from a Redis client.
type RedisClientInterface interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Ping(ctx context.Context) *redis.StatusCmd
	Close() error
}

type Config struct {
	GrpcPort         string
	HttpPort         string
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDBName   string
	RedisHost        string
	RedisPort        string
	RedisPassword    string
	ZookeeperHost    string
	ZookeeperPort    string
}

// UrlShortenerService encapsulates varies clients and counters for the service to work.
type UrlShortenerService struct {
	proto.UnimplementedURLShortenerServer
	Config            Config
	RedisClient       RedisClientInterface
	ZookeeperClient   ZkClientInterface
	currentCounterVal int64
	uppLimitVal       int64
	mu                sync.Mutex
	isCounterExists   bool
	db                dataModel.DataAccessLayer
}
