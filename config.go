package main

import (
	"os"
)

type ApiConfig struct {
	Host                 string
	DbConnectionUrl      string
	InitialSysAdminEmail string
	InitialSysAdminPass  string
	RedisUrl             string
}

func LoadEnv() ApiConfig {
	return ApiConfig{
		Host:                 os.Getenv("HOST_ADDRESS"),
		DbConnectionUrl:      os.Getenv("DB_CONNECTION_URL"),
		InitialSysAdminEmail: os.Getenv("INITIAL_SYSTEM_ADMIN_EMAIL"),
		InitialSysAdminPass:  os.Getenv("INITIAL_SYSTEM_ADMIN_PASSWORD"),
		RedisUrl:             os.Getenv("REDIS_CACHE_URL"),
	}
}
