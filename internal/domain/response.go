package domain

type APIResponse[T any] struct {
	Data  *T        `json:"data"`
	Error *APIError `json:"error"`
}

type APIError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func Ok[T any](data T) APIResponse[T] {
	return APIResponse[T]{Data: &data, Error: nil}
}

func Err(message, code string) APIResponse[any] {
	return APIResponse[any]{Data: nil, Error: &APIError{Message: message, Code: code}}
}
