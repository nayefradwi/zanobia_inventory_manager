package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
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
}

type systemServices struct {
	userService       user.IUserService
	permissionService user.IPermissionService
	roleService       user.IRoleService
	unitService       product.IUnitService
	warehouseService  warehouse.IWarehouseService
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
	dbPool := connectDatabasePool(ctx, config.DbConnectionUrl)
	redisClient := connectRedis(ctx, config.RedisUrl)
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
	return systemRepositories{
		userRepository:       userRepo,
		permissionRepository: permssionRepo,
		roleRepository:       roleRepo,
		unitRepository:       unitRepo,
		warehouseRepository:  warehouseRepo,
	}
}

func (s *ServiceProvider) registerServices(repositories systemRepositories) {
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
	s.services = systemServices{
		userService:       userService,
		permissionService: permissionService,
		roleService:       roleService,
		unitService:       unitService,
		warehouseService:  warehouseService,
	}
}

func cleanUp() {
	connections.dbPool.Close()
	connections.redisClient.Close()
}

func connectDatabasePool(ctx context.Context, connectionUrl string) *pgxpool.Pool {
	dbPool, err := pgxpool.Connect(ctx, connectionUrl)
	if err != nil {
		log.Fatalf("failed to set up db connection: %s", err)
	}
	log.Print("connected to database successfully")
	return dbPool
}

func connectRedis(ctx context.Context, connectionUrl string) *redis.Client {
	opt, parsingErr := redis.ParseURL(connectionUrl)
	if parsingErr != nil {
		log.Fatalf("failed to parse redis connection url: %s", parsingErr)
	}
	redisClient := redis.NewClient(opt)
	_, connectionErr := redisClient.Ping(ctx).Result()
	if connectionErr != nil {
		log.Fatalf("failed to set up redis connection: %s", connectionErr)
	}
	log.Print("connected to redis successfully")
	return redisClient
}
