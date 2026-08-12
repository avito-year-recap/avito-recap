package application

import (
	"context"
	"sync"

	"github.com/year-recap/internal/recap/model"
)

// generationFlightGroup deduplicates concurrent recap generations for the same
// idempotency key inside one application process.
//
// The expensive function runs on a shared context instead of inheriting the
// cancellation of whichever HTTP request happened to become the leader. Each
// caller may still leave independently. The shared work is canceled only when
// every waiter has gone away. Cross-process deduplication is deliberately out
// of scope for this in-memory group; a multi-replica deployment needs a
// distributed lock or a storage layer with an atomic uniqueness guarantee.
type generationFlightGroup struct {
	mu    sync.Mutex
	calls map[model.RecapKey]*generationCall
}

type generationCall struct {
	done       chan struct{}
	ctx        context.Context
	cancel     context.CancelFunc
	waiters    int
	recap      model.Recap
	err        error
	panicValue any
}

func (g *generationFlightGroup) Do(
	ctx context.Context,
	key model.RecapKey,
	fn func(context.Context) (model.Recap, error),
) (model.Recap, error) {
	if err := ctx.Err(); err != nil {
		return model.Recap{}, err
	}

	g.mu.Lock()
	if g.calls == nil {
		g.calls = make(map[model.RecapKey]*generationCall)
	}
	if call, ok := g.calls[key]; ok {
		call.waiters++
		g.mu.Unlock()
		return g.await(ctx, key, call)
	}

	// Preserve request-scoped values (for example tracing metadata), but do not
	// let the first caller's cancellation/deadline poison other callers that are
	// waiting for the same recap. Caller cancellation is tracked explicitly by
	// await; once the last waiter leaves, the shared context is canceled.
	sharedCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	call := &generationCall{
		done:    make(chan struct{}),
		ctx:     sharedCtx,
		cancel:  cancel,
		waiters: 1,
	}
	g.calls[key] = call
	g.mu.Unlock()

	go g.run(key, call, fn)
	return g.await(ctx, key, call)
}

func (g *generationFlightGroup) run(
	key model.RecapKey,
	call *generationCall,
	fn func(context.Context) (model.Recap, error),
) {
	defer func() {
		if recovered := recover(); recovered != nil {
			call.panicValue = recovered
		}
		call.cancel()

		g.mu.Lock()
		// The entry may already have been removed when all callers canceled and
		// a newer flight for the same key may have started. Never delete that one.
		if current, ok := g.calls[key]; ok && current == call {
			delete(g.calls, key)
		}
		close(call.done)
		g.mu.Unlock()
	}()

	call.recap, call.err = fn(call.ctx)
}

func (g *generationFlightGroup) await(
	ctx context.Context,
	key model.RecapKey,
	call *generationCall,
) (model.Recap, error) {
	select {
	case <-call.done:
		if call.panicValue != nil {
			panic(call.panicValue)
		}
		return call.recap, call.err

	case <-ctx.Done():
		g.mu.Lock()
		if current, ok := g.calls[key]; ok && current == call {
			call.waiters--
			if call.waiters == 0 {
				// Remove the abandoned flight immediately so a future request does
				// not join work that has no remaining consumers.
				delete(g.calls, key)
				call.cancel()
			}
		}
		g.mu.Unlock()
		return model.Recap{}, ctx.Err()
	}
}
