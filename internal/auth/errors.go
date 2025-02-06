package auth

import "errors"

var (
	ErrTryAgain      = errors.New("try again")
	ErrAlreadyExists = errors.New("already exists")
)
