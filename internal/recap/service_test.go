package recap

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewServiceRejectsMissingDependencies(t *testing.T) {
	profiles := &profileStorageStub{}
	analytics := &analyticsStorageStub{}
	recaps := &recapStorageStub{}

	tests := []struct {
		name      string
		profiles  ProfileStorage
		analytics AnalyticsStorage
		recaps    RecapStorage
	}{
		{name: "profiles", analytics: analytics, recaps: recaps},
		{name: "analytics", profiles: profiles, recaps: recaps},
		{name: "recaps", profiles: profiles, analytics: analytics},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, err := NewService(test.profiles, test.analytics, test.recaps)
			if service != nil {
				t.Fatalf("expected nil service, got %+v", service)
			}
			if !errors.Is(err, ErrMissingDependency) {
				t.Fatalf("expected ErrMissingDependency, got %v", err)
			}
		})
	}
}

func TestServiceListProfiles(t *testing.T) {
	profiles := &profileStorageStub{profiles: []Profile{validProfile()}}
	service := mustService(t, profiles, &analyticsStorageStub{}, &recapStorageStub{})

	actual, err := service.ListProfiles(context.Background())
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}
	if len(actual) != 1 || actual[0].ID != validProfile().ID {
		t.Fatalf("unexpected profiles: %+v", actual)
	}
}

func TestServiceListProfilesErrors(t *testing.T) {
	storageErr := errors.New("storage unavailable")

	t.Run("storage error", func(t *testing.T) {
		service := mustService(t,
			&profileStorageStub{listErr: storageErr},
			&analyticsStorageStub{},
			&recapStorageStub{},
		)
		_, err := service.ListProfiles(context.Background())
		if !errors.Is(err, storageErr) || !strings.Contains(err.Error(), "list profiles") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid returned profile", func(t *testing.T) {
		service := mustService(t,
			&profileStorageStub{profiles: []Profile{{ID: "id"}}},
			&analyticsStorageStub{},
			&recapStorageStub{},
		)
		_, err := service.ListProfiles(context.Background())
		if !errors.Is(err, ErrInvalidProfile) {
			t.Fatalf("expected ErrInvalidProfile, got %v", err)
		}
	})
}

func TestServiceGenerate(t *testing.T) {
	profiles := &profileStorageStub{profile: validProfile()}
	analytics := &analyticsStorageStub{metrics: validMetrics()}
	recaps := &recapStorageStub{}
	clockCalls := 0

	service := mustService(
		t,
		profiles,
		analytics,
		recaps,
		WithClock(func() time.Time {
			clockCalls++
			return fixedClock()
		}),
		WithIDGenerator(func() (string, error) { return "11111111-1111-4111-8111-111111111111", nil }),
	)

	result, err := service.Generate(context.Background(), "  profile-1  ", 2025)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if profiles.gotID != "profile-1" || analytics.gotID != "profile-1" || analytics.gotYear != 2025 {
		t.Fatalf("input was not normalized/forwarded correctly: profile=%q analytics=%q year=%d", profiles.gotID, analytics.gotID, analytics.gotYear)
	}
	if clockCalls != 1 {
		t.Fatalf("clock called %d times, want exactly once", clockCalls)
	}
	if result.ID != "11111111-1111-4111-8111-111111111111" || result.GeneratedAt != fixedClock() {
		t.Fatalf("unexpected generated metadata: id=%q time=%v", result.ID, result.GeneratedAt)
	}
	if result.RulesVersion != CurrentRulesVersion {
		t.Fatalf("rules version = %q, want %q", result.RulesVersion, CurrentRulesVersion)
	}
	if result.Metrics.FavoriteRate == 0 || result.Behavior.Code == "" || len(result.Cards) == 0 {
		t.Fatalf("recap was not fully generated: %+v", result)
	}
	if recaps.saveCalls != 1 || recaps.saved == nil || recaps.saved.ID != result.ID {
		t.Fatalf("recap was not saved exactly once: calls=%d saved=%+v", recaps.saveCalls, recaps.saved)
	}
}

