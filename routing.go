package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/product"
	"github.com/nayefradwi/zanobia_inventory_manager/user"
)

func RegisterRoutes(provider *ServiceProvider) chi.Router {
	r := chi.NewRouter()
	r.Use(common.JsonResponseMiddleware)
	r.Get("/health-check", healthCheck)
	registerUserRoutes(r, provider)
	registerPermissionRoutes(r, provider)
	registerRoleRoutes(r, provider)
	registerProductRoutes(r, provider)
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

func registerRoleRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	roleController := user.NewRoleController(provider.services.roleService)
	roleRouter := chi.NewRouter()
	roleRouter.Post("/", roleController.CreateRole)
	roleRouter.Get("/", roleController.GetRoles)
	mainRouter.Mount("/roles", roleRouter)
}

func registerProductRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	productRouter := chi.NewRouter()
	registerUnitRoutes(productRouter, provider)
	mainRouter.Mount("/products", productRouter)
}

func registerUnitRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	unitController := product.NewUnitController(provider.services.unitService)
	unitRouter := chi.NewRouter()
	unitRouter.Post("/", unitController.CreateUnit)
	unitRouter.Get("/", unitController.GetAllUnits)
	registerUnitConversions(unitRouter, provider)
	mainRouter.Mount("/units", unitRouter)
}

func registerUnitConversions(mainRouter *chi.Mux, provider *ServiceProvider) {
	unitConversionController := product.NewUnitController(provider.services.unitService)
	unitConversionRouter := chi.NewRouter()
	unitConversionRouter.Post("/", unitConversionController.CreateConversion)
	unitConversionRouter.Post("/from-name", unitConversionController.CreateConversionFromName)
	unitConversionRouter.Post("/convert", unitConversionController.ConvertUnit)
	mainRouter.Mount("/unit-conversions", unitConversionRouter)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	ok, _ := json.Marshal(map[string]interface{}{"status": "ok"})
	w.Header().Set("Content-Type", "application/json")
	w.Write(ok)
}
