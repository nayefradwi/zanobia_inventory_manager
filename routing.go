package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/product"
	"github.com/nayefradwi/zanobia_inventory_manager/retailer"
	"github.com/nayefradwi/zanobia_inventory_manager/transactions"
	"github.com/nayefradwi/zanobia_inventory_manager/user"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
)

func RegisterRoutes(provider *ServiceProvider) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
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
	r.Use(common.Recover)
	r.Use(common.JsonResponseMiddleware)
	r.Use(common.SetLanguageMiddleware)
	r.Use(common.SetPaginatedDataMiddleware)
	r.Use(warehouse.SetWarehouseIdFromHeader)
	r.Get("/health-check", healthCheck)
	registerUserRoutes(r, provider)
	registerPermissionRoutes(r, provider)
	authorizedRouter, _ := createSecureRouter(provider)
	// TODO: validate user warehouse
	registerRoleRoutes(authorizedRouter, provider)
	registerProductRoutes(authorizedRouter, provider)
	registerWarehouseRoutes(authorizedRouter, provider)
	registerRetailerRoutes(authorizedRouter, provider)
	registerTransactionRoutes(authorizedRouter, provider)
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
		r.With(middleware.HasPermissions(
			user.SysAdminPermissionHandle,
		)).Get("/", userController.GetAllUsers)
		r.Get("/me", userController.GetUserByContext)
		r.With(middleware.HasPermissions(
			user.SysAdminPermissionHandle,
		)).Post("/ban", userController.BanUser)
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
	roleRouter := chi.NewRouter()
	roleController := user.NewRoleController(provider.services.roleService)
	roleRouter.Use(common.AuthenticationHeaderMiddleware)
	roleRouter.Post("/", roleController.CreateRole)
	roleRouter.Get("/", roleController.GetRoles)
	mainRouter.Mount("/roles", roleRouter)
}

func registerProductRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	productRouter := chi.NewRouter()
	productController := product.NewProductController(provider.services.productService)
	productRouter.Post("/", productController.CreateProduct)
	productRouter.Get("/", productController.GetProducts)
	productRouter.Get("/{id}", productController.GetProduct)
	productRouter.Post("/translation", productController.TranslateProduct)
	registerUnitRoutes(productRouter, provider)
	registerProductVariantRoutes(productRouter, provider)
	mainRouter.Mount("/products", productRouter)
}

func registerUnitRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	unitRouter := chi.NewRouter()
	unitController := product.NewUnitController(provider.services.unitService)
	unitRouter.Post("/", unitController.CreateUnit)
	unitRouter.Get("/{id}", unitController.GetUnitById)
	unitRouter.Get("/", unitController.GetAllUnits)
	unitRouter.Post("/translation", unitController.TranslateUnit)
	registerUnitConversions(unitRouter, provider)
	mainRouter.Mount("/units", unitRouter)
}

func registerUnitConversions(mainRouter *chi.Mux, provider *ServiceProvider) {
	unitConversionRouter := chi.NewRouter()
	unitConversionController := product.NewUnitController(provider.services.unitService)
	unitConversionRouter.Post("/", unitConversionController.CreateConversion)
	unitConversionRouter.Post("/convert", unitConversionController.ConvertUnit)
	mainRouter.Mount("/unit-conversions", unitConversionRouter)
}

func registerProductVariantRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	productVariantRouter := chi.NewRouter()
	productController := product.NewProductController(provider.services.productService)
	productVariantRouter.Post("/", productController.CreateProductVariant)
	productVariantRouter.Get("/{id}", productController.GetProductVariant)
	productVariantRouter.Delete("/{id}", productController.DeleteProductVariant)
	productVariantRouter.Put("/", productController.UpdateProductVariantDetails)
	productVariantRouter.Put("/sku", productController.UpdateProductVariantSku)
	productVariantRouter.Post("/options/values", productController.AddOptionValue)
	registerRecipeRoutes(productVariantRouter, provider)
	registerBatchesRoutes(productVariantRouter, provider)
	mainRouter.Mount("/product-variants", productVariantRouter)
}
func registerRecipeRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	recipeRouter := chi.NewRouter()
	recipeController := product.NewRecipeController(provider.services.recipeService)
	recipeRouter.Post("/", recipeController.CreateRecipe)
	recipeRouter.Put("/recipe", recipeController.AddIngredientToRecipe)
	recipeRouter.Delete("/{id}", recipeController.DeleteRecipe)
	mainRouter.Mount("/recipes", recipeRouter)
}

func registerBatchesRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	batchRouter := chi.NewRouter()
	batchController := product.NewBatchController(provider.services.batchService)
	batchRouter.Post("/batch/stock", batchController.IncrementBatch)
	batchRouter.Delete("/batch/stock", batchController.DecrementBatch)
	batchRouter.Post("/stock", batchController.BulkIncrementBatch)
	batchRouter.Delete("/stock", batchController.BulkDecrementBatch)
	batchRouter.Post("/batch/stock/with-recipe", batchController.IncrementBatchWithRecipe)
	batchRouter.Post("/stock/with-recipe", batchController.BulkIncrementBatchWithRecipe)
	batchRouter.Get("/", batchController.GetBatches)
	batchRouter.Get("/search", batchController.SearchBatchesBySku)
	mainRouter.Mount("/batches", batchRouter)
}

func registerWarehouseRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	middleware := newUserMiddleWare(provider)
	warehouseRouter := chi.NewRouter()
	warehouseController := warehouse.NewWarehouseController(provider.services.warehouseService)
	warehouseRouter.Post("/", warehouseController.CreateWarehouse)
	warehouseRouter.
		With(middleware.HasPermissions(
			user.SysAdminPermissionHandle,
		)).
		Post("/user", warehouseController.AddUserToWarehouse)
	warehouseRouter.Get("/current", warehouseController.GetCurrentWarehouse)
	warehouseRouter.Get("/", warehouseController.GetWarehouses)
	warehouseRouter.
		With(middleware.HasPermissions(
			user.SysAdminPermissionHandle,
		)).
		Put("/", warehouseController.UpdateWarehouse)
	mainRouter.Mount("/warehouses", warehouseRouter)
}

func registerRetailerRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	middleware := newUserMiddleWare(provider)
	retailerRouter := chi.NewRouter()
	retailerController := retailer.NewRetailerController(provider.services.retailerService)
	batchController := retailer.NewRetailerBatchController(provider.services.retailerBatchService)
	retailerRouter.Post("/", retailerController.CreateRetailer)
	retailerRouter.Post("/{id}/contacts", retailerController.AddRetailerContacts)
	retailerRouter.Post("/{id}/contact", retailerController.AddRetailerContactInfo)
	retailerRouter.Get("/", retailerController.GetRetailers)
	retailerRouter.Get("/{id}", retailerController.GetRetailer)
	retailerRouter.Delete("/contact/{id}", retailerController.RemoveRetailerContactInfo)
	retailerRouter.
		With(middleware.HasPermissions(
			user.SysAdminPermissionHandle,
		)).
		Delete("/{id}", retailerController.RemoveRetailer)
	retailerRouter.Put("/", retailerController.UpdateRetailer)
	retailerRouter.Get("/{id}/batches", batchController.GetBatches)
	retailerRouter.Get("/{id}/batches/search", batchController.SearchBatchesBySku)
	registerRetailerBatchRoutes(retailerRouter, provider)
	mainRouter.Mount("/retailers", retailerRouter)
}

func registerRetailerBatchRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	batchRouter := chi.NewRouter()
	batchController := retailer.NewRetailerBatchController(provider.services.retailerBatchService)
	batchRouter.Post("/batch/stock", batchController.IncrementBatch)
	batchRouter.Delete("/batch/stock", batchController.DecrementBatch)
	batchRouter.Post("/stock", batchController.BulkIncrementBatch)
	batchRouter.Delete("/stock", batchController.BulkDecrementBatch)
	batchRouter.Post("/batch/stock/from-warehouse", batchController.MoveFromWarehouseToRetailer)
	batchRouter.Delete("/batch/stock/to-warehouse", batchController.ReturnToWarehouseToRetailer)
	mainRouter.Mount("/batches", batchRouter)
}

func registerTransactionRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	transactionRouter := chi.NewRouter()
	transactionController := transactions.NewTransactionController(provider.services.transactionService)
	transactionRouter.Post("/reasons", transactionController.CreateTransactionReason)
	transactionRouter.Get("/reasons", transactionController.GetTransactionReasons)
	transactionRouter.Get("/retailer/{id}", transactionController.GetTransactionsOfRetailer)
	transactionRouter.Get("/retailer/{id}/batch/{batchId}", transactionController.GetTransactionsOfRetailerBatch)
	transactionRouter.Get("/sku/{sku}", transactionController.GetTransactionsOfSku)
	transactionRouter.Get("/batch/{id}", transactionController.GetTransactionsOfBatch)
	transactionRouter.Get("/warehouse", transactionController.GetTransactionsOfMyWarehouse)
	mainRouter.Mount("/transactions", transactionRouter)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	ok, _ := json.Marshal(map[string]interface{}{"status": "ok"})
	w.Header().Set("Content-Type", "application/json")
	w.Write(ok)
}

func createSecureRouter(provider *ServiceProvider) (*chi.Mux, user.UserMiddleware) {
	r := chi.NewRouter()
	middleware := newUserMiddleWare(provider)
	r.Use(common.AuthenticationHeaderMiddleware)
	r.Use(middleware.SetUserFromHeader)
	return r, middleware
}

func newUserMiddleWare(provider *ServiceProvider) user.UserMiddleware {
	return user.NewUserMiddleware(provider.services.userService)
}
