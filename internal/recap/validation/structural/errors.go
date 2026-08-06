package structural

import (
	"errors"
)

var (
	ErrInvalidProfile         = errors.New("invalid profile")
	ErrInvalidMetrics         = errors.New("invalid metrics")
	ErrInvalidActionableState = errors.New("invalid actionable state")
	ErrInvalidRecap           = errors.New("invalid recap")
)
