package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/product"
	"github.com/nayefradwi/zanobia_inventory_manager/user"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
)

func RegisterRoutes(provider *ServiceProvider) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(common.Recover)
	r.Use(common.JsonResponseMiddleware)
	r.Use(common.SetLanguageMiddleware)
	r.Use(common.SetPaginatedDataMiddleware)
	r.Use(warehouse.SetWarehouseIdFromHeader)
	r.Get("/health-check", healthCheck)
	registerUserRoutes(r, provider)
	registerPermissionRoutes(r, provider)
	userMiddleWare := user.NewUserMiddleware(provider.services.userService)
	authorizedRouter := chi.NewRouter()
	authorizedRouter.Use(common.AuthenticationHeaderMiddleware)
	authorizedRouter.Use(userMiddleWare.SetUserFromHeader)
	registerRoleRoutes(authorizedRouter, provider)
	registerProductRoutes(authorizedRouter, provider)
	registerWarehouseRoutes(authorizedRouter, provider)
	r.Mount("/", authorizedRouter)
	return r
}

func registerUserRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	userController := user.NewUserController(provider.services.userService)
	userRouter := chi.NewRouter()
	userRouter.Group(func(r chi.Router) {
		middleware := user.NewUserMiddleware(provider.services.userService)
		r.Use(common.AuthenticationHeaderMiddleware)
		r.Use(middleware.SetUserFromHeader)
		r.With(middleware.HasPermissions(
			user.SysAdminPermissionHandle,
		)).Post("/", userController.CreateUser)
		r.Get("/", userController.GetAllUsers)
		r.Get("/me", userController.GetUserByContext)
		r.Post("/ban", userController.BanUser)
	})
	userRouter.Post("/initial-sys-admin", userController.InitiateSysAdminUser)
	userRouter.Post("/login", userController.LoginUser)
	mainRouter.Mount("/users", userRouter)
}

func registerPermissionRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	permissionController := user.NewPermissionController(provider.services.permissionService)
	permissionRouter := chi.NewRouter()
	permissionRouter.Post("/initial-permissions", permissionController.InitiateInitialPermissions)
	permissionRouter.Group(func(r chi.Router) {
		middleware := user.NewUserMiddleware(provider.services.userService)
		r.Use(common.AuthenticationHeaderMiddleware)
		r.Use(middleware.SetUserFromHeader)
		r.Post("/", permissionController.CreatePermission)
		r.Get("/", permissionController.GetAllPermissions)
		getPermissionByHandleRoute := fmt.Sprintf("/{%s}", user.PermissionHandleParam)
		r.Get(getPermissionByHandleRoute, permissionController.GetPermissionByHandle)
	})
	mainRouter.Mount("/permissions", permissionRouter)
}

func registerRoleRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	roleRouter, _ := createRouter(provider)
	roleController := user.NewRoleController(provider.services.roleService)
	roleRouter.Use(common.AuthenticationHeaderMiddleware)
	roleRouter.Post("/", roleController.CreateRole)
	roleRouter.Get("/", roleController.GetRoles)
	mainRouter.Mount("/roles", roleRouter)
}

func registerProductRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	productRouter, _ := createRouter(provider)
	productController := product.NewProductController(provider.services.productService)
	productRouter.Post("/", productController.CreateProduct)
	productRouter.Get("/", productController.GetProducts)
	productRouter.Get("/{id}", productController.GetProduct)
	productRouter.Post("/translation", productController.TranslateProduct)
	registerUnitRoutes(productRouter, provider)
	registerIngredientRoutes(productRouter, provider)
	registerProductVariantRoutes(productRouter, provider)
	mainRouter.Mount("/products", productRouter)
}

func registerUnitRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	unitRouter, _ := createRouter(provider)
	unitController := product.NewUnitController(provider.services.unitService)
	unitRouter.Post("/", unitController.CreateUnit)
	unitRouter.Get("/{id}", unitController.GetUnitById)
	unitRouter.Get("/", unitController.GetAllUnits)
	unitRouter.Post("/translation", unitController.TranslateUnit)
	registerUnitConversions(unitRouter, provider)
	mainRouter.Mount("/units", unitRouter)
}

