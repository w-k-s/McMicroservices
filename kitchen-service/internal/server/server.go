package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/w-k-s/McMicroservices/kitchen-service/log"

	"github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
)

type Server struct {
	*http.Server
	shutdownGracePeriod time.Duration
}

func New(config config.ServerConfig, handler http.Handler) Server {
	server := &http.Server{
		Addr:           config.ListenAddress(),
		Handler:        handler,
		ReadTimeout:    config.ReadTimeout(),
		WriteTimeout:   config.WriteTimeout(),
		MaxHeaderBytes: config.MaxHeaderBytes(),
	}
	return Server{
		server,
		config.ShutdownGracePeriod(),
	}
}

func (s Server) ListenAndServe() {
	go func() {
		if err := s.Server.ListenAndServe(); err != nil {
			log.Printf("Error while listening and serving. Details: %s", err)
		}
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (kill -2)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownGracePeriod)
	defer cancel()

	s.Server.SetKeepAlivesEnabled(false)
	if err := s.Server.Shutdown(ctx); err != nil {
		log.Printf("Failed to shut down server gracefully in %s", s.shutdownGracePeriod)
	}
}
