package serverutils

import "errors"

var (
	ErrNotFound     = errors.New("the requested resource was not found")
	ErrUnauthorized = errors.New("you are not authorized to access this resource")
	ErrInvalidFile  = errors.New("invalid file type or corrupted content")
	ErrInternal     = errors.New("something went wrong on our end, please try again later")
	ErrBadRequest   = errors.New("the request could not be processed due to invalid input")
)
