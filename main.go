package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/user"
)

var RegisteredServiceProvider *ServiceProvider
var RegisteredApiConfig ApiConfig

func main() {
	env := getEnvArgument()
	RegisteredApiConfig = LoadEnv(env)
	RegisteredServiceProvider = &ServiceProvider{}
	RegisteredServiceProvider.initiate(RegisteredApiConfig)
	r := RegisterRoutes(RegisteredServiceProvider)
	setUserIdExtractor()
	defer cleanUp()
	log.Printf("listening on: %s", RegisteredApiConfig.Host)
	http.ListenAndServe(RegisteredApiConfig.Host, r)
}

func getEnvArgument() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	return PROD
}

func setUserIdExtractor() {
	common.SetUserIdExtractor(func(ctx context.Context) int {
		if user, ok := ctx.Value(common.UserKey{}).(user.User); ok {
			return user.Id
		}
		return 0
	})
}
