package narrative

import (
	"context"
	"fmt"
)

// Limited is a Generator decorator backed by a buffered-channel semaphore.
// It caps the number of in-flight model calls so bursts of recap requests do
// not overload the local Ollama process with unbounded concurrent inference.
type Limited struct {
	Primary Generator
	slots   chan struct{}
}

func NewLimited(primary Generator, maxConcurrent int) (*Limited, error) {
	if primary == nil {
		return nil, fmt.Errorf("narrative limiter: primary generator is required")
	}
	if maxConcurrent <= 0 {
		return nil, fmt.Errorf("narrative limiter: max concurrency must be positive")
	}
	return &Limited{
		Primary: primary,
		slots:   make(chan struct{}, maxConcurrent),
	}, nil
}

func (g *Limited) Generate(ctx context.Context, facts Facts) (Story, error) {
	if g == nil || g.Primary == nil {
		return Story{}, nil
	}
	if err := ctx.Err(); err != nil {
		return Story{}, err
	}

	select {
	case g.slots <- struct{}{}:
		defer func() { <-g.slots }()
	case <-ctx.Done():
		return Story{}, ctx.Err()
	}

	return g.Primary.Generate(ctx, facts)
}
