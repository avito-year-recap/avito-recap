package application

import (
	"errors"

	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/validation/structural"
)

var (
	ErrInvalidProfileID  = errors.New("invalid profile id")
	ErrInvalidRecapID    = errors.New("invalid recap id")
	ErrInvalidShareID    = errors.New("invalid share id")
	ErrNotEnoughActivity = errors.New("not enough activity to generate recap")
	ErrMissingDependency = errors.New("missing service dependency")
	ErrGenerateID        = errors.New("generate recap id")
	ErrProfileIDMismatch = errors.New("profile storage returned another profile")
	ErrRecapIDMismatch   = errors.New("recap storage returned another recap")
	ErrShareIDMismatch   = errors.New("recap storage returned another share id")
	ErrRecapKeyMismatch  = errors.New("recap storage returned another idempotency key")
	ErrRecapNotFound     = errors.New("recap not found")

	ErrInvalidYear            = analytics.ErrInvalidYear
	ErrYearNotComplete        = analytics.ErrYearNotComplete
	ErrInvalidProfile         = structural.ErrInvalidProfile
	ErrInvalidMetrics         = structural.ErrInvalidMetrics
	ErrInvalidActionableState = structural.ErrInvalidActionableState
	ErrInvalidRecap           = structural.ErrInvalidRecap
)
