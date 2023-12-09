package common

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

func GetEnvArgument() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	return PROD
}

func LoadEnv() {
	env := GetEnvArgument()
	if env == DEV {
		ENV = DEV
	}
	envFileName := "." + ENV + ".env"
	log.Printf("loading environment: %s", envFileName)
	err := godotenv.Load(envFileName)
	if err != nil {
		log.Fatalf("failed to load environment: %s", err.Error())
	}
}

func IsDev() bool {
	return ENV == DEV
}

func IsProd() bool {
	return ENV == PROD
}

func ConfigEssentials() {
	LoadEnv()
	InitializeLogger()
}

func CleanUp() {
	if Logger != nil {
		Logger.Sync()
	}
}
