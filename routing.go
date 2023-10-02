package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/product"
	"github.com/nayefradwi/zanobia_inventory_manager/user"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
)

func RegisterRoutes(provider *ServiceProvider) chi.Router {
	r := chi.NewRouter()
	r.Use(common.JsonResponseMiddleware)
	r.Use(common.SetLanguageMiddleware)
	r.Use(common.SetPaginatedDataMiddleware)
	r.Use(warehouse.SetWarehouseIdFromHeader)
	r.Get("/health-check", healthCheck)
	registerUserRoutes(r, provider)
	registerPermissionRoutes(r, provider)
	registerRoleRoutes(r, provider)
	registerProductRoutes(r, provider)
	registerWarehouseRoutes(r, provider)
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
	productController := product.NewProductController(provider.services.productService)
	productRouter.Post("/", productController.CreateProduct)
	productRouter.Get("/", productController.GetProducts)
	productRouter.Post("/translation", productController.TranslateProduct)
	registerUnitRoutes(productRouter, provider)
	registerIngredientRoutes(productRouter, provider)
	mainRouter.Mount("/products", productRouter)
}

func registerUnitRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	unitController := product.NewUnitController(provider.services.unitService)
	unitRouter := chi.NewRouter()
	unitRouter.Post("/", unitController.CreateUnit)
	unitRouter.Get("/{id}", unitController.GetUnitById)
	unitRouter.Get("/", unitController.GetAllUnits)
	unitRouter.Post("/translation", unitController.TranslateUnit)
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

func registerIngredientRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	ingredientController := product.NewIngredientController(provider.services.ingredientService)
	ingredientRouter := chi.NewRouter()
	ingredientRouter.Post("/", ingredientController.CreateIngredient)
	ingredientRouter.Get("/", ingredientController.GetIngredients)
	registerInventoryRoutes(ingredientRouter, provider)
	mainRouter.Mount("/ingredients", ingredientRouter)
}

func registerInventoryRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	inventoryController := product.NewInventoryController(provider.services.inventoryService)
	inventoryRouter := chi.NewRouter()
	inventoryRouter.Post("/inventory/stock", inventoryController.IncrementInventory)
	inventoryRouter.Delete("/inventory/stock", inventoryController.DecrementInventory)
	inventoryRouter.Post("/stock", inventoryController.BulkIncrementInventory)
	inventoryRouter.Delete("/stock", inventoryController.BulkDecrementInventory)
	inventoryRouter.Get("/", inventoryController.GetInventories)
	mainRouter.Mount("/inventories", inventoryRouter)
}

func registerWarehouseRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	warehouseController := warehouse.NewWarehouseController(provider.services.warehouseService)
	warehouseRouter := chi.NewRouter()
	warehouseRouter.Post("/", warehouseController.CreateWarehouse)
	warehouseRouter.Get("/", warehouseController.GetWarehouses)
	mainRouter.Mount("/warehouses", warehouseRouter)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	ok, _ := json.Marshal(map[string]interface{}{"status": "ok"})
	w.Header().Set("Content-Type", "application/json")
	w.Write(ok)
}
