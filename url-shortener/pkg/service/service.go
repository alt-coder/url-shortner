package service

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	base "github.com/alt-coder/url-shortener/base/go"
	"github.com/alt-coder/url-shortener/url-shortener/pkg/dataModel"
	proto "github.com/alt-coder/url-shortener/url-shortener/proto"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func NewServer() (*UrlShortenerService, error) {
	cfg := Config{
		GrpcPort:         os.Getenv(GrpcPort),
		HttpPort:         os.Getenv(HttpPort),
		PostgresHost:     os.Getenv(PostgresHost),
		PostgresPort:     os.Getenv(PostgresPort),
		PostgresUser:     os.Getenv(PostgresUser),
		PostgresPassword: os.Getenv(PostgresPassword),
		PostgresDBName:   os.Getenv(PostgresDBName),
		RedisHost:        os.Getenv(RedisHost),
		RedisPort:        os.Getenv(RedisPort),
		ZookeeperHost:    os.Getenv(ZookeeperHost),
		ZookeeperPort:    os.Getenv(ZookeeperPort),
	}

	// TODO: Implement actual initialization logic
	log.Printf("Connecting to PostgreSQL: %s:%s@%s/%s", cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresDBName)
	log.Printf("Connecting to Redis: %s:%s", cfg.RedisHost, cfg.RedisPort)
	log.Printf("Connecting to ZooKeeper: %s:%s", cfg.ZookeeperHost, cfg.ZookeeperPort)
	postgresPort, err := strconv.Atoi(cfg.PostgresPort)
	if err != nil {
		log.Fatalf("Could not connect to postgress port %s", cfg.PostgresPort)
		return nil, err
	}
	postgresConfig := base.PostgresConfig{
		Host:     cfg.PostgresHost,
		Port:     postgresPort,
		User:     cfg.PostgresUser,
		Password: cfg.PostgresPassword,
		DBName:   cfg.PostgresDBName,
		SSLMode:  "disable", // TODO: Make this configurable
	}

	db, err := base.NewPostgresClient(postgresConfig)
	if err != nil {
		log.Printf("Error connecting to PostgreSQL: %v", err)
		return nil, err
	}

	redisConfig := base.RedisConfig{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: "", // TODO: Make this configurable
		DB:       0,  // TODO: Make this configurable
	}

	redisClient, err := base.NewRedisClient(redisConfig)
	if err != nil {
		log.Printf("Error connecting to Redis: %v", err)
		return nil, err
	}

	zookeeperConfig := base.ZookeeperConfig{
		Address: []string{cfg.ZookeeperHost + ":" + cfg.ZookeeperPort},
		Timeout: 5 * time.Second, // TODO: Make this configurable
	}

	zkClient, err := base.NewZookeeperClient(zookeeperConfig)
	if err != nil {
		log.Printf("Error connecting to Zookeeper: %v", err)
		return nil, err
	}
	datamodelDB := dataModel.NewDB(db)

	return &UrlShortenerService{
		Config:            cfg,
		db:    datamodelDB,
		RedisClient:       redisClient,
		ZookeeperClient:   zkClient,
		currentCounterVal: 0,
		uppLimitVal:       0,
	}, nil
}

func (s *UrlShortenerService) ShortenURL(ctx context.Context, req *proto.ShortenURLRequest) (*proto.ShortenURLResponse, error) {
	// Implement URL shortening logic here
	// TODO IMPLEMENT LOGIC HERE
	shortURL := "shortened_url"
	return &proto.ShortenURLResponse{ShortUrl: shortURL}, nil
}

func (s *UrlShortenerService) GetURL(ctx context.Context, req *proto.GetURLRequest) (*proto.GetURLResponse, error) {
	// TODO : IMPLEMENT lOGIC HERE
	longURL := "original_url"
	return &proto.GetURLResponse{LongUrl: longURL}, nil
}

func (s *UrlShortenerService) Start() error {
	// Auto migrate the database tables
	err := s.db.AutoMigrate(&dataModel.URLMapping{}, &dataModel.User{})
	if err != nil {
		log.Fatalf("failed to automigrate: %v", err)
		return err
	}
	//taking a mutex lock
	lis, err := net.Listen("tcp", ":"+s.Config.GrpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return err
	}

	// Create a gRPC server object
	grpcServer := grpc.NewServer()

	// Register the URLShortener service with the gRPC server
	proto.RegisterURLShortenerServer(grpcServer, s)

	// Enable reflection to allow clients to discover the service
	reflection.Register(grpcServer)

	// Serve gRPC server
	go func() {
		log.Println("Serving gRPC on :" + s.Config.GrpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Create a new HTTP router
	r := mux.NewRouter()

	// Create a gRPC dial option
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Create a gRPC connection to the server
	conn, err := grpc.NewClient(
		"localhost:"+s.Config.GrpcPort,
		opts...,
	)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
		return err
	}

	// Register the gRPC gateway
	gwmux := runtime.NewServeMux()
	err = proto.RegisterURLShortenerHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalf("Failed to register gateway:", err)
		return err
	}

	// Serve the gRPC gateway
	r.PathPrefix("/").Handler(gwmux)

	// Create HTTP server
	srv := &http.Server{
		Handler: r,
		Addr:    ":" + s.Config.HttpPort,
	}

	log.Println("Serving gRPC-Gateway on :" + s.Config.HttpPort)
	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return srv.ListenAndServe()
}

func (s *UrlShortenerService) requestCounter() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentCounterVal >= s.uppLimitVal {
		//connect to zk and fetch the current count from zk
		conn := s.ZookeeperClient
		if !s.isCounterExists{
			err := checkZkCounter(conn)
			if err != nil {
				log.Fatal("Failed to create Counter")
				return -1, err
			}
			s.isCounterExists = true
		}

		data, stat, err := conn.Get("/counter")
		if err != nil {
			log.Printf("Error getting data from Zookeeper: %v", err)
			return -1, err
		}
		counter, err := strconv.Atoi(string(data))
		if err != nil {
			log.Printf("Error converting data from Zookeeper to int: %v", err)
			return -1, err

		}
		//increment the zk counter by 10000
		newCounter := int64(counter + 10000)

		_, err = conn.Set("/counter", []byte(strconv.FormatInt(newCounter, 10)), stat.Version)
		if err != nil {
			log.Printf("Error setting data to Zookeeper: %v", err)
			return -1, err

		}
		//update the currentCounterVal and uppLimitVal with the updated zk value and zk value plus thousand respectively
		s.currentCounterVal = int64(counter)
		s.uppLimitVal = newCounter
		log.Printf("Updated currentCounterVal to %d and uppLimitVal to %d", s.currentCounterVal, s.uppLimitVal)
	}
	s.currentCounterVal += 1
	return s.currentCounterVal, nil
}
