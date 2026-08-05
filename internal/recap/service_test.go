package recap

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewServiceRejectsMissingDependenciesAndRuleset(t *testing.T) {
	profiles := &profileStorageStub{}
	analytics := &analyticsStorageStub{}
	states := &actionStateStorageStub{}
	recaps := &recapStorageStub{}
	tests := []struct {
		name string
		p    ProfileStorage
		a    AnalyticsStorage
		s    ActionStateStorage
		r    RecapStorage
	}{
		{name: "profiles", a: analytics, s: states, r: recaps}, {name: "analytics", p: profiles, s: states, r: recaps},
		{name: "state", p: profiles, a: analytics, r: recaps}, {name: "recaps", p: profiles, a: analytics, s: states},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewService(tt.p, tt.a, tt.s, tt.r)
			if service != nil || !errors.Is(err, ErrMissingDependency) {
				t.Fatalf("service=%v err=%v", service, err)
			}
		})
	}
	_, err := NewService(profiles, analytics, states, recaps, WithRuleset(Ruleset{}))
	if !errors.Is(err, ErrInvalidRuleset) {
		t.Fatalf("expected invalid ruleset, got %v", err)
	}
}

func TestServiceGenerateUsesCompletedCalendarYear(t *testing.T) {
	analytics := &analyticsStorageStub{metrics: validMetrics()}
	states := &actionStateStorageStub{state: validActionableState()}
	recaps := &recapStorageStub{}
	ids := []uuid.UUID{testRecapID, testShareID}
	var index int
	service := mustServiceWithState(t, &profileStorageStub{profile: validProfile()}, analytics, states, recaps,
		WithClock(func() time.Time { return fixedClock() }), WithIDGenerator(func() (uuid.UUID, error) { id := ids[index]; index++; return id, nil }))
	value, err := service.Generate(context.Background(), testProfileID, 2025)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !value.Period.Final || value.Period.Year != 2025 || !value.Period.StartAt.Equal(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)) || !value.Period.EndAt.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected period: %+v", value.Period)
	}
	if analytics.gotPeriod != value.Period {
		t.Fatalf("analytics period=%+v recap=%+v", analytics.gotPeriod, value.Period)
	}
	if value.ID != testRecapID || value.ShareID != testShareID || value.ID == value.ShareID {
		t.Fatalf("ids not separated: %+v", value)
	}
	if value.ActionableState.CapturedAt != fixedClock() {
		t.Fatalf("state timestamp=%v", value.ActionableState.CapturedAt)
	}
}

func TestServiceGenerateRejectsCurrentAndFutureYear(t *testing.T) {
	service := mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, &recapStorageStub{}, WithClock(func() time.Time { return fixedClock() }))
	if _, err := service.Generate(context.Background(), testProfileID, 2026); !errors.Is(err, ErrYearNotComplete) {
		t.Fatalf("current year: %v", err)
	}
	if _, err := service.Generate(context.Background(), testProfileID, 2027); !errors.Is(err, ErrInvalidYear) {
		t.Fatalf("future year: %v", err)
	}
}

func TestServiceGenerateIsIdempotentByProfileYearAndRulesVersion(t *testing.T) {
	analytics := &analyticsStorageStub{metrics: validMetrics()}
	states := &actionStateStorageStub{state: validActionableState()}
	recaps := &recapStorageStub{}
	var idCalls int
	ids := []uuid.UUID{testRecapID, testShareID}
	service := mustServiceWithState(t, &profileStorageStub{profile: validProfile()}, analytics, states, recaps,
		WithClock(func() time.Time { return fixedClock() }), WithIDGenerator(func() (uuid.UUID, error) { id := ids[idCalls]; idCalls++; return id, nil }))
	first, err := service.Generate(context.Background(), testProfileID, 2025)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Generate(context.Background(), testProfileID, 2025)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || first.ShareID != second.ShareID {
		t.Fatalf("non-idempotent ids: %s/%s vs %s/%s", first.ID, first.ShareID, second.ID, second.ShareID)
	}
	if analytics.calls != 1 || states.calls != 1 || recaps.createCalls != 1 || idCalls != 2 {
		t.Fatalf("analytics=%d states=%d create=%d ids=%d", analytics.calls, states.calls, recaps.createCalls, idCalls)
	}
	expected := RecapKey{ProfileID: testProfileID, Year: 2025, RulesVersion: CurrentRulesVersion, RulesDigest: DefaultRuleset().Digest()}
	if recaps.gotKey != expected {
		t.Fatalf("key=%+v want %+v", recaps.gotKey, expected)
	}
}

