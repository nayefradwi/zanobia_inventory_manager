package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/user"
)

func RegisterRoutes(provider *ServiceProvider) chi.Router {
	r := chi.NewRouter()
	r.Use(common.JsonResponseMiddleware)
	r.Get("/health-check", healthCheck)
	registerPermissionRoutes(r, provider)
	return r
}

func registerPermissionRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	permissionController := user.NewPermissionController(provider.services.permissionService)
	permissionRouter := chi.NewRouter()
	permissionRouter.Post("/initial-permissions", permissionController.InitiateInitialPermissions)
	permissionRouter.Post("/", permissionController.CreatePermission)
	mainRouter.Mount("/permissions", permissionRouter)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	ok, _ := json.Marshal(map[string]interface{}{"status": "ok"})
	w.Header().Set("Content-Type", "application/json")
	w.Write(ok)
}
