package narrative_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/year-recap/internal/recap/narrative"
)

type trackingGenerator struct {
	active  atomic.Int32
	max     atomic.Int32
	calls   atomic.Int32
	release <-chan struct{}
}

func (g *trackingGenerator) Generate(context.Context, narrative.Facts) (narrative.Story, error) {
	g.calls.Add(1)
	active := g.active.Add(1)
	for {
		current := g.max.Load()
		if active <= current || g.max.CompareAndSwap(current, active) {
			break
		}
	}
	defer g.active.Add(-1)
	if g.release != nil {
		<-g.release
	}
	return narrative.Story{}, nil
}

func TestLimitedCapsConcurrentNarrativeCalls(t *testing.T) {
	const (
		workers = 10
		limit   = 2
	)
	release := make(chan struct{})
	primary := &trackingGenerator{release: release}
	limited, err := narrative.NewLimited(primary, limit)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if _, err := limited.Generate(context.Background(), narrative.Facts{}); err != nil {
				t.Errorf("Generate() error = %v", err)
			}
		}()
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) && primary.max.Load() < limit {
		time.Sleep(time.Millisecond)
	}
	if got := primary.max.Load(); got != limit {
		t.Fatalf("max concurrent provider calls before release = %d, want %d", got, limit)
	}
	if got := primary.calls.Load(); got != limit {
		t.Fatalf("provider calls before release = %d, want exactly %d admitted calls", got, limit)
	}
	close(release)
	wg.Wait()

	if got := primary.calls.Load(); got != workers {
		t.Fatalf("provider calls = %d, want %d", got, workers)
	}
	if got := primary.max.Load(); got > limit {
		t.Fatalf("max concurrent provider calls = %d, limit = %d", got, limit)
	}
	if got := primary.max.Load(); got != limit {
		t.Fatalf("max concurrent provider calls = %d, want limiter to reach %d", got, limit)
	}
}

type blockingGenerator struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func (g *blockingGenerator) Generate(ctx context.Context, _ narrative.Facts) (narrative.Story, error) {
	g.once.Do(func() { close(g.entered) })
	select {
	case <-g.release:
		return narrative.Story{}, nil
	case <-ctx.Done():
		return narrative.Story{}, ctx.Err()
	}
}

func TestLimitedWaitingCallerRespectsContextCancellation(t *testing.T) {
	primary := &blockingGenerator{entered: make(chan struct{}), release: make(chan struct{})}
	limited, err := narrative.NewLimited(primary, 1)
	if err != nil {
		t.Fatal(err)
	}

	firstDone := make(chan error, 1)
	go func() {
		_, err := limited.Generate(context.Background(), narrative.Facts{})
		firstDone <- err
	}()

	select {
	case <-primary.entered:
	case <-time.After(time.Second):
		t.Fatal("first call did not acquire semaphore")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := limited.Generate(ctx, narrative.Facts{}); err == nil {
		t.Fatal("expected canceled context while waiting for semaphore")
	}

	close(primary.release)
	if err := <-firstDone; err != nil {
		t.Fatalf("first call failed: %v", err)
	}
}

func TestNewLimitedRejectsInvalidLimit(t *testing.T) {
	if _, err := narrative.NewLimited(&trackingGenerator{}, 0); err == nil {
		t.Fatal("expected invalid concurrency limit error")
	}
}

func TestLimitedRejectsAlreadyCanceledCallerBeforeAcquiringFreeSlot(t *testing.T) {
	primary := &trackingGenerator{}
	limited, err := narrative.NewLimited(primary, 1)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := limited.Generate(ctx, narrative.Facts{}); err == nil {
		t.Fatal("expected canceled context")
	}
	if got := primary.calls.Load(); got != 0 {
		t.Fatalf("provider calls = %d, want 0 for pre-canceled context", got)
	}
}
