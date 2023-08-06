package common

func newError(message string, status int) ApiError {
	return ApiError{
		Message: message,
		Status:  status,
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
	return newError(message, 401)
}

func NewInternalServerError(message string) ApiError {
	return newError(message, 500)
}

func NewNotFoundError(message string) ApiError {
	return newError(message, 404)
}

func NewBadRequestError(message string) ApiError {
	return newError(message, 400)
}

func NewValidationError(message string, errors ...ErrorDetails) ApiError {
	return NewCustomError(message, 400, 0, errors...)
}

func GenerateErrorFromStatus(status int) ApiError {
	switch status {
	case 401:
		return NewUnAuthorizedError("Unauthorized")
	case 404:
		return NewNotFoundError("Not Found")
	case 400:
		return NewBadRequestError("Bad Request")
	default:
		return NewInternalServerError("Internal Server Error")
	}
}
