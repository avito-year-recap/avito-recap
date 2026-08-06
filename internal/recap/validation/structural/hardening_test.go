package structural_test

import (
	"errors"
	"testing"

	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/testkit"
	"github.com/year-recap/internal/recap/validation/structural"
)

func TestAddressableStateAndTargetsAreStrict(t *testing.T) {
	for _, state := range []model.ActionableState{
		{CapturedAt: testkit.Clock(), CurrentDrafts: 1},
		{CapturedAt: testkit.Clock(), OpenDialogs: 1},
		{CapturedAt: testkit.Clock(), ActiveListings: 1},
	} {
		if err := structural.ValidateActionableState(state); !errors.Is(err, structural.ErrInvalidActionableState) {
			t.Fatalf("state accepted: %+v", state)
		}
	}
	badRoute := model.NextAction{
		Code: model.ActionOpenFavorites, Title: "T", Description: "D", ButtonText: "B", Reason: "R",
		Target: model.ActionTarget{Route: &model.RouteTarget{Route: "//evil.example"}},
	}
	if err := structural.ValidateNextAction(badRoute); err == nil {
		t.Fatal("unsafe route accepted")
	}
	badCategory := model.NextAction{
		Code: model.ActionOpenTopCategory, Title: "T", Description: "D", ButtonText: "B", Reason: "R",
		Target: model.ActionTarget{Category: &model.CategoryTarget{CategoryCode: "cars/../private"}},
	}
	if err := structural.ValidateNextAction(badCategory); err == nil {
		t.Fatal("unsafe category code accepted")
	}
}

func TestActionTargetRequiresOneCompatibleDestination(t *testing.T) {
	invalid := model.NextAction{
		Code: model.ActionOpenFavorites, Title: "x", Description: "y", ButtonText: "z", Reason: "r",
		Target: model.ActionTarget{Category: &model.CategoryTarget{CategoryCode: "electronics"}},
	}
	if err := structural.ValidateNextAction(invalid); err == nil {
		t.Fatal("expected incompatible target error")
	}
}
