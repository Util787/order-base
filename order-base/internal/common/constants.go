package common

import "time"

// use to make key in context
type ContextKey string

var DefaultTTL = time.Second * 30

const (
	MaxOrderIDLength = 50 // 50 in case uid needs to be modified and according to db tables
	MinOrderIDLength = 32
)
