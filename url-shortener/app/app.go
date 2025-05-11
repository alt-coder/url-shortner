package main

import (
	"log"

	"github.com/alt-coder/url-shortener/url-shortener/pkg/service"
)

func main() {
	// Start the service
	log.Println("Starting the URL shortener service...")
	cfg := service.Config{
		GrpcPort: "50051",
		HttpPort: "8080",
	}
	srv := service.NewServer(cfg)

	log.Println("Serving gRPC-Gateway on :8080")
	log.Fatal(srv.Start())
}