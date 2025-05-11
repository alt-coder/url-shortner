package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/alt-coder/url-shortener/url-shortener/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type urlShortenerServer struct {
	proto.UnimplementedURLShortenerServer
}

func (s *urlShortenerServer) ShortenURL(ctx context.Context, req *proto.ShortenURLRequest) (*proto.ShortenURLResponse, error) {
	// Implement URL shortening logic here
	// TODO IMPLEMENT LOGIC HERE
	shortURL := "shortened_url"
	return &proto.ShortenURLResponse{ShortUrl: shortURL}, nil
}

func (s *urlShortenerServer) GetURL(ctx context.Context, req *proto.GetURLRequest) (*proto.GetURLResponse, error) {
	// TODO : IMPLEMENT lOGIC HERE
	longURL := "original_url"
	return &proto.GetURLResponse{LongUrl: longURL}, nil
}

// TODO ADD CONFIG READ FROM CONFIG MAP
func main() {
	// BELOW IS JUST BOILERPLATE CODE TO START A GRPC SERVER AND GRPC REST GATEWAY..
	// Create a listener on TCP port 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()

	// Register the URLShortener service with the gRPC server
	proto.RegisterURLShortenerServer(s, &urlShortenerServer{})

	// Enable reflection to allow clients to discover the service
	reflection.Register(s)

	// Serve gRPC server
	go func() {
		log.Println("Serving gRPC on :50051")
		if err := s.Serve(lis); err != nil {
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
		"localhost:50051",
		opts...,
	)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}

	// Register the gRPC gateway
	gwmux := runtime.NewServeMux()
	err = proto.RegisterURLShortenerHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalf("Failed to register gateway:", err)
	}

	// Serve the gRPC gateway
	r.PathPrefix("/").Handler(gwmux)

	// Create HTTP server
	srv := &http.Server{
		Handler: r,
		Addr:    ":8080",
	}

	log.Println("Serving gRPC-Gateway on :8080")
	// Start HTTP server (and proxy calls to gRPC server endpoint)
	log.Fatal(srv.ListenAndServe())
}
