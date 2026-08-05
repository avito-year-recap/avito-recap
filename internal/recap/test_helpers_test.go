package recap

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	testProfileID       = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testRecapID         = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	testShareID         = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	testDraftListingID  = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	testActiveListingID = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	testDialogID        = uuid.MustParse("55555555-5555-4555-8555-555555555555")
	otherProfileID      = uuid.MustParse("22222222-2222-4222-8222-222222222222")
)

type profileStorageStub struct {
	mu       sync.Mutex
	profiles []Profile
	profile  Profile
	listErr  error
	getErr   error
	gotID    uuid.UUID
}

func (s *profileStorageStub) ListProfiles(context.Context) ([]Profile, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.profiles, nil
}
func (s *profileStorageStub) GetProfile(_ context.Context, id uuid.UUID) (Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gotID = id
	if s.getErr != nil {
		return Profile{}, s.getErr
	}
	return s.profile, nil
}

type analyticsStorageStub struct {
	mu        sync.Mutex
	metrics   Metrics
	err       error
	gotID     uuid.UUID
	gotPeriod RecapPeriod
	calls     int
}

func (s *analyticsStorageStub) CalculateMetrics(_ context.Context, id uuid.UUID, period RecapPeriod) (Metrics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.gotID = id
	s.gotPeriod = period
	if s.err != nil {
		return Metrics{}, s.err
	}
	return s.metrics, nil
}

type actionStateStorageStub struct {
	mu      sync.Mutex
	state   ActionableState
	err     error
	gotID   uuid.UUID
	gotAsOf time.Time
	calls   int
}

func (s *actionStateStorageStub) GetActionableState(_ context.Context, id uuid.UUID, asOf time.Time) (ActionableState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.gotID = id
	s.gotAsOf = asOf
	if s.err != nil {
		return ActionableState{}, s.err
	}
	return s.state, nil
}

type recapStorageStub struct {
	mu          sync.Mutex
	value       Recap
	byKey       Recap
	saved       *Recap
	getByKeyErr error
	createErr   error
	getErr      error
	getShareErr error
	gotKey      RecapKey
	gotRecapID  uuid.UUID
	gotShareID  uuid.UUID
	createCalls int
}

func (s *recapStorageStub) GetRecapByKey(_ context.Context, key RecapKey) (Recap, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gotKey = key
	if s.getByKeyErr != nil {
		return Recap{}, s.getByKeyErr
	}
	if s.byKey.ID != uuid.Nil {
		return s.byKey, nil
	}
	if s.saved != nil && s.saved.Key() == key {
		return *s.saved, nil
	}
	return Recap{}, ErrRecapNotFound
}
func (s *recapStorageStub) CreateRecapIfAbsent(_ context.Context, key RecapKey, value Recap) (Recap, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.createCalls++
	s.gotKey = key
	if s.createErr != nil {
		return Recap{}, s.createErr
	}
	if s.byKey.ID != uuid.Nil {
		return s.byKey, nil
	}
	if s.saved != nil {
		return *s.saved, nil
	}
	copyValue := value
	s.saved = &copyValue
	return copyValue, nil
}
func (s *recapStorageStub) GetRecap(_ context.Context, id uuid.UUID) (Recap, error) {
	s.gotRecapID = id
	if s.getErr != nil {
		return Recap{}, s.getErr
	}
	if s.value.ID != uuid.Nil {
		return s.value, nil
	}
	if s.saved != nil {
		return *s.saved, nil
	}
	return Recap{}, ErrRecapNotFound
}
func (s *recapStorageStub) GetRecapByShareID(_ context.Context, id uuid.UUID) (Recap, error) {
	s.gotShareID = id
	if s.getShareErr != nil {
		return Recap{}, s.getShareErr
	}
	if s.value.ShareID != uuid.Nil {
		return s.value, nil
	}
	if s.saved != nil {
		return *s.saved, nil
	}
	return Recap{}, ErrRecapNotFound
}

func validProfile() Profile {
	return Profile{ID: testProfileID, Code: "active-buyer", DisplayName: "Алексей", Description: "Тестовый профиль"}
}
func validMetrics() Metrics {
	return Metrics{TotalEvents: 243, Searches: 20, TotalViews: 180, UniqueListings: 130, RepeatedViews: 50, FavoritesAdded: 30, ChatsStarted: 3, ChatsWithPurchase: 1, PurchasesCompleted: 1, ActiveDays: 45, CategoriesCount: 4, TopCategoryCode: "electronics", TopCategory: "Электроника", TopCategoryViews: 80, TopCategoryShareable: true, MostActiveMonth: 10}
}
func validActionableState() ActionableState {
	return ActionableState{CapturedAt: fixedClock(), FavoritesCount: 5, HasEverPublishedListing: true}
}
func validPeriod() RecapPeriod {
	p, err := completedYearPeriod(2025, fixedClock())
	if err != nil {
		panic(err)
	}
	return p
}
func validRecap() Recap {
	profile := validProfile()
	metrics := EnrichMetrics(validMetrics())
	state := validActionableState()
	ruleset := DefaultRuleset()
	behavior := ruleset.DetectBehavior(metrics)
	achievements := ruleset.BuildAchievements(metrics)
	action := ruleset.BuildNextAction(metrics, state, behavior)
	return Recap{ID: testRecapID, ShareID: testShareID, Profile: profile, Year: 2025, Period: validPeriod(), RulesVersion: ruleset.Version, Metrics: metrics, ActionableState: state, Behavior: behavior, Achievements: achievements, Cards: BuildCards(profile, 2025, testShareID, metrics, behavior, achievements, action), NextAction: action, GeneratedAt: fixedClock()}
}

func mustService(t *testing.T, profiles ProfileStorage, analytics AnalyticsStorage, recaps RecapStorage, options ...Option) *Service {
	t.Helper()
	return mustServiceWithState(t, profiles, analytics, &actionStateStorageStub{state: validActionableState()}, recaps, options...)
}
func mustServiceWithState(t *testing.T, profiles ProfileStorage, analytics AnalyticsStorage, states ActionStateStorage, recaps RecapStorage, options ...Option) *Service {
	t.Helper()
	service, err := NewService(profiles, analytics, states, recaps, options...)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return service
}
func fixedClock() time.Time { return time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC) }
