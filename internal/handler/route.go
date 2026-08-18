package handler

import (
	"net/http"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/appraisal-crm/notification-service/internal/middleware"
	"github.com/appraisal-crm/notification-service/internal/service"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(svc service.Service, db *pgxpool.Pool, jwks keyfunc.Keyfunc, allowedOrigins []string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	}))
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// Public health check
	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		if err := db.Ping(req.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwks))

		h := NewHandler(svc)
		r.Get("/notifications", h.List)
		r.Get("/notifications/{id}", h.GetByID)
		r.Patch("/notifications/{id}/read", h.MarkRead)
	})

	return r
}
