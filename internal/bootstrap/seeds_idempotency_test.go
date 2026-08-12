package bootstrap

import (
	"context"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
)

type seedEventKey struct {
	profileID uuid.UUID
	year      uint32
}

type seedStorageSpy struct {
	mu          sync.Mutex
	counts      map[seedEventKey]uint64
	insertCalls int
}

func newSeedStorageSpy() *seedStorageSpy {
	return &seedStorageSpy{counts: make(map[seedEventKey]uint64)}
}

func (s *seedStorageSpy) UpsertProfiles(context.Context, []model.Profile) error { return nil }

func (s *seedStorageSpy) CountEvents(_ context.Context, profileID uuid.UUID, year uint32) (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.counts[seedEventKey{profileID: profileID, year: year}], nil
}

func (s *seedStorageSpy) InsertEvents(_ context.Context, events []model.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.insertCalls++
	if len(events) == 0 {
		return nil
	}
	key := seedEventKey{profileID: events[0].ProfileID, year: uint32(events[0].OccurredAt.Year())}
	s.counts[key] += uint64(len(events))
	return nil
}

func (s *seedStorageSpy) PutActionableState(context.Context, uuid.UUID, time.Time, model.ActionableState) error {
	return nil
}

func TestLoadDemoDataDoesNotDuplicateExistingDemoEvents(t *testing.T) {
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(current), "..", ".."))
	profilesPath := filepath.Join(root, "seeds", "profiles.json")
	scenariosPath := filepath.Join(root, "seeds", "scenarios.json")

	storage := newSeedStorageSpy()
	if err := LoadDemoData(context.Background(), storage, profilesPath, scenariosPath); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	firstCalls := storage.insertCalls
	if firstCalls == 0 {
		t.Fatal("first seed did not insert events")
	}
	if err := LoadDemoData(context.Background(), storage, profilesPath, scenariosPath); err != nil {
		t.Fatalf("second seed: %v", err)
	}
	if storage.insertCalls != firstCalls {
		t.Fatalf("event inserts after second seed = %d, want %d", storage.insertCalls, firstCalls)
	}
}
