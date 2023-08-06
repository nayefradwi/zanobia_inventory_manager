package common

type ApiError struct {
	Message string         `json:"message"`
	Status  int            `json:"status"`
	Code    int            `json:"code,omitempty"`
	Errors  []ErrorDetails `json:"errors,omitempty"`
}

type ErrorDetails struct {
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
	Field   string `json:"field,omitempty"`
}

func (e ApiError) Error() string {
	return e.Message
}
