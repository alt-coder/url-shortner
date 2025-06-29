package service

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/alt-coder/url-shortener/url-shortener/pkg/dataModel"

	base "github.com/alt-coder/url-shortener/base/go"
	proto "github.com/alt-coder/url-shortener/url-shortener/proto"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

var (
	// requestCounterFunc will be used for mocking.
	requestCounterFunc = func(s *UrlShortenerService) (int64, error) { return s.requestCounter() }
)

// NewUrlShortnerService creates and initializes a new UrlShortenerService.
// It sets up database connections (PostgreSQL), Redis client, and Zookeeper client.
// Configuration is read from environment variables.
func NewUrlShortnerService() (*UrlShortenerService, error) {
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
		RedisPassword:    os.Getenv(RedisPassword),
		ZookeeperHost:    os.Getenv(ZookeeperHost),
		ZookeeperPort:    os.Getenv(ZookeeperPort),
	}

	log.Printf("Connecting to PostgreSQL: %s:%s@%s/%s", cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresDBName)
	log.Printf("Connecting to Redis: %s:%s", cfg.RedisHost, cfg.RedisPort)
	log.Printf("Connecting to ZooKeeper: %s:%s", cfg.ZookeeperHost, cfg.ZookeeperPort)
	postgresPort, err := strconv.Atoi(cfg.PostgresPort)
	if err != nil {
		log.Printf("Could not connect to postgress port %s", cfg.PostgresPort)
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
		Password: cfg.RedisPassword,
		DB:       0, // TODO: Make this configurable
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
		db:                datamodelDB,
		RedisClient:       redisClient, // base.NewRedisClient returns *redis.Client which implements RedisClientInterface
		ZookeeperClient:   zkClient,    // base.NewZookeeperClient returns *zk.Conn which implements ZkClientInterface
		currentCounterVal: 0,
		uppLimitVal:       0,
		mu:                sync.Mutex{}, // Initialize the mutex
	}, nil
}

// ShortenURL takes a long URL and an API key, generates a unique short URL,
// stores the mapping, and returns the short URL.
// It validates the API key and uses a distributed counter (via Zookeeper) to generate unique IDs.
func (s *UrlShortenerService) ShortenURL(ctx context.Context, req *proto.ShortenURLRequest) (*proto.ShortenURLResponse, error) {
	apiKey := req.ApiKey
	originalURL := req.LongUrl

	if apiKey == "" {
		return nil, ErrMissingApiKey
	}

	//check if api key exists
	isValidApiKey, err := s.db.CheckAPIKey(apiKey)

	if err != nil {
		return nil, err
	}
	if !isValidApiKey {
		return nil, ErrInvalidApiKey
	}

	counter, err := requestCounterFunc(s)
	if err != nil {
		return nil, err
	}

	shortURL := base62Encode(counter)

	urlMapping := &dataModel.URLMapping{
		ShortURLID: shortURL,
		LongURL:    originalURL,
	}

	if err := s.db.CreateURLMapping(urlMapping); err != nil {
		return nil, err
	}

	return &proto.ShortenURLResponse{ShortUrl: shortURL}, nil
}

// GetURL retrieves the original long URL corresponding to a given short URL.
// It queries the database for the URL mapping.
func (s *UrlShortenerService) GetURL(ctx context.Context, req *proto.GetURLRequest) (*proto.GetURLResponse, error) {
	shortURL := req.ShortUrl

	longURL, err := s.db.GetLongURL(shortURL)
	if err != nil {
		return nil, err
	}

	return &proto.GetURLResponse{LongUrl: longURL}, nil
}

// CreateUser creates a new user in the system with the provided first name, last name, and email.
// It generates a unique API key for the new user and stores the user details in the database.
func (s *UrlShortenerService) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	user := &dataModel.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	if err := s.db.CreateUser(user); err != nil {
		return nil, err
	}

	return &proto.CreateUserResponse{
		UserId: strconv.FormatUint(uint64(user.ID), 10),
		ApiKey: user.APIKey.String(),
	}, nil

}

// FetchApiKey retrieves the API key for a user based on their email address.
// It queries the database for the user's API key.
func (s *UrlShortenerService) FetchApiKey(ctx context.Context, req *proto.FetchApiKeyRequest) (*proto.FetchApiKeyResponse, error) {
	apiKey, err := s.db.GetAPIKeyByEmail(req.Email)
	if err != nil {
		return nil, err
	}

	return &proto.FetchApiKeyResponse{
		ApiKey: apiKey,
	}, nil
}

// GetTopDomains retrieves the top 3 domains that have been shortened the most.
func (s *UrlShortenerService) GetTopDomains(ctx context.Context, req *proto.GetTopDomainsRequest) (*proto.GetTopDomainsResponse, error) {
	// Fetch top 3 domains from the database
	domainCounts, err := s.db.GetTopDomains(3)
	if err != nil {
		log.Printf("Error fetching top domains: %v", err)
		return nil, err
	}

	// Convert dataModel.DomainCount to proto.DomainMetric
	var protoDomainMetrics []*proto.DomainMetric
	for _, dc := range domainCounts {
		protoDomainMetrics = append(protoDomainMetrics, &proto.DomainMetric{
			Domain: dc.DomainName,
			Count:  dc.Count,
		})
	}

	return &proto.GetTopDomainsResponse{TopDomains: protoDomainMetrics}, nil
}

// Start initializes and starts the URL shortener service.
// It performs database auto-migration, starts the gRPC server,
// and sets up the HTTP gateway (proxy) to handle RESTful API calls.
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
		log.Fatalf("Failed to register gateway:%v", err)
		return err
	}

	// Add a handler for /d/{shortChar} to redirect to the full URL
	r.HandleFunc("/d/{shortChar}", s.redirectHandler)

	// Serve the gRPC gateway
	apiRouter := r.PathPrefix("/").Subrouter()
	apiRouter.PathPrefix("/").Handler(gwmux)

	// Create HTTP server
	srv := &http.Server{
		Handler: r,
		Addr:    ":" + s.Config.HttpPort,
	}

	log.Println("Serving gRPC-Gateway on :" + s.Config.HttpPort)
	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return srv.ListenAndServe()
}

// redirectHandler is an HTTP handler that takes a short URL character code from the path,
// retrieves the corresponding long URL using the GetURL service method,
// and then redirects the client to the long URL.
func (s *UrlShortenerService) redirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortChar := vars["shortChar"]

	// Use GetURL to get the full URL
	getURLRequest := &proto.GetURLRequest{ShortUrl: shortChar}
	resp, err := s.GetURL(context.Background(), getURLRequest)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	// Redirect to the full URL
	http.Redirect(w, r, resp.LongUrl, http.StatusFound)
}

// requestCounter provides a unique, incrementing counter value.
// It uses a local cache (currentCounterVal, uppLimitVal) for a batch of IDs
// and fetches a new batch from Zookeeper when the local cache is exhausted.
// This helps in reducing frequent calls to Zookeeper for every request.
// It ensures thread safety using a mutex.
func (s *UrlShortenerService) requestCounter() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentCounterVal >= s.uppLimitVal {
		//connect to zk and fetch the current count from zk
		conn := s.ZookeeperClient
		if !s.isCounterExists {
			err := checkZkCounter(conn)
			if err != nil {
				log.Print("Failed to create Counter")
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
