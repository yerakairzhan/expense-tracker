package apperror

import "net/http"

type Error struct {
	Code    string
	Message string
	Status  int
}

func (e *Error) Error() string {
	return e.Message
}

func New(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

func Validation(message string) *Error {
	return New(http.StatusBadRequest, "VALIDATION_ERROR", message)
}

func Unauthorized(message string) *Error {
	return New(http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func Forbidden(message string) *Error {
	return New(http.StatusForbidden, "FORBIDDEN", message)
}

func NotFound(message string) *Error {
	return New(http.StatusNotFound, "NOT_FOUND", message)
}

func Conflict(message string) *Error {
	return New(http.StatusConflict, "CONFLICT", message)
}

func InsufficientFunds(message string) *Error {
	return New(http.StatusUnprocessableEntity, "INSUFFICIENT_FUNDS", message)
}

func Internal(message string) *Error {
	return New(http.StatusInternalServerError, "INTERNAL_ERROR", message)
}
