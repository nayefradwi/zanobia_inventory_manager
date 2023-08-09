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
	Host string
}

func LoadEnv(env string) ApiConfig {
	if env == DEV {
		ENV = DEV
	}
	envFileName := "." + ENV + ".env"
	err := godotenv.Load(envFileName)
	if err != nil {
		log.Fatalf("failed to load environment: %s", err.Error())
	}
	return ApiConfig{
		Host: os.Getenv("HOST"),
	}
}
