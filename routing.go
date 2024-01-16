package main

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/product"
	"github.com/nayefradwi/zanobia_inventory_manager/retailer"
	"github.com/nayefradwi/zanobia_inventory_manager/transactions"
	"github.com/nayefradwi/zanobia_inventory_manager/unit"
	"github.com/nayefradwi/zanobia_inventory_manager/user"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
)

func RegisterRoutes(provider *ServiceProvider) chi.Router {
	baseRouter := common.NewRouter()
	r := chi.NewRouter()
	r.Use(warehouse.SetWarehouseIdFromHeader)
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
	baseRouter.Mount("/", r)
	return baseRouter
}

func registerUserRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	userController := user.NewUserController(provider.services.userService)
	userRouter := chi.NewRouter()
	userRouter.Group(func(r chi.Router) {
		middleware := user.NewUserMiddleware(provider.services.userService)
		adminMiddleware := middleware.HasPermissions(user.SysAdminPermissionHandle)
		r.Use(common.AuthenticationHeaderMiddleware)
		r.Use(middleware.SetUserFromHeader)
		r.Get("/me", userController.GetUserByContext)
		r.Group(func(r chi.Router) {
			r.With(adminMiddleware).Get("/", userController.GetAllUsers)
			r.With(adminMiddleware).Post("/", userController.CreateUser)
			r.With(adminMiddleware).Post("/ban", userController.BanUser)
		})
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
		r.Use(middleware.HasPermissions(user.SysAdminPermissionHandle))
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
	userMiddleware := newUserMiddleWare(provider)
	roleRouter.Use(common.AuthenticationHeaderMiddleware)
	roleRouter.Use(userMiddleware.HasPermissions(user.SysAdminPermissionHandle))
	roleRouter.Post("/", roleController.CreateRole)
	roleRouter.Get("/", roleController.GetRoles)
	mainRouter.Mount("/roles", roleRouter)
}

func registerProductRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	productRouter := chi.NewRouter()
	productController := product.NewProductController(provider.services.productService)
	productRouter.Group(func(r chi.Router) {
		userMiddlware := newUserMiddleWare(provider)
		controlProductMiddlware := userMiddlware.HasPermissions(user.HasProductControlPermission)
		deleteProductMiddleware := userMiddlware.HasPermissions(user.CanDeleteProductPermission)
		r.Use(controlProductMiddlware)
		r.Post("/", productController.CreateProduct)
		r.Put("/{id}/archive", productController.ArchiveProduct)
		r.Put("/{id}/unarchive", productController.UnarchiveProduct)
		r.Post("/translation", productController.TranslateProduct)
		r.Post("/options", productController.AddProductOptionToProduct)
		r.With(deleteProductMiddleware).Delete("/{id}", productController.DeleteProduct)
	})
	productRouter.Get("/", productController.GetProducts)
	productRouter.Get("/{id}", productController.GetProduct)
	registerUnitRoutes(productRouter, provider)
	registerProductVariantRoutes(productRouter, provider)
	mainRouter.Mount("/products", productRouter)
}

func registerUnitRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	unitRouter := chi.NewRouter()
	unitController := unit.NewUnitController(provider.services.unitService)
	unitRouter.Get("/{id}", unitController.GetUnitById)
	unitRouter.Get("/", unitController.GetAllUnits)
	unitRouter.Group(func(r chi.Router) {
		userMiddleware := newUserMiddleWare(provider)
		adminPermissionMiddleware := userMiddleware.HasPermissions(user.SysAdminPermissionHandle)
		r.Use(adminPermissionMiddleware)
		r.Post("/", unitController.CreateUnit)
		r.Post("/translation", unitController.TranslateUnit)
	})
	registerUnitConversions(unitRouter, provider)
	mainRouter.Mount("/units", unitRouter)
}

func registerUnitConversions(mainRouter *chi.Mux, provider *ServiceProvider) {
	unitConversionRouter := chi.NewRouter()
	unitConversionController := unit.NewUnitController(provider.services.unitService)
	userMiddleware := newUserMiddleWare(provider)
	adminPermissionMiddleware := userMiddleware.HasPermissions(user.SysAdminPermissionHandle)
	unitConversionRouter.With(adminPermissionMiddleware).Post("/", unitConversionController.CreateConversion)
	unitConversionRouter.Post("/convert", unitConversionController.ConvertUnit)
	mainRouter.Mount("/unit-conversions", unitConversionRouter)
}

func registerProductVariantRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	productVariantRouter := chi.NewRouter()
	productController := product.NewProductController(provider.services.productService)
	productVariantRouter.Group(func(r chi.Router) {
		userMiddleware := newUserMiddleWare(provider)
		controlProductMiddleware := userMiddleware.HasPermissions(user.HasProductControlPermission)
		deleteProductMiddleware := userMiddleware.HasPermissions(user.CanDeleteProductPermission)
		r.Use(controlProductMiddleware)
		r.Post("/", productController.CreateProductVariant)
		r.Put("/", productController.UpdateProductVariantDetails)
		r.Put("/sku", productController.UpdateProductVariantSku)
		r.Put("/{id}/archive", productController.ArchiveProductVariant)
		r.Put("/{id}/unarchive", productController.UnarchiveProductVariant)
		r.Post("/options/values", productController.AddOptionValue)
		r.With(deleteProductMiddleware).Delete("/{id}", productController.DeleteProductVariant)
	})
	productVariantRouter.Get("/{id}", productController.GetProductVariant)
	productVariantRouter.Get("/sku/{sku}", productController.GetProductVariantBySku)
	productVariantRouter.Post("/search", productController.SearchProductVariantByName)
	registerRecipeRoutes(productVariantRouter, provider)
	registerBatchesRoutes(productVariantRouter, provider)
	mainRouter.Mount("/product-variants", productVariantRouter)
}

func registerRecipeRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	recipeRouter := chi.NewRouter()
	recipeController := product.NewRecipeController(provider.services.recipeService)
	newUserMiddleWare := newUserMiddleWare(provider)
	controlProductMiddleware := newUserMiddleWare.HasPermissions(user.HasProductControlPermission)
	recipeRouter.Use(controlProductMiddleware)
	recipeRouter.Post("/", recipeController.CreateRecipe)
	recipeRouter.Put("/recipe", recipeController.AddIngredientToRecipe)
	recipeRouter.Delete("/{id}", recipeController.DeleteRecipe)
	mainRouter.Mount("/recipes", recipeRouter)
}

func registerBatchesRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	batchRouter := chi.NewRouter()
	batchController := product.NewBatchController(provider.services.batchService)
	batchRouter.Group(func(r chi.Router) {
		newUserMiddleWare := newUserMiddleWare(provider)
		controlBatchMiddleware := newUserMiddleWare.HasPermissions(user.HasBatchControlPermission)
		r.Use(controlBatchMiddleware)
		r.Post("/batch/stock", batchController.IncrementBatch)
		r.Delete("/batch/stock", batchController.DecrementBatch)
		r.Post("/stock", batchController.BulkIncrementBatch)
		r.Delete("/stock", batchController.BulkDecrementBatch)
		r.Post("/batch/stock/with-recipe", batchController.IncrementBatchWithRecipe)
		r.Post("/stock/with-recipe", batchController.BulkIncrementBatchWithRecipe)
	})
	batchRouter.Get("/", batchController.GetBatches)
	batchRouter.Get("/search", batchController.SearchBatchesBySku)
	batchRouter.Get("/{id}", batchController.GetBatchById)
	mainRouter.Mount("/batches", batchRouter)
}

func registerWarehouseRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	warehouseRouter := chi.NewRouter()
	warehouseController := warehouse.NewWarehouseController(provider.services.warehouseService)
	warehouseRouter.Group(func(r chi.Router) {
		middleware := newUserMiddleWare(provider)
		adminMiddleware := middleware.HasPermissions(user.SysAdminPermissionHandle)
		r.Use(adminMiddleware)
		r.Post("/user", warehouseController.AddUserToWarehouse)
		r.Put("/", warehouseController.UpdateWarehouse)
		r.Post("/", warehouseController.CreateWarehouse)
	})
	warehouseRouter.Get("/current", warehouseController.GetCurrentWarehouse)
	warehouseRouter.Get("/", warehouseController.GetWarehouses)
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
	retailerRouter.Get("/{id}/batches", batchController.GetBatchesOfRetailer)
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
	batchRouter.Get("/", batchController.GetBatches)
	// batchRouter.Post("/batch/stock/from-warehouse", batchController.MoveFromWarehouseToRetailer)
	// batchRouter.Delete("/batch/stock/to-warehouse", batchController.ReturnToWarehouseToRetailer)
	mainRouter.Mount("/batches", batchRouter)
}

func registerTransactionRoutes(mainRouter *chi.Mux, provider *ServiceProvider) {
	transactionRouter := chi.NewRouter()
	transactionController := transactions.NewTransactionController(provider.services.transactionService)
	userMiddleware := newUserMiddleWare(provider)
	adminMiddleware := userMiddleware.HasPermissions(user.SysAdminPermissionHandle)
	transactionRouter.With(adminMiddleware).Post("/reasons", transactionController.CreateTransactionReason)
	transactionRouter.Get("/reasons", transactionController.GetTransactionReasons)
	transactionRouter.Get("/retailer/{id}", transactionController.GetTransactionsOfRetailer)
	transactionRouter.Get("/retailer/{id}/batch/{batchId}", transactionController.GetTransactionsOfRetailerBatch)
	transactionRouter.Get("/sku/{sku}", transactionController.GetTransactionsOfSku)
	transactionRouter.Get("/batch/{id}", transactionController.GetTransactionsOfBatch)
	transactionRouter.Get("/warehouse", transactionController.GetTransactionsOfMyWarehouse)
	mainRouter.Mount("/transactions", transactionRouter)
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
