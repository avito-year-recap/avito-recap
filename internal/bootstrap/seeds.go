package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/validation/structural"
)

type SeedStorage interface {
	UpsertProfiles(context.Context, []model.Profile) error
	CountEvents(context.Context, uuid.UUID, uint32) (uint64, error)
	InsertEvents(context.Context, []model.Event) error
	PutActionableState(context.Context, uuid.UUID, time.Time, model.ActionableState) error
}

type scenario struct {
	ProfileCode     string                `json:"profileCode"`
	Year            uint32                `json:"year"`
	Activity        activity              `json:"activity"`
	Categories      []weightedCategory    `json:"categories"`
	Months          []weightedMonth       `json:"months"`
	ActionableState model.ActionableState `json:"actionableState"`
}

type activity struct {
	Searches           uint64 `json:"searches"`
	ListingViews       uint64 `json:"listingViews"`
	UniqueListings     uint64 `json:"uniqueListings"`
	FavoritesAdded     uint64 `json:"favoritesAdded"`
	ChatsStarted       uint64 `json:"chatsStarted"`
	ChatsWithPurchase  uint64 `json:"chatsWithPurchase"`
	ListingsCreated    uint64 `json:"listingsCreated"`
	ListingsPublished  uint64 `json:"listingsPublished"`
	PurchasesCompleted uint64 `json:"purchasesCompleted"`
	SalesCompleted     uint64 `json:"salesCompleted"`
	ActiveDays         uint64 `json:"activeDays"`
}

type weightedCategory struct {
	Code               string `json:"code"`
	Title              string `json:"title"`
	Weight             uint32 `json:"weight"`
	Shareable          bool   `json:"shareable"`
	Views              uint64 `json:"views,omitempty"`
	FavoritesAdded     uint64 `json:"favoritesAdded,omitempty"`
	PurchasesCompleted uint64 `json:"purchasesCompleted,omitempty"`
}

type weightedMonth struct {
	Month  uint32 `json:"month"`
	Weight uint32 `json:"weight"`
}

// LoadDemoData seeds the profile catalogue and, for each scenario, the raw
// events a real user's year would have produced. It deliberately does not
// compute and store a final Metrics blob: AnalyticsStorage.CalculateMetrics
// derives that itself from these events the first time it is asked, the same
// way it would for events ingested from the live platform.
func LoadDemoData(ctx context.Context, storage SeedStorage, profilesPath, scenariosPath string) error {
	profiles, err := readJSON[[]model.Profile](profilesPath)
	if err != nil {
		return fmt.Errorf("load profiles seed: %w", err)
	}
	scenarios, err := readJSON[[]scenario](scenariosPath)
	if err != nil {
		return fmt.Errorf("load scenarios seed: %w", err)
	}
	byCode := make(map[string]model.Profile, len(profiles))
	for index := range profiles {
		profiles[index] = model.NormalizeProfile(profiles[index])
		profile := profiles[index]
		if err := structural.ValidateProfile(profile); err != nil {
			return fmt.Errorf("invalid profile seed at index %d: %w", index, err)
		}
		if _, exists := byCode[profile.Code]; exists {
			return fmt.Errorf("duplicate profile code %q", profile.Code)
		}
		byCode[profile.Code] = profile
	}
	if err := storage.UpsertProfiles(ctx, profiles); err != nil {
		return fmt.Errorf("seed profiles: %w", err)
	}

	observedAt := time.Now().UTC()
	for _, item := range scenarios {
		profile, ok := byCode[strings.TrimSpace(item.ProfileCode)]
		if !ok {
			return fmt.Errorf("scenario references unknown profile code %q", item.ProfileCode)
		}
		events, err := eventsFromScenario(profile.ID, item)
		if err != nil {
			return fmt.Errorf("build events for %s/%d: %w", profile.Code, item.Year, err)
		}

		// Docker Compose keeps ClickHouse data in a named volume. Seed loading can
		// therefore run again after an API restart. Never append the same demo year
		// twice: an unchanged seed is skipped, while a partial/mismatched seed is
		// surfaced explicitly instead of silently corrupting annual metrics.
		existingEvents, err := storage.CountEvents(ctx, profile.ID, item.Year)
		if err != nil {
			return fmt.Errorf("count existing events for %s/%d: %w", profile.Code, item.Year, err)
		}
		expectedEvents := uint64(len(events))
		switch {
		case existingEvents == 0:
			if err := storage.InsertEvents(ctx, events); err != nil {
				return fmt.Errorf("seed events for %s/%d: %w", profile.Code, item.Year, err)
			}
		case existingEvents == expectedEvents:
			// Already seeded with the same demo cardinality.
		default:
			return fmt.Errorf(
				"existing demo events for %s/%d have count %d, want 0 or %d; reset the demo volume before reseeding changed scenarios",
				profile.Code, item.Year, existingEvents, expectedEvents,
			)
		}
		if err := storage.PutActionableState(ctx, profile.ID, observedAt, item.ActionableState); err != nil {
			return fmt.Errorf("seed actionable state for %s: %w", profile.Code, err)
		}
	}
	return nil
}

func readJSON[T any](path string) (T, error) {
	var zero T
	data, err := os.ReadFile(path)
	if err != nil {
		return zero, err
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return zero, err
	}
	return value, nil
}