func TestServiceGenerateValidationErrorsDoNotSave(t *testing.T) {
	futureYear := uint32(fixedClock().Year() + 1)

	tests := []struct {
		name      string
		profileID string
		year      uint32
		profile   Profile
		metrics   Metrics
		expected  error
	}{
		{name: "empty profile id", profileID: " ", year: 2025, profile: validProfile(), metrics: validMetrics(), expected: ErrInvalidProfileID},
		{name: "zero year", profileID: "id", year: 0, profile: validProfile(), metrics: validMetrics(), expected: ErrInvalidYear},
		{name: "future year", profileID: "id", year: futureYear, profile: validProfile(), metrics: validMetrics(), expected: ErrInvalidYear},
		{name: "invalid profile", profileID: "id", year: 2025, profile: Profile{ID: "id"}, metrics: validMetrics(), expected: ErrInvalidProfile},
		{name: "invalid metrics", profileID: "id", year: 2025, profile: validProfile(), metrics: Metrics{TotalEvents: 1, TotalViews: 2}, expected: ErrInvalidMetrics},
		{name: "too little activity", profileID: "id", year: 2025, profile: validProfile(), metrics: Metrics{TotalEvents: minEventsForRecap - 1}, expected: ErrNotEnoughActivity},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recaps := &recapStorageStub{}
			service := mustService(
				t,
				&profileStorageStub{profile: test.profile},
				&analyticsStorageStub{metrics: test.metrics},
				recaps,
				WithClock(func() time.Time { return fixedClock() }),
			)

			_, err := service.Generate(context.Background(), test.profileID, test.year)
			if !errors.Is(err, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, err)
			}
			if recaps.saveCalls != 0 {
				t.Fatalf("invalid recap must not be saved, calls=%d", recaps.saveCalls)
			}
		})
	}
}

func TestServiceGenerateDependencyErrors(t *testing.T) {
	profileErr := errors.New("profile storage failed")
	analyticsErr := errors.New("analytics failed")
	saveErr := errors.New("save failed")

	tests := []struct {
		name      string
		profiles  *profileStorageStub
		analytics *analyticsStorageStub
		recaps    *recapStorageStub
		expected  error
		contains  string
	}{
		{
			name:      "profile storage",
			profiles:  &profileStorageStub{getErr: profileErr},
			analytics: &analyticsStorageStub{},
			recaps:    &recapStorageStub{},
			expected:  profileErr,
			contains:  "get profile",
		},
		{
			name:      "analytics storage",
			profiles:  &profileStorageStub{profile: validProfile()},
			analytics: &analyticsStorageStub{err: analyticsErr},
			recaps:    &recapStorageStub{},
			expected:  analyticsErr,
			contains:  "calculate metrics",
		},
		{
			name:      "recap storage",
			profiles:  &profileStorageStub{profile: validProfile()},
			analytics: &analyticsStorageStub{metrics: validMetrics()},
			recaps:    &recapStorageStub{saveErr: saveErr},
			expected:  saveErr,
			contains:  "save recap",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := mustService(
				t,
				test.profiles,
				test.analytics,
				test.recaps,
				WithClock(func() time.Time { return fixedClock() }),
				WithIDGenerator(func() (string, error) { return "11111111-1111-4111-8111-111111111111", nil }),
			)

			_, err := service.Generate(context.Background(), "profile-1", 2025)
			if !errors.Is(err, test.expected) || !strings.Contains(err.Error(), test.contains) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestServiceGenerateIDErrors(t *testing.T) {
	generatorErr := errors.New("random source failed")

	tests := []struct {
		name      string
		generator IDGenerator
	}{
		{name: "generator error", generator: func() (string, error) { return "", generatorErr }},
		{name: "empty id", generator: func() (string, error) { return " ", nil }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recaps := &recapStorageStub{}
			service := mustService(
				t,
				&profileStorageStub{profile: validProfile()},
				&analyticsStorageStub{metrics: validMetrics()},
				recaps,
				WithClock(func() time.Time { return fixedClock() }),
				WithIDGenerator(test.generator),
			)

			_, err := service.Generate(context.Background(), "profile-1", 2025)
			if !errors.Is(err, ErrGenerateID) {
				t.Fatalf("expected ErrGenerateID, got %v", err)
			}
			if test.name == "generator error" && !errors.Is(err, generatorErr) {
				t.Fatalf("expected source generator error to be wrapped, got %v", err)
			}
			if recaps.saveCalls != 0 {
				t.Fatalf("recap with invalid id must not be saved")
			}
		})
	}
}

func TestServiceGet(t *testing.T) {
	stored := Recap{ID: "11111111-1111-4111-8111-111111111111"}
	recaps := &recapStorageStub{value: stored}
	service := mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, recaps)

	actual, err := service.Get(context.Background(), " 11111111-1111-4111-8111-111111111111 ")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if actual.ID != stored.ID || recaps.gotRecapID != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("unexpected Get result: actual=%+v id=%q", actual, recaps.gotRecapID)
	}
}

