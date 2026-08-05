package recap

import (
	"errors"
	"testing"
)

func TestValidateProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		wantErr bool
	}{
		{name: "valid", profile: validProfile()},
		{name: "empty id", profile: Profile{Code: "code", DisplayName: "Name"}, wantErr: true},
		{name: "blank id", profile: Profile{ID: "  ", Code: "code", DisplayName: "Name"}, wantErr: true},
		{name: "empty code", profile: Profile{ID: "id", DisplayName: "Name"}, wantErr: true},
		{name: "blank code", profile: Profile{ID: "id", Code: " ", DisplayName: "Name"}, wantErr: true},
		{name: "empty display name", profile: Profile{ID: "id", Code: "code"}, wantErr: true},
		{name: "blank display name", profile: Profile{ID: "id", Code: "code", DisplayName: "\t"}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateProfile(test.profile)
			if test.wantErr && !errors.Is(err, ErrInvalidProfile) {
				t.Fatalf("expected ErrInvalidProfile, got %v", err)
			}
			if !test.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateMetrics(t *testing.T) {
	valid := validMetrics()

	tests := []struct {
		name   string
		mutate func(*Metrics)
	}{
		{name: "known events overflow", mutate: func(m *Metrics) { m.TotalEvents = ^uint64(0); m.Searches = ^uint64(0); m.TotalViews = 1 }},
		{name: "known events exceed total", mutate: func(m *Metrics) { m.TotalEvents = 1 }},
		{name: "unique listings exceed views", mutate: func(m *Metrics) { m.UniqueListings = m.TotalViews + 1 }},
		{name: "repeated views exceed views", mutate: func(m *Metrics) { m.RepeatedViews = m.TotalViews + 1 }},
		{name: "linked chats exceed started chats", mutate: func(m *Metrics) { m.ChatsWithPurchase = m.ChatsStarted + 1 }},
		{name: "top views exceed views", mutate: func(m *Metrics) { m.TopCategoryViews = m.TotalViews + 1 }},
		{name: "title without code", mutate: func(m *Metrics) { m.TopCategoryCode = "" }},
		{name: "code without title", mutate: func(m *Metrics) { m.TopCategory = ""; m.TopCategoryViews = 0 }},
		{name: "category without views", mutate: func(m *Metrics) { m.TopCategoryViews = 0 }},
		{name: "invalid month", mutate: func(m *Metrics) { m.MostActiveMonth = 13 }},
		{name: "shareable empty category", mutate: func(m *Metrics) {
			m.TopCategoryCode = ""
			m.TopCategory = ""
			m.TopCategoryViews = 0
			m.TopCategoryShareable = true
		}},
		{name: "categories exceed events", mutate: func(m *Metrics) { m.CategoriesCount = m.TotalEvents + 1 }},
		{name: "active days exceed events", mutate: func(m *Metrics) {
			*m = Metrics{TotalEvents: 10, ActiveDays: 11}
		}},
		{name: "too many active days", mutate: func(m *Metrics) { m.TotalEvents = 500; m.ActiveDays = 367 }},
	}

	if err := validateMetrics(valid); err != nil {
		t.Fatalf("valid metrics rejected: %v", err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metrics := valid
			test.mutate(&metrics)
			err := validateMetrics(metrics)
			if !errors.Is(err, ErrInvalidMetrics) {
				t.Fatalf("expected ErrInvalidMetrics, got %v", err)
			}
		})
	}
}

func TestValidateMetricsAllowsCrossPeriodCounters(t *testing.T) {
	metrics := Metrics{
		TotalEvents:        17,
		TotalViews:         1,
		UniqueListings:     1,
		FavoritesAdded:     5,
		ChatsStarted:       3,
		ListingsPublished:  3,
		PurchasesCompleted: 2,
		SalesCompleted:     3,
	}

	if err := validateMetrics(metrics); err != nil {
		t.Fatalf("cross-period counters must not be treated as one funnel: %v", err)
	}
}

func TestIsUUID(t *testing.T) {
	tests := []struct {
		value string
		valid bool
	}{
		{value: "11111111-1111-4111-8111-111111111111", valid: true},
		{value: "00000000-0000-0000-0000-000000000000", valid: false},
		{value: "", valid: false},
		{value: "11111111111141118111111111111111", valid: false},
		{value: "11111111-1111-4111-8111-11111111111z", valid: false},
		{value: "11111111-1111-4111-8111", valid: false},
	}

	for _, test := range tests {
		if actual := isUUID(test.value); actual != test.valid {
			t.Fatalf("isUUID(%q) = %t, want %t", test.value, actual, test.valid)
		}
	}
}
