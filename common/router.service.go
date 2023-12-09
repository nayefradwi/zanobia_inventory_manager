package common

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	// TODO: find a way to modify cors options from main.go
	r.Use(cors.Handler(cors.Options{
		// TODO change when in production
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Warehouse-Id",
		},
		ExposedHeaders: []string{"Link"},
	}))
	r.Use(Recover)
	r.Use(JsonResponseMiddleware)
	r.Use(SetLanguageMiddleware)
	r.Use(SetPaginatedDataMiddleware)
	r.Get("/health-check", healthCheck)
	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	ok, _ := json.Marshal(map[string]interface{}{"status": "ok"})
	w.Header().Set("Content-Type", "application/json")
	w.Write(ok)
}