func TestServiceGetErrors(t *testing.T) {
	storageErr := errors.New("get failed")

	t.Run("invalid id", func(t *testing.T) {
		service := mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, &recapStorageStub{})
		_, err := service.Get(context.Background(), " ")
		if !errors.Is(err, ErrInvalidRecapID) {
			t.Fatalf("expected ErrInvalidRecapID, got %v", err)
		}
	})

	t.Run("storage error", func(t *testing.T) {
		service := mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, &recapStorageStub{getErr: storageErr})
		_, err := service.Get(context.Background(), "11111111-1111-4111-8111-111111111111")
		if !errors.Is(err, storageErr) || !strings.Contains(err.Error(), "get recap") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestServiceGetShareCard(t *testing.T) {
	recaps := &recapStorageStub{value: Recap{
		ID:       "11111111-1111-4111-8111-111111111111",
		Year:     2025,
		Behavior: Behavior{Title: "Исследователь"},
	}}
	service := mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, recaps)

	actual, err := service.GetShareCard(context.Background(), "11111111-1111-4111-8111-111111111111")
	if err != nil {
		t.Fatalf("GetShareCard() error = %v", err)
	}
	if actual.RecapID != "11111111-1111-4111-8111-111111111111" || actual.BehaviorTitle != "Исследователь" {
		t.Fatalf("unexpected share card: %+v", actual)
	}
}

func TestGenerateID(t *testing.T) {
	first, err := generateID()
	if err != nil {
		t.Fatalf("generateID() error = %v", err)
	}
	second, err := generateID()
	if err != nil {
		t.Fatalf("generateID() second error = %v", err)
	}
	if len(first) != 36 || len(second) != 36 {
		t.Fatalf("unexpected ID lengths: %d, %d", len(first), len(second))
	}
	if first == second {
		t.Fatalf("generated IDs must differ: %q", first)
	}
	if first[14] != '4' {
		t.Fatalf("generated UUID must be version 4: %q", first)
	}
	if !strings.ContainsRune("89ab", rune(first[19])) {
		t.Fatalf("generated UUID has invalid RFC 4122 variant: %q", first)
	}
}

type failingReader struct {
	err error
}

func (r failingReader) Read([]byte) (int, error) {
	return 0, r.err
}

func TestGenerateIDFromReaderError(t *testing.T) {
	sourceErr := errors.New("reader failed")
	value, err := generateIDFrom(failingReader{err: sourceErr})
	if value != "" {
		t.Fatalf("expected empty value, got %q", value)
	}
	if !errors.Is(err, sourceErr) {
		t.Fatalf("expected source error, got %v", err)
	}
}

func TestServiceGetShareCardPropagatesGetError(t *testing.T) {
	storageErr := errors.New("get failed")
	service := mustService(
		t,
		&profileStorageStub{},
		&analyticsStorageStub{},
		&recapStorageStub{getErr: storageErr},
	)

	_, err := service.GetShareCard(context.Background(), "11111111-1111-4111-8111-111111111111")
	if !errors.Is(err, storageErr) {
		t.Fatalf("expected storage error, got %v", err)
	}
}
