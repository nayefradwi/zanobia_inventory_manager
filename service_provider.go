package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/user"
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
}

type systemServices struct {
	userService       user.IUserService
	permissionService user.IPermissionService
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
	return systemRepositories{
		userRepository:       userRepo,
		permissionRepository: permssionRepo,
	}
}
func (s *ServiceProvider) registerServices(repositories systemRepositories) {
	userService := user.NewUserService(repositories.userRepository)
	permissionService := user.NewPermissionService(repositories.permissionRepository)
	s.services = systemServices{
		userService:       userService,
		permissionService: permissionService,
	}
}

func (s *ServiceProvider) cleanUp() {
	connections.dbPool.Close()
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
