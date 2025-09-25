package go_clipper2

import "errors"

var (
	ErrPrecisionRange         = errors.New("precision is out of range")
	ErrInvalidRemoveListIndex = errors.New("invalid remove index from list")
)
