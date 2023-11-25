package main

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/product"
	"github.com/nayefradwi/zanobia_inventory_manager/retailer"
	"github.com/nayefradwi/zanobia_inventory_manager/user"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
	"github.com/redis/go-redis/v9"
)

var connections systemConnections

type systemConnections struct {
	dbPool      *pgxpool.Pool
	redisClient *redis.Client
}
type systemRepositories struct {
	userRepository          user.IUserRepository
	permissionRepository    user.IPermissionRepository
	roleRepository          user.IRoleRepository
	unitRepository          product.IUnitRepository
	warehouseRepository     warehouse.IWarehouseRepository
	productRepository       product.IProductRepo
	recipeRepository        product.IRecipeRepository
	batchRepository         product.IBatchRepository
	retailerRepository      retailer.IRetailerRepository
	retailerBatchRepository retailer.IRetailerBatchRepository
}

type systemServices struct {
	userService          user.IUserService
	permissionService    user.IPermissionService
	roleService          user.IRoleService
	unitService          product.IUnitService
	warehouseService     warehouse.IWarehouseService
	lockingService       common.IDistributedLockingService
	productService       product.IProductService
	recipeService        product.IRecipeService
	batchService         product.IBatchService
	retailerService      retailer.IRetailerService
	retailerBatchService retailer.IRetailerBatchService
}
type ServiceProvider struct {
	services systemServices
}

func (s *ServiceProvider) initiate(config ApiConfig) {
	connections = s.setUpConnections(config)
	repositories := s.registerRepositories(connections)
	s.registerServices(repositories)
}

func (s *ServiceProvider) setUpConnections(config ApiConfig) systemConnections {
	ctx := context.Background()
	dbPool := common.ConnectDatabasePool(ctx, config.DbConnectionUrl)
	redisClient := common.ConnectRedis(ctx, config.RedisUrl)
	return systemConnections{
		dbPool:      dbPool,
		redisClient: redisClient,
	}
}

func (s *ServiceProvider) registerRepositories(connections systemConnections) systemRepositories {
	userRepo := user.NewUserRepository(connections.dbPool)
	permssionRepo := user.NewPermissionRepository(connections.dbPool)
	roleRepo := user.NewRoleRepository(connections.dbPool)
	unitRepo := product.NewUnitRepository(connections.dbPool)
	warehouseRepo := warehouse.NewWarehouseRepository(connections.dbPool)
	productRepo := product.NewProductRepository(connections.dbPool)
	recipeRepo := product.NewRecipeRepository(connections.dbPool)
	batchRepo := product.NewBatchRepository(connections.dbPool)
	retailerRepo := retailer.NewRetailerRepository(connections.dbPool)
	retailerBatchRepo := retailer.NewRetailerBatchRepository(connections.dbPool)
	return systemRepositories{
		userRepository:          userRepo,
		permissionRepository:    permssionRepo,
		roleRepository:          roleRepo,
		unitRepository:          unitRepo,
		warehouseRepository:     warehouseRepo,
		productRepository:       productRepo,
		recipeRepository:        recipeRepo,
		batchRepository:         batchRepo,
		retailerRepository:      retailerRepo,
		retailerBatchRepository: retailerBatchRepo,
	}
}

func (s *ServiceProvider) registerServices(repositories systemRepositories) {
	lockingService := common.CreateNewRedisLockService(connections.redisClient)
	userServiceInput := user.UserServiceInput{
		Repository:       repositories.userRepository,
		SysAdminEmail:    RegisteredApiConfig.InitialSysAdminEmail,
		SysAdminPassword: RegisteredApiConfig.InitialSysAdminPass,
	}
	userService := user.NewUserService(userServiceInput)
	permissionService := user.NewPermissionService(repositories.permissionRepository)
	roleService := user.NewRoleService(repositories.roleRepository)
	unitService := product.NewUnitService(repositories.unitRepository)
	warehouseService := warehouse.NewWarehouseService(repositories.warehouseRepository)
	recipeService := product.NewRecipeService(repositories.recipeRepository, unitService)
	productService := product.NewProductService(repositories.productRepository, recipeService)
	batchService := product.NewBatchService(
		repositories.batchRepository,
		productService,
		lockingService,
		unitService,
	)
	retailerBatchService := retailer.NewRetailerBatchService(
		repositories.retailerBatchRepository,
		productService,
		lockingService,
		unitService,
	)
	retailerService := retailer.NewRetailerService(repositories.retailerRepository, retailerBatchService)
	s.services = systemServices{
		userService:          userService,
		permissionService:    permissionService,
		roleService:          roleService,
		unitService:          unitService,
		warehouseService:     warehouseService,
		lockingService:       lockingService,
		productService:       productService,
		recipeService:        recipeService,
		batchService:         batchService,
		retailerService:      retailerService,
		retailerBatchService: retailerBatchService,
	}
}

func cleanUp() {
	connections.dbPool.Close()
	connections.redisClient.Close()
}
