package model_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/year-recap/internal/recap/engine"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
	"github.com/year-recap/internal/recap/testkit"
)

func TestRecapJSONRoundTripRestoresClosedCardUnion(t *testing.T) {
	original := testkit.Recap()
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded model.Recap
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	core, err := engine.New(ruleset.DefaultRuleset())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := core.ValidateStored(decoded, testkit.Clock()); err != nil {
		t.Fatalf("round-tripped recap is invalid: %v", err)
	}
	for index := range original.Cards {
		if reflect.TypeOf(decoded.Cards[index].Payload) != reflect.TypeOf(original.Cards[index].Payload) {
			t.Fatalf("card %d payload type changed: %T -> %T", index, original.Cards[index].Payload, decoded.Cards[index].Payload)
		}
	}
}

func TestCardJSONRejectsImpossibleTypePayloadPairs(t *testing.T) {
	bad := model.Card{ID: "bad", Type: model.CardYearActivity, Position: 1, Title: "T", Description: "D", Payload: model.ActiveMonthPayload{Month: 1}}
	if _, err := json.Marshal(bad); err == nil {
		t.Fatal("mismatched card was serialized")
	}
	data := []byte(`{"id":"bad","type":"TOP_CATEGORY","position":1,"title":"T","description":"D","shareable":false,"payload":{"month":1}}`)
	var decoded model.Card
	if err := json.Unmarshal(data, &decoded); err == nil {
		t.Fatal("mismatched stored payload was accepted")
	}
}
