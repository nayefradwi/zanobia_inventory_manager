package main

import (
	"net/http"
)

var RegisteredServiceProvider *ServiceProvider
var RegisteredApiConfig ApiConfig

func main() {
	RegisteredApiConfig = LoadEnv(DEV)
	RegisteredServiceProvider = &ServiceProvider{}
	RegisteredServiceProvider.initiate(RegisteredApiConfig)
	r := RegisterRoutes(RegisteredServiceProvider)
	defer RegisteredServiceProvider.cleanUp()
	http.ListenAndServe(RegisteredApiConfig.Host, r)
}
