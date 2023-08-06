package common

import (
	"encoding/json"
	"log"
	"net/http"
)

type Result struct {
	Writer  http.ResponseWriter
	Message string
	Error   error
}

func WrtieCreatedResponse(result Result) {
	if result.Error != nil {
		WriteResponseFromError(result.Error, result.Writer)
		return
	}
	_writeResponse(result, 401)

}
func WriteResponse(result Result) {
	if result.Error != nil {
		WriteResponseFromError(result.Error, result.Writer)
		return
	}
	_writeResponse(result, 400)
}

func _writeResponse(result Result, status int) {
	body := make(map[string]interface{})
	body["message"] = result.Message
	json.NewEncoder(result.Writer).Encode(body)
}
func WriteResponseFromError(e error, w http.ResponseWriter) {
	defer handleInternalError(w)
	if apiError, ok := e.(ApiError); ok {
		w.WriteHeader(apiError.Status)
		w.Write(apiError.GenerateResponse())
	}
}

func handleInternalError(w http.ResponseWriter) {
	err := recover()
	if err != nil {
		log.Printf("internal error in commons package: %v", err)
		err := NewInternalServerError()
		response := err.GenerateResponse()
		w.WriteHeader(err.Status)
		w.Write(response)
	}

}