func TestServiceGenerateConcurrentCallsReturnOneStoredRecap(t *testing.T) {
	analytics := &analyticsStorageStub{metrics: validMetrics()}
	states := &actionStateStorageStub{state: validActionableState()}
	recaps := &recapStorageStub{}
	var seq atomic.Uint32
	generator := func() (uuid.UUID, error) {
		n := seq.Add(1)
		return uuid.MustParse(map[uint32]string{1: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", 2: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", 3: "cccccccc-cccc-4ccc-8ccc-cccccccccccc", 4: "dddddddd-dddd-4ddd-8ddd-dddddddddddd"}[n]), nil
	}
	service := mustServiceWithState(t, &profileStorageStub{profile: validProfile()}, analytics, states, recaps, WithClock(func() time.Time { return fixedClock() }), WithIDGenerator(generator))
	var wg sync.WaitGroup
	results := make(chan Recap, 2)
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, e := service.Generate(context.Background(), testProfileID, 2025)
			results <- v
			errs <- e
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent generate: %v", err)
		}
	}
	var values []Recap
	for v := range results {
		values = append(values, v)
	}
	if len(values) != 2 || values[0].ID != values[1].ID || values[0].ShareID != values[1].ShareID {
		t.Fatalf("concurrent results differ: %+v", values)
	}
	if recaps.createCalls < 1 || recaps.createCalls > 2 {
		t.Fatalf("create calls=%d", recaps.createCalls)
	}
}

func TestServiceGenerateUsesConfiguredRulesetVersionInKey(t *testing.T) {
	rules := DefaultRuleset()
	rules.Version = "3.3.0-test"
	recaps := &recapStorageStub{byKey: func() Recap {
		v := validRecap()
		v.RulesVersion = rules.Version
		v.RulesDigest = rules.Digest()
		return v
	}()}
	service := mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, recaps, WithRuleset(rules), WithClock(func() time.Time { return fixedClock() }))
	value, err := service.Generate(context.Background(), testProfileID, 2025)
	if err != nil {
		t.Fatal(err)
	}
	if value.RulesVersion != rules.Version || recaps.gotKey.RulesVersion != rules.Version {
		t.Fatalf("rules version not used: %+v %+v", value, recaps.gotKey)
	}
}

func TestServiceGenerateActionComesFromCurrentState(t *testing.T) {
	state := validActionableState()
	state.CurrentDrafts = 1
	state.DraftListingID = testDraftListingID
	service := mustServiceWithState(t, &profileStorageStub{profile: validProfile()}, &analyticsStorageStub{metrics: validMetrics()}, &actionStateStorageStub{state: state}, &recapStorageStub{},
		WithClock(func() time.Time { return fixedClock() }), WithIDGenerator(sequenceIDGenerator(testRecapID, testShareID)))
	value, err := service.Generate(context.Background(), testProfileID, 2025)
	if err != nil {
		t.Fatal(err)
	}
	if value.NextAction.Code != ActionFinishDraft || value.NextAction.Target.Listing == nil || value.NextAction.Target.Listing.ListingID != testDraftListingID {
		t.Fatalf("action not state-backed: %+v", value.NextAction)
	}
}

func TestServiceGetUsesInternalIDAndShareUsesPublicID(t *testing.T) {
	stored := validRecap()
	recaps := &recapStorageStub{value: stored}
	service := mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, recaps)
	value, err := service.Get(context.Background(), testRecapID)
	if err != nil {
		t.Fatal(err)
	}
	if value.ID != testRecapID || recaps.gotRecapID != testRecapID {
		t.Fatalf("internal get mismatch")
	}
	card, err := service.GetShareCard(context.Background(), testShareID)
	if err != nil {
		t.Fatal(err)
	}
	if card.ShareID != testShareID || recaps.gotShareID != testShareID {
		t.Fatalf("share get mismatch: %+v", card)
	}
	if strings.Contains(strings.ToLower(card.BehaviorTitle), testRecapID.String()) {
		t.Fatal("internal id leaked")
	}
}

func TestServiceGetRejectsIdentifierMismatches(t *testing.T) {
	wrong := validRecap()
	wrong.ID = otherProfileID
	service := mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, &recapStorageStub{value: wrong})
	if _, err := service.Get(context.Background(), testRecapID); !errors.Is(err, ErrRecapIDMismatch) {
		t.Fatalf("internal mismatch: %v", err)
	}
	wrong = validRecap()
	wrong.ShareID = otherProfileID
	service = mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, &recapStorageStub{value: wrong})
	if _, err := service.GetShareCard(context.Background(), testShareID); !errors.Is(err, ErrShareIDMismatch) {
		t.Fatalf("share mismatch: %v", err)
	}
}

func TestServiceGenerateNormalizesStoredStrings(t *testing.T) {
	profile := validProfile()
	profile.Code = " active-buyer "
	profile.DisplayName = " Алексей "
	metrics := validMetrics()
	metrics.TopCategoryCode = " electronics "
	metrics.TopCategory = " Электроника\n"
	service := mustService(t, &profileStorageStub{profile: profile}, &analyticsStorageStub{metrics: metrics}, &recapStorageStub{}, WithClock(func() time.Time { return fixedClock() }), WithIDGenerator(sequenceIDGenerator(testRecapID, testShareID)))
	value, err := service.Generate(context.Background(), testProfileID, 2025)
	if err != nil {
		t.Fatal(err)
	}
	if value.Profile.Code != "active-buyer" || value.Metrics.TopCategoryCode != "electronics" || value.Metrics.TopCategory != "Электроника" {
		t.Fatalf("not normalized: %+v", value)
	}
}

