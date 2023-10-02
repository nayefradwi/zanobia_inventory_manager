package main

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/product"
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
	userRepository       user.IUserRepository
	permissionRepository user.IPermissionRepository
	roleRepository       user.IRoleRepository
	unitRepository       product.IUnitRepository
	warehouseRepository  warehouse.IWarehouseRepository
	ingredientRepository product.IIngredientRepository
	inventoryRepository  product.IInventoryRepository
	productRepository    product.IProductRepo
}

type systemServices struct {
	userService       user.IUserService
	permissionService user.IPermissionService
	roleService       user.IRoleService
	unitService       product.IUnitService
	warehouseService  warehouse.IWarehouseService
	ingredientService product.IIngredientService
	lockingService    common.IDistributedLockingService
	inventoryService  product.IInventoryService
	productService    product.IProductService
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
	ingredientRepo := product.NewIngredientRepository(connections.dbPool)
	inventoryRepo := product.NewInventoryRepository(connections.dbPool)
	productRepo := product.NewProductRepository(connections.dbPool)
	return systemRepositories{
		userRepository:       userRepo,
		permissionRepository: permssionRepo,
		roleRepository:       roleRepo,
		unitRepository:       unitRepo,
		warehouseRepository:  warehouseRepo,
		ingredientRepository: ingredientRepo,
		inventoryRepository:  inventoryRepo,
		productRepository:    productRepo,
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
	ingredientService := product.NewIngredientService(repositories.ingredientRepository, lockingService)
	inventoryServiceWorkUnit := product.InventoryServiceWorkUnit{
		InventoryRepo:  repositories.inventoryRepository,
		IngredientRepo: repositories.ingredientRepository,
		LockingService: lockingService,
		UnitService:    unitService,
	}
	inventoryService := product.NewInventoryService(inventoryServiceWorkUnit)
	productService := product.NewProductService(repositories.productRepository)
	s.services = systemServices{
		userService:       userService,
		permissionService: permissionService,
		roleService:       roleService,
		unitService:       unitService,
		warehouseService:  warehouseService,
		ingredientService: ingredientService,
		lockingService:    lockingService,
		inventoryService:  inventoryService,
		productService:    productService,
	}
}

func cleanUp() {
	connections.dbPool.Close()
	connections.redisClient.Close()
}
