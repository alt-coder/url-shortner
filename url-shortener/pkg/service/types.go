package service

import (
	"sync"

	"github.com/alt-coder/url-shortener/url-shortener/pkg/dataModel"
	proto "github.com/alt-coder/url-shortener/url-shortener/proto"
	"github.com/go-zookeeper/zk"
	"github.com/redis/go-redis/v9"
)

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

type UrlShortenerService struct {
	proto.UnimplementedURLShortenerServer
	Config            Config
	RedisClient       *redis.Client
	ZookeeperClient   *zk.Conn
	currentCounterVal int64
	uppLimitVal       int64
	mu                sync.Mutex
	isCounterExists   bool
	db                dataModel.DataAccessLayer
}
