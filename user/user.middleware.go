package user

import (
	"context"
	"net/http"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"go.uber.org/zap"
)

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
			return
		}
		userId := claims["id"].(float64)
		user, err := m.userService.GetUserById(ctx, int(userId))
		if err != nil {
			err := common.NewUnAuthorizedError("Invalid user")
			common.WriteResponseFromError(w, err)
			return
		}
		if user.Id == 0 {
			err := common.NewUnAuthorizedError("Invalid user")
			common.WriteResponseFromError(w, err)
			return
		}
		logger := common.LoggerFromCtx(ctx)
		logger = logger.With(
			zap.Int("userId", user.Id),
			zap.String("name", user.FirstName+" "+user.LastName),
		)
		ctx = common.SetLoggerToCtx(ctx, logger)
		ctx = context.WithValue(ctx, common.UserKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
	return handler
}

func (m *UserMiddleware) HasPermissions(permissions ...string) func(next http.Handler) http.Handler {
	middleware := func(next http.Handler) http.Handler {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			user := GetUserFromContext(ctx)
			if user.Id == 0 {
				err := common.NewUnAuthorizedError("Invalid user")
				common.WriteResponseFromError(w, err)
				return
			}
			if user.HasPermission(SysAdminPermissionHandle) {
				next.ServeHTTP(w, r)
				return
			}
			for _, permission := range permissions {
				if !user.HasPermission(permission) {
					err := common.NewForbiddenError("User does not have permission", permission)
					common.WriteResponseFromError(w, err)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
		return handler
	}
	return middleware
}

func GetUserFromContext(ctx context.Context) User {
	user := ctx.Value(common.UserKey{})
	if user != nil {
		return user.(User)
	}
	return User{}
}
