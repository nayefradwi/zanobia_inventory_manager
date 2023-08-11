package zimutils

import (
	"github.com/jackc/pgconn"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

const (
	duplicateErrorCode = "DUPLICATE"
)

func GetErrorCodeFromError(err *pgconn.PgError) string {
	switch err.Code {
	case "23505":
		return duplicateErrorCode
	default:
		return common.NOT_FOUND_CODE
	}
}
