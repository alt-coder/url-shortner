package service

import (
	proto "github.com/alt-coder/url-shortener/url-shortener/proto"
)

import (
	"gorm.io/gorm"
	"github.com/redis/go-redis/v9"
	"github.com/go-zookeeper/zk"
	"sync"
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
	ZookeeperHost    string
	ZookeeperPort    string
}

type UrlShortenerService struct {
	proto.UnimplementedURLShortenerServer
	Config Config
	PostgresClient *gorm.DB
	RedisClient *redis.Client
	ZookeeperClient *zk.Conn
	currentCounterVal int64
	uppLimitVal int64
	mu sync.Mutex
	isCounterExists bool
}
