package common

import (
	"encoding/json"
	"log"
	"net/http"
)

func (e ApiError) GenerateResponse() []byte {
	errorResponse, err := json.Marshal(e)
	if err != nil {
		internalServerError, _ := json.Marshal(NewInternalServerError())
		return internalServerError
	}
	return errorResponse
}

func Recover(f http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("internal error in commons package: %v", err)
				err := NewInternalServerError()
				WriteResponse[interface{}](Result[interface{}]{Error: err, Writer: w})
			}
		}()
		f.ServeHTTP(w, r)
	})
	return handler
}
