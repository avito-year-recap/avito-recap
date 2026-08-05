package recap

import (
	"context"
	"testing"
	"time"
)

type profileStorageStub struct {
	profiles []Profile
	profile  Profile
	listErr  error
	getErr   error
	gotID    string
}

func (s *profileStorageStub) ListProfiles(context.Context) ([]Profile, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.profiles, nil
}

func (s *profileStorageStub) GetProfile(_ context.Context, profileID string) (Profile, error) {
	s.gotID = profileID
	if s.getErr != nil {
		return Profile{}, s.getErr
	}
	return s.profile, nil
}

type analyticsStorageStub struct {
	metrics Metrics
	err     error
	gotID   string
	gotYear uint32
}

func (s *analyticsStorageStub) CalculateMetrics(
	_ context.Context,
	profileID string,
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
	gotRecapID string
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

func (s *recapStorageStub) GetRecap(_ context.Context, recapID string) (Recap, error) {
	s.gotRecapID = recapID
	if s.getErr != nil {
		return Recap{}, s.getErr
	}
	return s.value, nil
}

func validProfile() Profile {
	return Profile{
		ID:          "profile-1",
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
