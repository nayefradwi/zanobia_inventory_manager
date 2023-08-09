package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(provider *ServiceProvider) chi.Router {
	r := chi.NewRouter()
	r.Get("/health-check", healthCheck)
	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	ok, _ := json.Marshal(map[string]interface{}{"status": "ok"})
	w.Header().Set("Content-Type", "application/json")
	w.Write(ok)
}
