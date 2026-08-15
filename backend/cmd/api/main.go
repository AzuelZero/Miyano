package main

import (
	"log"
	"net/http"
	"time"

	"miyano/internal/config"
	transport "miyano/internal/transport/http"
)

func main() {
	cfg := config.Load()

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      transport.NewRouter(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Miyano API listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
