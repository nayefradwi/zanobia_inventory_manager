package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	PROD = "prod"
	DEV  = "dev"
)

var ENV = PROD

type ApiConfig struct {
	Host            string
	DbConnectionUrl string
}

func LoadEnv(env string) ApiConfig {
	if env == DEV {
		ENV = DEV
	}
	envFileName := "." + ENV + ".env"
	log.Printf("loading environment: %s", envFileName)
	err := godotenv.Load(envFileName)
	if err != nil {
		log.Fatalf("failed to load environment: %s", err.Error())
	}
	return ApiConfig{
		Host:            os.Getenv("HOST_ADDRESS"),
		DbConnectionUrl: os.Getenv("DB_CONNECTION_URL"),
	}
}
