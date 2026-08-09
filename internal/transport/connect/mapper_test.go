package connect

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	recapv1 "github.com/year-recap/gen/go/recap/v1"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/testkit"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestRecapToProtoMapsPrivateContractWithoutActionableState(t *testing.T) {
	value := testkit.Recap()
	mapped, err := recapToProto(value)
	if err != nil {
		t.Fatal(err)
	}
	if mapped.Id != value.ID.String() {
		t.Fatalf("unexpected id: %q", mapped.Id)
	}
	if len(mapped.Cards) != len(value.Cards) {
		t.Fatalf("card count = %d, want %d", len(mapped.Cards), len(value.Cards))
	}
	if mapped.Cards[len(mapped.Cards)-1].GetShare() == nil {
		t.Fatal("final card does not contain share projection")
	}
	profile := profileToProto(value.Profile)
	if profile.ProfileCode != value.Profile.Code {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	data, err := protojson.Marshal(mapped)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "actionableState") ||
		strings.Contains(string(data), "draftListingId") {
		t.Fatalf("private data leaked into transport: %s", data)
	}
}

func TestSetCardPayloadMapsEveryClosedUnionVariant(t *testing.T) {
	action := model.ActionPayload{
		Code: model.ActionOpenFavorites,
		Target: model.ActionTarget{
			Route: &model.RouteTarget{Route: "/favorites"},
		},
	}
	tests := []struct {
		name     string
		cardType model.CardType
		payload  model.CardPayload
		wantType reflect.Type
	}{
		{
			name:     "intro",
			cardType: model.CardIntro,
			wantType: reflect.TypeOf((*recapv1.RecapCard_Intro)(nil)),
		},
		{
			name:     "year activity",
			cardType: model.CardYearActivity,
			payload:  model.YearActivityPayload{TotalEvents: 1},
			wantType: reflect.TypeOf((*recapv1.RecapCard_YearActivity)(nil)),
		},
		{
			name:     "top category",
			cardType: model.CardTopCategory,
			payload:  model.TopCategoryPayload{CategoryCode: "auto"},
			wantType: reflect.TypeOf((*recapv1.RecapCard_TopCategory)(nil)),
		},
		{
			name:     "active month",
			cardType: model.CardActiveMonth,
			payload:  model.ActiveMonthPayload{Month: 1},
			wantType: reflect.TypeOf((*recapv1.RecapCard_ActiveMonth)(nil)),
		},
		{
			name:     "behavior",
			cardType: model.CardBehavior,
			payload:  model.BehaviorPayload{Code: model.BehaviorResearcher},
			wantType: reflect.TypeOf((*recapv1.RecapCard_Behavior)(nil)),
		},
		{
			name:     "achievement",
			cardType: model.CardAchievement,
			payload:  model.AchievementPayload{Codes: []model.AchievementCode{model.AchievementBookworm}},
			wantType: reflect.TypeOf((*recapv1.RecapCard_Achievement)(nil)),
		},
		{
			name:     "missed opportunity",
			cardType: model.CardMissedOpportunity,
			payload:  action,
			wantType: reflect.TypeOf((*recapv1.RecapCard_MissedOpportunity)(nil)),
		},
		{
			name:     "next action",
			cardType: model.CardNextAction,
			payload:  action,
			wantType: reflect.TypeOf((*recapv1.RecapCard_NextAction)(nil)),
		},
		{
			name:     "share",
			cardType: model.CardShare,
			payload:  model.ShareCard{ShareID: testkit.ShareID, Year: 2025},
			wantType: reflect.TypeOf((*recapv1.RecapCard_Share)(nil)),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			card := &recapv1.RecapCard{}
			if err := setCardPayload(card, test.cardType, test.payload); err != nil {
				t.Fatal(err)
			}
			if actual := reflect.TypeOf(card.Payload); actual != test.wantType {
				t.Fatalf("oneof type = %v, want %v", actual, test.wantType)
			}
		})
	}
}

func TestSetCardPayloadRejectsTypeMismatch(t *testing.T) {
	err := setCardPayload(&recapv1.RecapCard{}, model.CardBehavior, model.ActiveMonthPayload{Month: 1})
	if !errors.Is(err, errInvalidProjection) {
		t.Fatalf("error = %v, want invalid projection", err)
	}
}

func TestActionTargetToProtoMapsEveryDestination(t *testing.T) {
	tests := []struct {
		name     string
		target   model.ActionTarget
		wantType reflect.Type
	}{
		{
			name:     "route",
			target:   model.ActionTarget{Route: &model.RouteTarget{Route: "/favorites"}},
			wantType: reflect.TypeOf((*recapv1.ActionTarget_Route)(nil)),
		},
		{
			name:     "category",
			target:   model.ActionTarget{Category: &model.CategoryTarget{CategoryCode: "auto"}},
			wantType: reflect.TypeOf((*recapv1.ActionTarget_Category)(nil)),
		},
		{
			name: "listing",
			target: model.ActionTarget{
				Listing: &model.ListingTarget{ListingID: testkit.DraftListingID},
			},
			wantType: reflect.TypeOf((*recapv1.ActionTarget_Listing)(nil)),
		},
		{
			name: "dialog",
			target: model.ActionTarget{
				Dialog: &model.DialogTarget{DialogID: testkit.DraftListingID},
			},
			wantType: reflect.TypeOf((*recapv1.ActionTarget_Dialog)(nil)),
		},
		{
			name:     "search",
			target:   model.ActionTarget{Search: &model.SearchTarget{CategoryCode: "books"}},
			wantType: reflect.TypeOf((*recapv1.ActionTarget_Search)(nil)),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mapped, err := actionTargetToProto(test.target)
			if err != nil {
				t.Fatal(err)
			}
			if actual := reflect.TypeOf(mapped.Target); actual != test.wantType {
				t.Fatalf("target type = %v, want %v", actual, test.wantType)
			}
		})
	}
}

func TestActionTargetToProtoRejectsMultipleDestinations(t *testing.T) {
	_, err := actionTargetToProto(model.ActionTarget{
		Route:    &model.RouteTarget{Route: "/favorites"},
		Category: &model.CategoryTarget{CategoryCode: "auto"},
	})
	if !errors.Is(err, errInvalidProjection) {
		t.Fatalf("error = %v, want invalid projection", err)
	}
}

func TestEnumMappersRejectUnknownValues(t *testing.T) {
	if _, err := behaviorCodeToProto("UNKNOWN"); !errors.Is(err, errInvalidProjection) {
		t.Fatalf("behavior error = %v", err)
	}
	if _, err := achievementCodeToProto("UNKNOWN"); !errors.Is(err, errInvalidProjection) {
		t.Fatalf("achievement error = %v", err)
	}
	if _, err := actionCodeToProto("UNKNOWN"); !errors.Is(err, errInvalidProjection) {
		t.Fatalf("action error = %v", err)
	}
	if _, err := cardTypeToProto("UNKNOWN"); !errors.Is(err, errInvalidProjection) {
		t.Fatalf("card error = %v", err)
	}
}
