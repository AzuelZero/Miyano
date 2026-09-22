package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"miyano/internal/adapter/argon2"
	"miyano/internal/adapter/jwttoken"
	"miyano/internal/adapter/postgres"
	"miyano/internal/config"
	"miyano/internal/ports"
	"miyano/internal/service"
	transport "miyano/internal/transport/http"
	"miyano/internal/transport/http/handler"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	pool, err := postgres.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database unavailable: %v", err)
	}
	defer pool.Close()

	var exerciseRepo ports.ExerciseRepository = postgres.NewExerciseRepository(pool)
	exerciseSvc := service.NewExerciseService(exerciseRepo)

	var (
		hasher  ports.PasswordHasher     = argon2.NewHasher()
		tokens  ports.TokenGenerator     = jwttoken.NewGenerator(cfg.JWTSecret, accessTokenTTL, refreshTokenTTL)
		userRepo ports.UserAuthRepository = postgres.NewUserAuthRepository(pool)
	)
	authSvc := service.NewAuthService(userRepo, hasher, tokens)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      transport.NewRouter(transport.Handlers{Health: handler.Health(), Exercises: handler.NewExerciseHandler(exerciseSvc), Auth: handler.NewAuthHandler(authSvc), Verifier: tokens}),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Miyano API listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
