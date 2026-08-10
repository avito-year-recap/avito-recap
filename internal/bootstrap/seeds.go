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

// DemoData is everything LoadDemoData would otherwise write straight to
// storage, held in memory instead. GenerateDemoData builds it so that
// callers other than the API's own startup seeding — e.g. an out-of-process
// Kafka producer generating events on demand — can reuse the exact same
// scenario expansion without depending on a SeedStorage.
type DemoData struct {
	Profiles         []model.Profile
	Events           []model.Event
	ActionableStates []ProfileActionableState
	ObservedAt       time.Time
}

type ProfileActionableState struct {
	ProfileID uuid.UUID
	State     model.ActionableState
}

// GenerateDemoData expands the profile catalogue and every scenario into the
// raw events a real user's year would have produced. It deliberately does
// not compute a final Metrics blob: AnalyticsStorage.CalculateMetrics
// derives that itself from events the first time it is asked, the same way
// it would for events ingested from the live platform.
func GenerateDemoData(profilesPath, scenariosPath string) (DemoData, error) {
	profiles, err := readJSON[[]model.Profile](profilesPath)
	if err != nil {
		return DemoData{}, fmt.Errorf("load profiles seed: %w", err)
	}
	scenarios, err := readJSON[[]scenario](scenariosPath)
	if err != nil {
		return DemoData{}, fmt.Errorf("load scenarios seed: %w", err)
	}
	byCode := make(map[string]model.Profile, len(profiles))
	for index := range profiles {
		profiles[index] = model.NormalizeProfile(profiles[index])
		profile := profiles[index]
		if err := structural.ValidateProfile(profile); err != nil {
			return DemoData{}, fmt.Errorf("invalid profile seed at index %d: %w", index, err)
		}
		if _, exists := byCode[profile.Code]; exists {
			return DemoData{}, fmt.Errorf("duplicate profile code %q", profile.Code)
		}
		byCode[profile.Code] = profile
	}

	data := DemoData{
		Profiles:   profiles,
		ObservedAt: time.Now().UTC(),
	}
	for _, item := range scenarios {
		profile, ok := byCode[strings.TrimSpace(item.ProfileCode)]
		if !ok {
			return DemoData{}, fmt.Errorf("scenario references unknown profile code %q", item.ProfileCode)
		}
		events, err := eventsFromScenario(profile.ID, item)
		if err != nil {
			return DemoData{}, fmt.Errorf("build events for %s/%d: %w", profile.Code, item.Year, err)
		}
		data.Events = append(data.Events, events...)
		data.ActionableStates = append(data.ActionableStates, ProfileActionableState{
			ProfileID: profile.ID,
			State:     item.ActionableState,
		})
	}
	return data, nil
}

// LoadDemoData seeds the profile catalogue and, for each scenario, the raw
// events a real user's year would have produced. It's a thin SeedStorage
// wrapper around GenerateDemoData.
func LoadDemoData(ctx context.Context, storage SeedStorage, profilesPath, scenariosPath string) error {
	data, err := GenerateDemoData(profilesPath, scenariosPath)
	if err != nil {
		return err
	}

	if err := storage.UpsertProfiles(ctx, data.Profiles); err != nil {
		return fmt.Errorf("seed profiles: %w", err)
	}
	if err := storage.InsertEvents(ctx, data.Events); err != nil {
		return fmt.Errorf("seed events: %w", err)
	}
	for _, entry := range data.ActionableStates {
		if err := storage.PutActionableState(ctx, entry.ProfileID, data.ObservedAt, entry.State); err != nil {
			return fmt.Errorf("seed actionable state for %s: %w", entry.ProfileID, err)
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
