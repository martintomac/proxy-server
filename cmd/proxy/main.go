package main

import (
	"log"
	"os"
	"proxy-server/internal/proxy"
	"proxy-server/internal/server"
)

func main() {
	file, err := os.ReadFile("cmd/proxy/config.json")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	config, err := proxy.ReadConfigFromBytes(file)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	router, err := config.CreateRouter()
	if err != nil {
		log.Fatalf("Failed to create router: %v", err)
	}

	srv := server.New(router, server.DefaultConfig())
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
