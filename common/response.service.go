package common

import (
	"encoding/json"
	"log"
	"net/http"
)

type EmptyResult struct {
	Writer  http.ResponseWriter
	Message string
	Error   error
}

type Result[T any] struct {
	Writer http.ResponseWriter
	Data   T
	Error  error
}

/*
example to use this is:

	ParseBody[TYPE](w, r.Body, func(data TYPE){
		...DO SOMETHING WITH DATA
		WriteResponse[TYPE](Result[TYPE]{Writer: w, Data: data, Error: err})
	})

in this case if the error is nil it will write the body as json
if the error is an api error then the error will be written as json
this creates a consistent way of handling 3 cases:
  - api returns data
  - api returns success / created without data
  - api returns error

so the pattern is always:
  - parse body if needed
  - call service which calls repo
  - write response
*/
func WriteResponse[T any](result Result[T]) {
	if result.Error != nil {
		WriteResponseFromError(result.Error, result.Writer)
		return
	}
	json.NewEncoder(result.Writer).Encode(result.Data)
}

func WriteCreatedResponse(result EmptyResult) {
	if result.Error != nil {
		WriteResponseFromError(result.Error, result.Writer)
		return
	}
	writeResponse(result, 401)

}
func WriteEmptyResponse(result EmptyResult) {
	if result.Error != nil {
		WriteResponseFromError(result.Error, result.Writer)
		return
	}
	writeResponse(result, 400)
}

func writeResponse(result EmptyResult, status int) {
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
