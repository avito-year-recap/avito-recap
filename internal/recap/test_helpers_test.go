package recap

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	testProfileID  = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testRecapID    = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	otherProfileID = uuid.MustParse("22222222-2222-4222-8222-222222222222")
)

type profileStorageStub struct {
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

func (s *profileStorageStub) GetProfile(_ context.Context, profileID uuid.UUID) (Profile, error) {
	s.gotID = profileID
	if s.getErr != nil {
		return Profile{}, s.getErr
	}
	return s.profile, nil
}

type analyticsStorageStub struct {
	metrics Metrics
	err     error
	gotID   uuid.UUID
	gotYear uint32
}

func (s *analyticsStorageStub) CalculateMetrics(
	_ context.Context,
	profileID uuid.UUID,
	year uint32,
) (Metrics, error) {
	s.gotID = profileID
	s.gotYear = year
	if s.err != nil {
		return Metrics{}, s.err
	}
	return s.metrics, nil
}

type recapStorageStub struct {
	saved      *Recap
	value      Recap
	saveErr    error
	getErr     error
	gotRecapID uuid.UUID
	saveCalls  int
}

func (s *recapStorageStub) SaveRecap(_ context.Context, value Recap) error {
	s.saveCalls++
	if s.saveErr != nil {
		return s.saveErr
	}
	copyValue := value
	s.saved = &copyValue
	return nil
}

func (s *recapStorageStub) GetRecap(_ context.Context, recapID uuid.UUID) (Recap, error) {
	s.gotRecapID = recapID
	if s.getErr != nil {
		return Recap{}, s.getErr
	}
	return s.value, nil
}

func validProfile() Profile {
	return Profile{
		ID:          testProfileID,
		Code:        "active-buyer",
		DisplayName: "Алексей",
		Description: "Тестовый профиль",
	}
}

func validMetrics() Metrics {
	return Metrics{
		TotalEvents:          243,
		Searches:             20,
		TotalViews:           180,
		UniqueListings:       130,
		RepeatedViews:        50,
		FavoritesAdded:       30,
		ChatsStarted:         3,
		ChatsWithPurchase:    1,
		PurchasesCompleted:   1,
		ActiveDays:           45,
		CategoriesCount:      4,
		TopCategoryCode:      "electronics",
		TopCategory:          "Электроника",
		TopCategoryViews:     80,
		TopCategoryShareable: true,
		MostActiveMonth:      10,
	}
}

func validRecap() Recap {
	profile := validProfile()
	metrics := EnrichMetrics(validMetrics())
	behavior := DetectBehavior(metrics)
	achievements := BuildAchievements(metrics)
	nextAction := BuildNextAction(metrics)

	return Recap{
		ID:           testRecapID,
		Profile:      profile,
		Year:         2025,
		RulesVersion: CurrentRulesVersion,
		Metrics:      metrics,
		Behavior:     behavior,
		Achievements: achievements,
		Cards:        BuildCards(profile, 2025, metrics, behavior, achievements, nextAction),
		NextAction:   nextAction,
		GeneratedAt:  fixedClock(),
	}
}

func mustService(
	t *testing.T,
	profiles ProfileStorage,
	analytics AnalyticsStorage,
	recaps RecapStorage,
	options ...Option,
) *Service {
	t.Helper()

	service, err := NewService(profiles, analytics, recaps, options...)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return service
}

func fixedClock() time.Time {
	return time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC)
}
