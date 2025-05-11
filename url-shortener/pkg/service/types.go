package service

import (
	proto "github.com/alt-coder/url-shortener/url-shortener/proto"
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
}
