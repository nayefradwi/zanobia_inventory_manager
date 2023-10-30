package user

import (
	"context"
	"net/http"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type UserKey struct{}
type UserMiddleware struct {
	userService IUserService
	// TODO: add cache service
}

func NewUserMiddleware(userService IUserService) UserMiddleware {
	return UserMiddleware{
		userService: userService,
	}
}

func (m *UserMiddleware) SetUserFromHeader(next http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims := common.GetClaimsFromContext(ctx)
		if claims == nil {
			err := common.NewUnAuthorizedError("Invalid user claims")
			common.WriteResponseFromError(w, err)
		}
		userId := claims["id"].(float64)
		user, err := m.userService.GetUserById(ctx, int(userId))
		if err != nil {
			err := common.NewUnAuthorizedError("Invalid user claims")
			common.WriteResponseFromError(w, err)
		}
		ctx = context.WithValue(ctx, UserKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
	return handler
}

func GetUserFromContext(ctx context.Context) User {
	user := ctx.Value(UserKey{})
	if user != nil {
		return user.(User)
	}
	return User{}
}
