package main

import (
	"os"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"go.uber.org/zap"
)

type ApiConfig struct {
	Host                 string
	Port                 string
	DbConnectionUrl      string
	InitialSysAdminEmail string
	InitialSysAdminPass  string
	RedisUrl             string
	Secret               string
}

func LoadEnv() ApiConfig {
	return ApiConfig{
		Host:                 os.Getenv("HOST_ADDRESS"),
		DbConnectionUrl:      os.Getenv("DB_CONNECTION_URL"),
		InitialSysAdminEmail: os.Getenv("INITIAL_SYSTEM_ADMIN_EMAIL"),
		InitialSysAdminPass:  os.Getenv("INITIAL_SYSTEM_ADMIN_PASSWORD"),
		RedisUrl:             os.Getenv("REDIS_CACHE_URL"),
		Port:                 os.Getenv("PORT"),
		Secret:               os.Getenv("SECRET"),
	}
}

func (c ApiConfig) GetListeningAddress(defaultPort string) string {
	listeningAddress := c.Host + ":" + defaultPort
	if c.Port != "" {
		listeningAddress = c.Host + ":" + c.Port
	}
	common.GetLogger().Info("listening on", zap.String("host", listeningAddress))
	return listeningAddress
}
