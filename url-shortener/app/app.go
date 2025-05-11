package main

import (
	"log"

	"github.com/alt-coder/url-shortener/url-shortener/pkg/service"
)

func main() {
	// Start the service
	log.Println("Starting the URL shortener service...")

	srv, err := service.NewServer()

	if err != nil {
		log.Fatalf("error occured while creating server %s", err)
	}

	log.Println("Serving gRPC-Gateway on :8080")
	log.Fatal(srv.Start())
}
