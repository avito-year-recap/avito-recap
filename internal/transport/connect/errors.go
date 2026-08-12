package connect

import (
	"context"
	"errors"

	connectrpc "connectrpc.com/connect"
	"github.com/year-recap/internal/recap/application"
)

var errInvalidProjection = errors.New("invalid transport projection")

func transportError(err error) error {
	if err == nil {
		return nil
	}
	var connectErr *connectrpc.Error
	if errors.As(err, &connectErr) {
		return connectErr
	}
	switch {
	case errors.Is(err, context.Canceled):
		return connectrpc.NewError(connectrpc.CodeCanceled, err)
	case errors.Is(err, context.DeadlineExceeded):
		return connectrpc.NewError(connectrpc.CodeDeadlineExceeded, err)
	case errors.Is(err, application.ErrInvalidProfileID),
		errors.Is(err, application.ErrInvalidRecapID),
		errors.Is(err, application.ErrInvalidShareID),
		errors.Is(err, application.ErrInvalidYear):
		return connectrpc.NewError(connectrpc.CodeInvalidArgument, err)
	case errors.Is(err, application.ErrYearNotComplete),
		errors.Is(err, application.ErrNotEnoughActivity),
		errors.Is(err, application.ErrMetricsNotFound):
		return connectrpc.NewError(connectrpc.CodeFailedPrecondition, err)
	case errors.Is(err, application.ErrProfileNotFound),
		errors.Is(err, application.ErrRecapNotFound):
		return connectrpc.NewError(connectrpc.CodeNotFound, err)
	case errors.Is(err, errInvalidProjection),
		errors.Is(err, application.ErrInvalidProfile),
		errors.Is(err, application.ErrInvalidMetrics),
		errors.Is(err, application.ErrInvalidActionableState),
		errors.Is(err, application.ErrInvalidRecap),
		errors.Is(err, application.ErrProfileIDMismatch),
		errors.Is(err, application.ErrRecapIDMismatch),
		errors.Is(err, application.ErrShareIDMismatch),
		errors.Is(err, application.ErrRecapKeyMismatch):
		return connectrpc.NewError(connectrpc.CodeDataLoss, err)
	default:
		return connectrpc.NewError(connectrpc.CodeInternal, err)
	}
}
