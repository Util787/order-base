package models

import (
	"errors"
	"fmt"
)

var (
	ErrOrdersNotFound = errors.New("orders not found")
)

// ErrValidation is an abstraction that should be used only to get right status code in handlers
//
// Any validation error should contain this abstraction
var ErrValidation = errors.New("validation error")
var (
	ErrInvalidOrderId = fmt.Errorf("%w: invalid order id", ErrValidation)
)