func registerUnitConversions(mainRouter *chi.Mux, provider *ServiceProvider) {
	unitConversionRouter, _ := createRouter(provider)
	unitConversionController := product.NewUnitController(provider.services.unitService)
	unitConversionRouter.Post("/", unitConversionController.CreateConversion)
	unitConversionRouter.Post("/from-name", unitConversionController.CreateConversionFromName)
	unitConversionRouter.Post("/convert", unitConversionController.ConvertUnit)
	mainRouter.Mount("/unit-conversions", unitConversionRouter)
}

func registerIngredientRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	ingredientRouter, _ := createRouter(provider)
	ingredientController := product.NewIngredientController(provider.services.ingredientService)
	ingredientRouter.Post("/", ingredientController.CreateIngredient)
	ingredientRouter.Get("/", ingredientController.GetIngredients)
	registerInventoryRoutes(ingredientRouter, provider)
	mainRouter.Mount("/ingredients", ingredientRouter)
}

func registerInventoryRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	inventoryRouter, _ := createRouter(provider)
	inventoryController := product.NewInventoryController(provider.services.inventoryService)
	inventoryRouter.Post("/inventory/stock", inventoryController.IncrementInventory)
	inventoryRouter.Delete("/inventory/stock", inventoryController.DecrementInventory)
	inventoryRouter.Post("/stock", inventoryController.BulkIncrementInventory)
	inventoryRouter.Delete("/stock", inventoryController.BulkDecrementInventory)
	inventoryRouter.Get("/", inventoryController.GetInventories)
	mainRouter.Mount("/inventories", inventoryRouter)
}

func registerProductVariantRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	productVariantRouter, _ := createRouter(provider)
	productController := product.NewProductController(provider.services.productService)
	productVariantRouter.Post("/", productController.CreateProductVariant)
	productVariantRouter.Get("/{id}", productController.GetProductVariant)
	registerRecipeRoutes(productVariantRouter, provider)
	registerBatchesRoutes(productVariantRouter, provider)
	mainRouter.Mount("/product-variants", productVariantRouter)
}
func registerRecipeRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	recipeRouter, _ := createRouter(provider)
	recipeController := product.NewRecipeController(provider.services.recipeService)
	recipeRouter.Post("/", recipeController.CreateRecipe)
	recipeRouter.Put("/recipe", recipeController.AddIngredientToRecipe)
	recipeRouter.Delete("/{id}", recipeController.DeleteRecipe)
	mainRouter.Mount("/recipes", recipeRouter)
}

func registerBatchesRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	batchRouter, _ := createRouter(provider)
	batchController := product.NewBatchController(provider.services.batchService)
	batchRouter.Post("/batch/stock", batchController.IncrementBatch)
	batchRouter.Delete("/batch/stock", batchController.DecrementBatch)
	batchRouter.Post("/stock", batchController.BulkIncrementBatch)
	batchRouter.Delete("/stock", batchController.BulkDecrementBatch)
	batchRouter.Get("/", batchController.GetBatches)
	batchRouter.Get("/search", batchController.SearchBatchesBySku)
	mainRouter.Mount("/batches", batchRouter)
}

func registerWarehouseRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	warehouseRouter, _ := createRouter(provider)
	warehouseController := warehouse.NewWarehouseController(provider.services.warehouseService)
	warehouseRouter.Post("/", warehouseController.CreateWarehouse)
	warehouseRouter.Get("/", warehouseController.GetWarehouses)
	mainRouter.Mount("/warehouses", warehouseRouter)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	ok, _ := json.Marshal(map[string]interface{}{"status": "ok"})
	w.Header().Set("Content-Type", "application/json")
	w.Write(ok)
}

func createRouter(provider *ServiceProvider) (*chi.Mux, user.UserMiddleware) {
	r := chi.NewRouter()
	middleware := user.NewUserMiddleware(provider.services.userService)
	r.Use(common.AuthenticationHeaderMiddleware)
	r.Use(middleware.SetUserFromHeader)
	return r, middleware
}
