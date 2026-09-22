// Package http wires the HTTP transport: router, middlewares and route groups.
package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"miyano/internal/ports"
	"miyano/internal/transport/http/handler"
	appmiddleware "miyano/internal/transport/http/middleware"
)

// Handlers bundles the HTTP handlers the router serves.
type Handlers struct {
	Health    http.HandlerFunc
	Exercises *handler.ExerciseHandler
	Auth      *handler.AuthHandler
	// Verifier authenticates the protected group (JWT access tokens).
	Verifier ports.TokenVerifier
}

// NewRouter builds the application router with the standard middleware stack
// and the /api/v1 route tree: public health, rate-limited auth endpoints and
// a JWT-protected group.
func NewRouter(h Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", h.Health)

		// Auth is the brute-force surface: 5 requests per minute per IP.
		r.Route("/auth", func(r chi.Router) {
			r.Use(httprate.LimitByIP(5, time.Minute))
			r.Post("/register", h.Auth.Register)
			r.Post("/login", h.Auth.Login)
			r.Post("/refresh", h.Auth.Refresh)
		})

		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.RequireAuth(h.Verifier))
			r.Get("/exercises", h.Exercises.List)
			r.Get("/exercises/{id}", h.Exercises.GetByID)
		})
	})

	return r
}
