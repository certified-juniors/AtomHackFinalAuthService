package domain

import (
	"errors"
	"net/http"
)

const DateOutOfRangeErrCode = "23514"

var (
	ErrInternalServerError = errors.New("internal Server Error")
	ErrNotFound            = errors.New("requested Item is not found")
	ErrBadRequest          = errors.New("request is not valid")
	ErrUnauthorized        = errors.New("need to authorize")
	ErrWrongCredentials    = errors.New("username or password is invalid")
	ErrInvalidToken        = errors.New("session token is invalid")
	ErrAlreadyExists       = errors.New("resource already exists")
	ErrOutOfRange          = errors.New("id is out of range")
)

func GetStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	switch {
	case errors.Is(err, ErrWrongCredentials):
		return http.StatusBadRequest
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrInvalidToken):
		return http.StatusBadRequest
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrOutOfRange):
		return http.StatusNotFound
	case errors.Is(err, ErrAlreadyExists):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
