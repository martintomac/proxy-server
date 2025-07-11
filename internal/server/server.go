package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		Port:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

type Server struct {
	config *Config
	server *http.Server
}

func New(handler http.Handler, config *Config) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	return &Server{
		config: config,
		server: &http.Server{
			Addr:         config.Port,
			Handler:      handler,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
	}
}

func (s *Server) Start() error {
	// Channel to listen for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on %s", s.config.Port)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-stop
	log.Println("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	log.Println("Server gracefully stopped")
	return nil
}
