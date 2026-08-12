package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
)

func testFlightKey() model.RecapKey {
	return model.RecapKey{
		ProfileID:    uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Year:         2025,
		RulesVersion: "test",
		RulesDigest:  "digest",
	}
}

func waitForFlightWaiters(t *testing.T, group *generationFlightGroup, key model.RecapKey, want int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		group.mu.Lock()
		call := group.calls[key]
		got := 0
		if call != nil {
			got = call.waiters
		}
		group.mu.Unlock()
		if got == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("flight waiters did not reach %d", want)
}

func TestGenerationFlightLeaderCancellationDoesNotPoisonFollower(t *testing.T) {
	var group generationFlightGroup
	key := testFlightKey()
	started := make(chan struct{})
	release := make(chan struct{})

	leaderCtx, cancelLeader := context.WithCancel(context.Background())
	leaderErr := make(chan error, 1)
	go func() {
		_, err := group.Do(leaderCtx, key, func(sharedCtx context.Context) (model.Recap, error) {
			close(started)
			select {
			case <-release:
				return model.Recap{ID: uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")}, nil
			case <-sharedCtx.Done():
				return model.Recap{}, sharedCtx.Err()
			}
		})
		leaderErr <- err
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("shared generation did not start")
	}

	followerResult := make(chan model.Recap, 1)
	followerErr := make(chan error, 1)
	duplicateExecuted := make(chan struct{}, 1)
	go func() {
		recap, err := group.Do(context.Background(), key, func(context.Context) (model.Recap, error) {
			duplicateExecuted <- struct{}{}
			return model.Recap{}, nil
		})
		if err != nil {
			followerErr <- err
			return
		}
		followerResult <- recap
	}()

	waitForFlightWaiters(t, &group, key, 2)
	cancelLeader()

	select {
	case err := <-leaderErr:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("leader error = %v, want context canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("canceled leader did not return")
	}

	close(release)
	select {
	case err := <-followerErr:
		t.Fatalf("follower failed after leader cancellation: %v", err)
	case recap := <-followerResult:
		if recap.ID == uuid.Nil {
			t.Fatal("follower received empty recap")
		}
	case <-time.After(time.Second):
		t.Fatal("follower did not receive shared result")
	}
	select {
	case <-duplicateExecuted:
		t.Fatal("follower executed duplicate generation")
	default:
	}
}

func TestGenerationFlightCancelsSharedWorkWhenLastWaiterLeaves(t *testing.T) {
	var group generationFlightGroup
	key := testFlightKey()
	sharedCanceled := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := group.Do(ctx, key, func(sharedCtx context.Context) (model.Recap, error) {
			<-sharedCtx.Done()
			close(sharedCanceled)
			return model.Recap{}, sharedCtx.Err()
		})
		done <- err
	}()

	waitForFlightWaiters(t, &group, key, 1)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("caller error = %v, want context canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("canceled caller did not return")
	}
	select {
	case <-sharedCanceled:
	case <-time.After(time.Second):
		t.Fatal("shared work was not canceled after last waiter left")
	}
}

func TestGenerationFlightPanicDoesNotLeaveStaleEntry(t *testing.T) {
	var group generationFlightGroup
	key := testFlightKey()

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected shared function panic to propagate")
			}
		}()
		_, _ = group.Do(context.Background(), key, func(context.Context) (model.Recap, error) {
			panic("boom")
		})
	}()

	recap, err := group.Do(context.Background(), key, func(context.Context) (model.Recap, error) {
		return model.Recap{ID: uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if recap.ID == uuid.Nil {
		t.Fatal("second flight did not execute after panic cleanup")
	}
}