func TestGenerateID(t *testing.T) {
	first, err := generateID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := generateID()
	if err != nil {
		t.Fatal(err)
	}
	if first == uuid.Nil || second == uuid.Nil || first == second || first.Version() != 4 || first.Variant() != uuid.RFC4122 {
		t.Fatalf("bad ids: %s %s", first, second)
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }
func TestGenerateIDFromReaderError(t *testing.T) {
	source := errors.New("reader failed")
	value, err := generateIDFrom(failingReader{source})
	if value != uuid.Nil || !errors.Is(err, source) {
		t.Fatalf("value=%s err=%v", value, err)
	}
}

func sequenceIDGenerator(ids ...uuid.UUID) IDGenerator {
	index := 0
	return func() (uuid.UUID, error) {
		if index >= len(ids) {
			return uuid.Nil, errors.New("no ids")
		}
		id := ids[index]
		index++
		return id, nil
	}
}

func TestServiceRejectsCurrentYearBeforeCallingDependencies(t *testing.T) {
	profiles := &profileStorageStub{profile: validProfile()}
	analytics := &analyticsStorageStub{metrics: validMetrics()}
	states := &actionStateStorageStub{state: validActionableState()}
	recaps := &recapStorageStub{}
	service := mustServiceWithState(t, profiles, analytics, states, recaps, WithClock(fixedClock))

	_, err := service.Generate(context.Background(), testProfileID, 2026)
	if !errors.Is(err, ErrYearNotComplete) {
		t.Fatalf("Generate current year error = %v, want %v", err, ErrYearNotComplete)
	}
	if profiles.gotID != uuid.Nil || analytics.calls != 0 || states.calls != 0 || recaps.gotKey != (RecapKey{}) || recaps.createCalls != 0 {
		t.Fatalf("dependencies were called for unfinished year: profile=%s analytics=%d states=%d key=%+v creates=%d",
			profiles.gotID, analytics.calls, states.calls, recaps.gotKey, recaps.createCalls)
	}
}

func TestServiceGetRejectsCorruptStoredRecap(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Recap)
	}{
		{name: "stale rate", mutate: func(value *Recap) { value.Metrics.RepeatRate = 0.999 }},
		{name: "duplicate card id", mutate: func(value *Recap) { value.Cards[1].ID = value.Cards[0].ID }},
		{name: "broken position", mutate: func(value *Recap) { value.Cards[1].Position = 99 }},
		{name: "missing action target", mutate: func(value *Recap) { value.NextAction.Target = ActionTarget{} }},
		{name: "wrong payload type", mutate: func(value *Recap) { value.Cards[1].Payload = ActiveMonthPayload{Month: 1} }},
		{name: "blank user text", mutate: func(value *Recap) { value.Behavior.Title = "   " }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stored := validRecap()
			test.mutate(&stored)
			service := mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, &recapStorageStub{value: stored})
			_, err := service.Get(context.Background(), testRecapID)
			if !errors.Is(err, ErrInvalidRecap) {
				t.Fatalf("Get corrupt recap error = %v, want %v", err, ErrInvalidRecap)
			}
		})
	}
}

func TestServiceGetShareCardRejectsCorruptStoredRecap(t *testing.T) {
	stored := validRecap()
	stored.Cards[1].ID = stored.Cards[0].ID
	service := mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, &recapStorageStub{value: stored})
	_, err := service.GetShareCard(context.Background(), testShareID)
	if !errors.Is(err, ErrInvalidRecap) {
		t.Fatalf("GetShareCard corrupt recap error = %v, want %v", err, ErrInvalidRecap)
	}
}

func TestServiceReturnsOnlyRecapsThatPassValidation(t *testing.T) {
	stored := validRecap()
	service := mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, &recapStorageStub{value: stored})
	value, err := service.Get(context.Background(), testRecapID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if err := validateRecap(value); err != nil {
		t.Fatalf("service returned recap that does not pass validation: %v", err)
	}
}

func TestServiceGenerateRejectsCorruptExistingRecapByKey(t *testing.T) {
	stored := validRecap()
	stored.NextAction.Target = ActionTarget{}
	recaps := &recapStorageStub{byKey: stored}
	analytics := &analyticsStorageStub{metrics: validMetrics()}
	states := &actionStateStorageStub{state: validActionableState()}
	service := mustServiceWithState(t, &profileStorageStub{profile: validProfile()}, analytics, states, recaps, WithClock(fixedClock))

	_, err := service.Generate(context.Background(), testProfileID, 2025)
	if !errors.Is(err, ErrInvalidRecap) {
		t.Fatalf("Generate with corrupt existing recap error = %v, want %v", err, ErrInvalidRecap)
	}
	if analytics.calls != 0 || states.calls != 0 || recaps.createCalls != 0 {
		t.Fatalf("corrupt existing recap must fail before regeneration: analytics=%d states=%d creates=%d", analytics.calls, states.calls, recaps.createCalls)
	}
}
