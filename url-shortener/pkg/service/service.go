package service

import (
	"context"
	"log"
	"net"
	"net/http"

	proto "github.com/alt-coder/url-shortener/url-shortener/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func NewServer(cfg Config) *UrlShortenerService {
	return &UrlShortenerService{
		Config: cfg,
	}
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
	// Create a listener on TCP port 50051
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
		log.Fatal("Failed to register gateway:", err)
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
