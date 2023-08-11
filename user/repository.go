package user

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

type IUserRepository interface {
}

type UserRepository struct {
	*pgxpool.Pool
}

func NewUserRepository(dbPool *pgxpool.Pool) IUserRepository {
	return &UserRepository{Pool: dbPool}
}
