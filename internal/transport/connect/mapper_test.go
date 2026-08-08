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
	if mapped.InternalId != value.ID.String() || mapped.ShareId != value.ShareID.String() {
		t.Fatalf("unexpected ids: internal=%q share=%q", mapped.InternalId, mapped.ShareId)
	}
	if mapped.Profile.Id != value.Profile.ID.String() || mapped.Profile.AvatarUrl != nil {
		t.Fatalf("unexpected profile: %+v", mapped.Profile)
	}
	if len(mapped.Cards) != len(value.Cards) {
		t.Fatalf("card count = %d, want %d", len(mapped.Cards), len(value.Cards))
	}
	if mapped.Cards[len(mapped.Cards)-1].GetPayload().GetShare() == nil {
		t.Fatal("final card does not contain share projection")
	}
	data, err := protojson.Marshal(mapped)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "actionableState") ||
		strings.Contains(string(data), "draftListingId") {
		t.Fatalf("private actionable state leaked into transport: %s", data)
	}
}

func TestCardPayloadToProtoMapsEveryClosedUnionVariant(t *testing.T) {
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
		{name: "intro", cardType: model.CardIntro},
		{
			name:     "year activity",
			cardType: model.CardYearActivity,
			payload:  model.YearActivityPayload{TotalEvents: 1},
			wantType: reflect.TypeOf((*recapv1.CardPayload_YearActivity)(nil)),
		},
		{
			name:     "top category",
			cardType: model.CardTopCategory,
			payload:  model.TopCategoryPayload{CategoryCode: "auto"},
			wantType: reflect.TypeOf((*recapv1.CardPayload_TopCategory)(nil)),
		},
		{
			name:     "active month",
			cardType: model.CardActiveMonth,
			payload:  model.ActiveMonthPayload{Month: 1},
			wantType: reflect.TypeOf((*recapv1.CardPayload_ActiveMonth)(nil)),
		},
		{
			name:     "behavior",
			cardType: model.CardBehavior,
			payload:  model.BehaviorPayload{Code: model.BehaviorResearcher},
			wantType: reflect.TypeOf((*recapv1.CardPayload_Behavior)(nil)),
		},
		{
			name:     "achievement",
			cardType: model.CardAchievement,
			payload:  model.AchievementPayload{Codes: []model.AchievementCode{model.AchievementBookworm}},
			wantType: reflect.TypeOf((*recapv1.CardPayload_Achievement)(nil)),
		},
		{
			name:     "missed opportunity",
			cardType: model.CardMissedOpportunity,
			payload:  action,
			wantType: reflect.TypeOf((*recapv1.CardPayload_Action)(nil)),
		},
		{
			name:     "next action",
			cardType: model.CardNextAction,
			payload:  action,
			wantType: reflect.TypeOf((*recapv1.CardPayload_Action)(nil)),
		},
		{
			name:     "share",
			cardType: model.CardShare,
			payload:  model.ShareCard{ShareID: testkit.ShareID, Year: 2025},
			wantType: reflect.TypeOf((*recapv1.CardPayload_Share)(nil)),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mapped, err := cardPayloadToProto(test.cardType, test.payload)
			if err != nil {
				t.Fatal(err)
			}
			if test.wantType == nil {
				if mapped != nil {
					t.Fatalf("payload = %+v, want nil", mapped)
				}
				return
			}
			if mapped == nil {
				t.Fatal("payload is nil")
			}
			if actual := reflect.TypeOf(mapped.Value); actual != test.wantType {
				t.Fatalf("oneof type = %v, want %v", actual, test.wantType)
			}
		})
	}
}

func TestCardPayloadToProtoRejectsTypeMismatch(t *testing.T) {
	_, err := cardPayloadToProto(model.CardBehavior, model.ActiveMonthPayload{Month: 1})
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
			if actual := reflect.TypeOf(mapped.Destination); actual != test.wantType {
				t.Fatalf("destination type = %v, want %v", actual, test.wantType)
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
