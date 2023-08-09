package main

import (
	"net/http"
	"os"
)

var RegisteredServiceProvider *ServiceProvider
var RegisteredApiConfig ApiConfig

func main() {
	env := getEnvArgument()
	RegisteredApiConfig = LoadEnv(env)
	RegisteredServiceProvider = &ServiceProvider{}
	RegisteredServiceProvider.initiate(RegisteredApiConfig)
	r := RegisterRoutes(RegisteredServiceProvider)
	defer RegisteredServiceProvider.cleanUp()
	http.ListenAndServe(RegisteredApiConfig.Host, r)
}

func getEnvArgument() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	return PROD
}
