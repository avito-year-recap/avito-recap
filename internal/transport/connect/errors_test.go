package connect

import (
	"context"
	"errors"
	"fmt"
	"testing"

	connectrpc "connectrpc.com/connect"
	"github.com/year-recap/internal/recap/application"
)

func TestTransportErrorMapsApplicationErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code connectrpc.Code
	}{
		{name: "invalid profile id", err: application.ErrInvalidProfileID, code: connectrpc.CodeInvalidArgument},
		{name: "invalid recap id", err: application.ErrInvalidRecapID, code: connectrpc.CodeInvalidArgument},
		{name: "invalid share id", err: application.ErrInvalidShareID, code: connectrpc.CodeInvalidArgument},
		{name: "invalid year", err: application.ErrInvalidYear, code: connectrpc.CodeInvalidArgument},
		{name: "unfinished year", err: application.ErrYearNotComplete, code: connectrpc.CodeFailedPrecondition},
		{name: "not enough activity", err: application.ErrNotEnoughActivity, code: connectrpc.CodeFailedPrecondition},
		{name: "metrics not found", err: application.ErrMetricsNotFound, code: connectrpc.CodeFailedPrecondition},
		{name: "profile not found", err: application.ErrProfileNotFound, code: connectrpc.CodeNotFound},
		{name: "recap not found", err: application.ErrRecapNotFound, code: connectrpc.CodeNotFound},
		{name: "invalid recap", err: application.ErrInvalidRecap, code: connectrpc.CodeDataLoss},
		{name: "projection", err: errInvalidProjection, code: connectrpc.CodeDataLoss},
		{name: "canceled", err: context.Canceled, code: connectrpc.CodeCanceled},
		{name: "deadline", err: context.DeadlineExceeded, code: connectrpc.CodeDeadlineExceeded},
		{name: "unknown", err: errors.New("storage unavailable"), code: connectrpc.CodeInternal},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := transportError(fmt.Errorf("wrapped: %w", test.err))
			if code := connectrpc.CodeOf(actual); code != test.code {
				t.Fatalf("code = %s, want %s: %v", code, test.code, actual)
			}
		})
	}
}

func TestTransportErrorPreservesConnectError(t *testing.T) {
	source := connectrpc.NewError(connectrpc.CodePermissionDenied, errors.New("denied"))
	if actual := transportError(source); !errors.Is(actual, source) {
		t.Fatalf("connect error was replaced: %v", actual)
	}
}
