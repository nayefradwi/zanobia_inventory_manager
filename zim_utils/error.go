package zimutils

import (
	"errors"

	"github.com/jackc/pgconn"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

const (
	duplicateErrorCode = "DUPLICATE"
)

func GetErrorCodeFromError(err error) string {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return common.UNKNOWN_ERROR_CODE
	}
	switch pgErr.Code {
	case "23505":
		return duplicateErrorCode
	default:
		return common.UNKNOWN_ERROR_CODE
	}
}
