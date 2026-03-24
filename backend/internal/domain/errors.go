package domain

import "errors"

// Common domain errors that can occur across different services/repositories.
var (
	ErrNotFound       = errors.New("resource not found")
	ErrAlreadyExists  = errors.New("resource already exists")
	ErrInvalidInput   = errors.New("invalid input data")
	ErrUnauthorized   = errors.New("unauthorized access")
	ErrInternal       = errors.New("internal server error")
	ErrNotImplemented = errors.New("method not implemented")
	ErrChallengeFull  = errors.New("challenge is already full")
)
