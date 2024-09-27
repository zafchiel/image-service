package errors

import "errors"

var (
	ErrEmailInUse    = errors.New("email already in use")
	ErrImageNotFound = errors.New("image not found")
	ErrInvalidID     = errors.New("invalid ID")
	ErrFileTooLarge  = errors.New("file too large")
	ErrInvalidFormat = errors.New("invalid image format")
)
