package application_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/narrative"
	"github.com/year-recap/internal/recap/testkit"
)

func TestServiceGeneratesAndReusesFinalRecap(t *testing.T) {
	profiles := &testkit.ProfileStorage{Profile: testkit.Profile()}
	analytics := &testkit.AnalyticsStorage{Metrics: testkit.Metrics()}
	states := &testkit.ActionStateStorage{State: model.ActionableState{FavoritesCount: 5, HasEverPublishedListing: true}}
	recaps := &testkit.RecapStorage{}
	ids := []uuid.UUID{testkit.RecapID, testkit.ShareID}
	index := 0
	service, err := application.NewService(profiles, analytics, states, recaps,
		application.WithClock(testkit.Clock),
		application.WithIDGenerator(func() (uuid.UUID, error) { value := ids[index]; index++; return value, nil }),
	)
	if err != nil {
		t.Fatal(err)
	}
	first, err := service.Generate(context.Background(), testkit.ProfileID, 2025)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Generate(context.Background(), testkit.ProfileID, 2025)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || recaps.Creates != 1 || analytics.Calls != 1 {
		t.Fatalf("not idempotent: first=%s second=%s creates=%d analytics=%d", first.ID, second.ID, recaps.Creates, analytics.Calls)
	}
	if first.NextAction.Title != "Вернись к своим находкам" {
		t.Fatalf("unexpected action: %+v", first.NextAction)
	}
}

func TestServiceRejectsCurrentYearBeforeStorageCalls(t *testing.T) {
	service, err := application.NewService(&testkit.ProfileStorage{Profile: testkit.Profile()}, &testkit.AnalyticsStorage{Metrics: testkit.Metrics()}, &testkit.ActionStateStorage{}, &testkit.RecapStorage{}, application.WithClock(testkit.Clock))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Generate(context.Background(), testkit.ProfileID, uint32(testkit.Clock().Year())); err == nil {
		t.Fatal("expected incomplete year error")
	}
}

type narrativeStub struct{}

func (narrativeStub) Generate(_ context.Context, facts narrative.Facts) (narrative.Story, error) {
	cards := make([]narrative.CardNarrative, 0, len(facts.EditableCardIDs))
	for _, id := range facts.EditableCardIDs {
		description := "AI обновил описание карточки."
		if id == "intro" {
			description = "AI собрал агрегированные факты в короткую историю."
		}
		cards = append(cards, narrative.CardNarrative{ID: id, Description: description})
	}
	return narrative.Story{Cards: cards}, nil
}

func TestServiceAppliesNarrativeAfterDeterministicDerivation(t *testing.T) {
	profiles := &testkit.ProfileStorage{Profile: testkit.Profile()}
	analytics := &testkit.AnalyticsStorage{Metrics: testkit.Metrics()}
	states := &testkit.ActionStateStorage{State: model.ActionableState{FavoritesCount: 5, HasEverPublishedListing: true}}
	recaps := &testkit.RecapStorage{}
	ids := []uuid.UUID{testkit.RecapID, testkit.ShareID}
	index := 0
	service, err := application.NewService(profiles, analytics, states, recaps,
		application.WithClock(testkit.Clock),
		application.WithIDGenerator(func() (uuid.UUID, error) { value := ids[index]; index++; return value, nil }),
		application.WithNarrativeEnricher(narrative.BestEffort{Primary: narrativeStub{}}),
	)
	if err != nil {
		t.Fatal(err)
	}

	recap, err := service.Generate(context.Background(), testkit.ProfileID, 2025)
	if err != nil {
		t.Fatal(err)
	}
	if recap.Cards[0].Description != "AI собрал агрегированные факты в короткую историю." {
		t.Fatalf("intro description = %q", recap.Cards[0].Description)
	}
	if recap.Behavior.Code == "" || recap.NextAction.Code == "" {
		t.Fatal("narrative must not replace deterministic decisions")
	}

	// AI copy is part of the immutable stored recap presentation. Re-reading the
	// same idempotency key must validate and reuse it rather than reject it as a
	// forged deterministic card or call the generator again.
	reused, err := service.Generate(context.Background(), testkit.ProfileID, 2025)
	if err != nil {
		t.Fatalf("reuse AI-enriched recap: %v", err)
	}
	if reused.ID != recap.ID || reused.Cards[0].Description != recap.Cards[0].Description {
		t.Fatalf("AI-enriched recap was not reused: first=%+v second=%+v", recap.Cards[0], reused.Cards[0])
	}
	stored, err := service.Get(context.Background(), recap.ID)
	if err != nil {
		t.Fatalf("read AI-enriched recap: %v", err)
	}
	if stored.Cards[0].Description != recap.Cards[0].Description {
		t.Fatalf("stored narrative changed: %q vs %q", stored.Cards[0].Description, recap.Cards[0].Description)
	}
}

type blockingAnalytics struct {
	calls   atomic.Int32
	entered chan struct{}
	release chan struct{}
	once    sync.Once
	metrics model.Metrics
}

func (s *blockingAnalytics) CalculateMetrics(
	ctx context.Context,
	_ uuid.UUID,
	_ model.RecapPeriod,
) (model.Metrics, error) {
	s.calls.Add(1)
	s.once.Do(func() { close(s.entered) })
	select {
	case <-s.release:
		return s.metrics, nil
	case <-ctx.Done():
		return model.Metrics{}, ctx.Err()
	}
}

func TestServiceSingleflightDeduplicatesConcurrentGenerate(t *testing.T) {
	const workers = 12

	profiles := &testkit.ProfileStorage{Profile: testkit.Profile()}
	analytics := &blockingAnalytics{
		entered: make(chan struct{}),
		release: make(chan struct{}),
		metrics: testkit.Metrics(),
	}
	states := &testkit.ActionStateStorage{State: model.ActionableState{FavoritesCount: 5, HasEverPublishedListing: true}}
	recaps := &testkit.RecapStorage{}
	service, err := application.NewService(
		profiles,
		analytics,
		states,
		recaps,
		application.WithClock(testkit.Clock),
	)
	if err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	results := make(chan model.Recap, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			<-start
			recap, err := service.Generate(context.Background(), testkit.ProfileID, 2025)
			if err != nil {
				errs <- err
				return
			}
			results <- recap
		}()
	}
	close(start)

	select {
	case <-analytics.entered:
	case <-time.After(time.Second):
		t.Fatal("generation did not reach analytics")
	}

	// Keep the leader blocked long enough for the remaining callers to observe
	// the cache miss and join the same in-process flight.
	time.Sleep(40 * time.Millisecond)
	if got := analytics.calls.Load(); got != 1 {
		t.Fatalf("analytics calls while leader is blocked = %d, want 1", got)
	}
	close(analytics.release)
	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	var firstID uuid.UUID
	count := 0
	for recap := range results {
		count++
		if firstID == uuid.Nil {
			firstID = recap.ID
			continue
		}
		if recap.ID != firstID {
			t.Fatalf("singleflight returned different recap IDs: %s vs %s", firstID, recap.ID)
		}
	}
	if count != workers {
		t.Fatalf("result count = %d, want %d", count, workers)
	}
	if got := analytics.calls.Load(); got != 1 {
		t.Fatalf("analytics calls = %d, want 1", got)
	}
	if recaps.Creates != 1 {
		t.Fatalf("recap creates = %d, want 1", recaps.Creates)
	}
}
