package common

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	PROD    = "prod"
	DEV     = "dev"
	STAGING = "staging"
)

var ENV = PROD

func GetEnvArgument() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	return PROD
}

func LoadEnv() {
	env := GetEnvArgument()
	setEnv(env)
	envFileName := "." + ENV + ".env"
	log.Printf("loading environment: %s", envFileName)
	err := godotenv.Load(envFileName)
	if err != nil {
		log.Fatalf("failed to load environment: %s", err.Error())
	}
}

func setEnv(env string) {
	if env == DEV {
		ENV = DEV
	} else if env == STAGING {
		ENV = STAGING
	} else {
		ENV = PROD
	}
}

func IsDev() bool {
	return ENV == DEV
}

func IsProd() bool {
	return ENV == PROD
}

func IsStaging() bool {
	return ENV == STAGING
}

func ConfigEssentials() {
	LoadEnv()
	InitializeLogger()
}

func CleanUp() {
	if logger != nil {
		logger.Sync()
	}
}
