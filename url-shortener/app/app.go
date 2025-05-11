package main

import (
	"log"

	"github.com/alt-coder/url-shortener/url-shortener/pkg/service"
)

func main() {
	// Start the service
	log.Println("Starting the URL shortener service...")

	srv := service.NewServer()

	log.Println("Serving gRPC-Gateway on :8080")
	log.Fatal(srv.Start())
}
