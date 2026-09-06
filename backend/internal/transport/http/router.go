// Package http wires the HTTP transport: router, middlewares and route groups.
package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"miyano/internal/transport/http/handler"
)

// Handlers bundles the HTTP handlers the router serves.
type Handlers struct {
	Health    http.HandlerFunc
	Exercises *handler.ExerciseHandler
}

// NewRouter builds the application router with the standard middleware stack
// and the /api/v1 route tree (public routes, plus a protected group ready
// for authentication).
func NewRouter(h Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", h.Health)

		// Catalog endpoints are public PROVISIONALLY: they become protected
		// when the auth capability lands (see OpenSpec exercise-catalog).
		r.Get("/exercises", h.Exercises.List)
		r.Get("/exercises/{id}", h.Exercises.GetByID)

		r.Group(func(r chi.Router) {
			// Protected routes: auth middleware will be added here (JWT, ADR D003).
		})
	})

	return r
}
