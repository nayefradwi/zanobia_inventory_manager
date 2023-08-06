package common

import "encoding/json"

func (e ApiError) GenerateResponse() []byte {
	errorResponse, err := json.Marshal(e)
	if err != nil {
		internalServerError, _ := json.Marshal(NewInternalServerError())
		return internalServerError
	}
	return errorResponse
}
