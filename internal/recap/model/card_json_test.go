package model

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCardJSONRoundTripKeepsTypedPayload(t *testing.T) {
	original := Card{ID: "action", Type: CardNextAction, Position: 1, Title: "Дальше", Description: "Описание", Payload: ActionPayload{Code: ActionOpenFavorites, Target: ActionTarget{Route: &RouteTarget{Route: "/favorites"}}}}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var restored Card
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(original, restored) {
		t.Fatalf("round trip mismatch:\n%+v\n%+v", original, restored)
	}
}

func TestCardJSONRejectsMismatchedPayload(t *testing.T) {
	value := Card{ID: "bad", Type: CardBehavior, Title: "x", Description: "y", Payload: ActionPayload{Code: ActionOpenFavorites}}
	if _, err := json.Marshal(value); err == nil {
		t.Fatal("expected mismatched payload error")
	}
}
