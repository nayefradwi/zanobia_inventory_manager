package common

func newError(message string, status int) ApiError {
	return ApiError{
		Message: message,
		Status:  status,
		Code:    status,
	}
}

func NewCustomError(message string, status int, code int, errors ...ErrorDetails) ApiError {
	return ApiError{
		Message: message,
		Status:  status,
		Code:    code,
		Errors:  errors,
	}
}

func NewUnAuthorizedError(message string) ApiError {
	return newError(message, UNAUTHORIZED)
}

func NewInternalServerError(message string) ApiError {
	return newError(message, INTERNAL_SERVER_ERROR)
}

func NewNotFoundError(message string) ApiError {
	return newError(message, NOT_FOUND)
}

func NewBadRequestError(message string, code int) ApiError {
	return newError(message, BAD_REQUEST)
}

func NewForbiddenError(message string, code int) ApiError {
	return newError(message, FORBIDDEN)
}

func NewValidationError(message string, errors ...ErrorDetails) ApiError {
	return NewCustomError(message, BAD_REQUEST, 0, errors...)
}

func GenerateErrorFromStatus(status int) ApiError {
	switch status {
	case UNAUTHORIZED:
		return NewUnAuthorizedError("Unauthorized")
	case NOT_FOUND:
		return NewNotFoundError("Not Found")
	case BAD_REQUEST:
		return NewBadRequestError("Bad Request", status)
	case FORBIDDEN:
		return NewForbiddenError("Forbidden", status)
	default:
		return NewInternalServerError("Internal Server Error")
	}
}
