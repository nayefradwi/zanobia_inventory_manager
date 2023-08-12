package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/user"
)

func RegisterRoutes(provider *ServiceProvider) chi.Router {
	r := chi.NewRouter()
	r.Use(common.JsonResponseMiddleware)
	r.Get("/health-check", healthCheck)
	registerUserRoutes(r, provider)
	registerPermissionRoutes(r, provider)
	return r
}

func registerUserRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	userController := user.NewUserController(provider.services.userService)
	userRouter := chi.NewRouter()
	userRouter.Post("/", userController.CreateUser)
	userRouter.Post("/initial-sys-admin", userController.InitiateSysAdminUser)
	userRouter.Get("/", userController.GetAllUsers)
	mainRouter.Mount("/users", userRouter)
}

func registerPermissionRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	permissionController := user.NewPermissionController(provider.services.permissionService)
	permissionRouter := chi.NewRouter()
	permissionRouter.Post("/initial-permissions", permissionController.InitiateInitialPermissions)
	permissionRouter.Post("/", permissionController.CreatePermission)
	permissionRouter.Get("/", permissionController.GetAllPermissions)
	getPermissionByHandleRoute := fmt.Sprintf("/{%s}", user.PermissionHandleParam)
	permissionRouter.Get(getPermissionByHandleRoute, permissionController.GetPermissionByHandle)
	mainRouter.Mount("/permissions", permissionRouter)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	ok, _ := json.Marshal(map[string]interface{}{"status": "ok"})
	w.Header().Set("Content-Type", "application/json")
	w.Write(ok)
}
