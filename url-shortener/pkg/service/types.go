package service

import (
	proto "github.com/alt-coder/url-shortener/url-shortener/proto"
)

type Config struct {
	GrpcPort string
	HttpPort string
}

type UrlShortenerService struct {
	proto.UnimplementedURLShortenerServer
	Config Config
}
