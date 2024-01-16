package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/user"
)

var RegisteredServiceProvider *ServiceProvider
var RegisteredApiConfig ApiConfig

func main() {
	r := setUp()
	defer cleanUp()
	http.ListenAndServe(RegisteredApiConfig.GetListeningAddress("3000"), r)
}

func setUserIdExtractor() {
	common.SetUserIdExtractor(func(ctx context.Context) int {
		if user, ok := ctx.Value(common.UserKey{}).(user.User); ok {
			return user.Id
		}
		return 0
	})
}

func setUp() chi.Router {
	common.ConfigEssentials()
	RegisteredApiConfig = LoadEnv()
	common.SetSecret(RegisteredApiConfig.Secret)
	RegisteredServiceProvider = &ServiceProvider{}
	RegisteredServiceProvider.initiate(RegisteredApiConfig)
	setUserIdExtractor()
	r := RegisterRoutes(RegisteredServiceProvider)
	return r
}
