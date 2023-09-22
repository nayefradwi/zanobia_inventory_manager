package common

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
)

// later I can have a nice way of passing the connection url
func ConnectDatabasePool(ctx context.Context, connectionUrl string) *pgxpool.Pool {
	dbPool, err := pgxpool.Connect(ctx, connectionUrl)
	if err != nil {
		log.Fatalf("failed to set up db connection: %s", err)
	}
	log.Print("connected to database successfully")
	return dbPool
}

func ConnectRedis(ctx context.Context, connectionUrl string) *redis.Client {
	opt, parsingErr := redis.ParseURL(connectionUrl)
	if parsingErr != nil {
		log.Fatalf("failed to parse redis connection url: %s", parsingErr)
	}
	opt.MaxRetries = 5
	redisClient := redis.NewClient(opt)
	_, connectionErr := redisClient.Ping(ctx).Result()
	if connectionErr != nil {
		log.Fatalf("failed to set up redis connection: %s", connectionErr)
	}
	log.Print("connected to redis successfully")
	return redisClient
}
