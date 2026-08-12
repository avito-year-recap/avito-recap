package narrative_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/year-recap/internal/recap/narrative"
)

type failingGenerator struct{}

func (failingGenerator) Generate(context.Context, narrative.Facts) (narrative.Story, error) {
	return narrative.Story{}, errors.New("provider unavailable")
}

type partialGenerator struct{}

func (partialGenerator) Generate(context.Context, narrative.Facts) (narrative.Story, error) {
	return narrative.Story{Cards: []narrative.CardNarrative{{ID: "intro", Description: "only one card"}}}, nil
}

func TestBestEffortFallsBackOnProviderError(t *testing.T) {
	called := false
	recap := sampleRecap()
	enricher := narrative.BestEffort{
		Primary: failingGenerator{},
		OnError: func(error) { called = true },
	}
	enriched, err := enricher.Enrich(context.Background(), recap)
	if err != nil {
		t.Fatal(err)
	}
	if !called || enriched.Cards[0].Description != recap.Cards[0].Description {
		t.Fatalf("fallback not used: called=%v enriched=%+v", called, enriched.Cards[0])
	}
}

func TestBestEffortReportsApplyFailureAndKeepsDeterministicCopy(t *testing.T) {
	var reported error
	recap := sampleRecap()
	enricher := narrative.BestEffort{
		Primary: partialGenerator{},
		OnError: func(err error) { reported = err },
	}
	enriched, err := enricher.Enrich(context.Background(), recap)
	if err != nil {
		t.Fatal(err)
	}
	if reported == nil || !strings.Contains(reported.Error(), "apply AI narrative") {
		t.Fatalf("apply failure was not reported: %v", reported)
	}
	if enriched.Cards[0].Description != recap.Cards[0].Description || enriched.Cards[1].Description != recap.Cards[1].Description {
		t.Fatalf("invalid AI output changed deterministic recap: %+v", enriched.Cards)
	}
}
