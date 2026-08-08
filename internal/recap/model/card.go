package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type CardType string

const (
	CardIntro             CardType = "INTRO"
	CardYearActivity      CardType = "YEAR_ACTIVITY"
	CardTopCategory       CardType = "TOP_CATEGORY"
	CardActiveMonth       CardType = "ACTIVE_MONTH"
	CardBehavior          CardType = "BEHAVIOR"
	CardAchievement       CardType = "ACHIEVEMENT"
	CardMissedOpportunity CardType = "MISSED_OPPORTUNITY"
	CardNextAction        CardType = "NEXT_ACTION"
	CardShare             CardType = "SHARE"
)

// CardPayload is a sealed union: only the payload types below can be assigned.
type CardPayload interface {
	isCardPayload()
}

type YearActivityPayload struct {
	TotalEvents        uint64 `json:"totalEvents"`
	Searches           uint64 `json:"searches"`
	TotalViews         uint64 `json:"totalViews"`
	FavoritesAdded     uint64 `json:"favoritesAdded"`
	ChatsStarted       uint64 `json:"chatsStarted"`
	ListingsPublished  uint64 `json:"listingsPublished"`
	PurchasesCompleted uint64 `json:"purchasesCompleted"`
	SalesCompleted     uint64 `json:"salesCompleted"`
}

func (YearActivityPayload) isCardPayload() {}

type TopCategoryPayload struct {
	CategoryCode  string `json:"categoryCode"`
	Category      string `json:"category"`
	CategoryViews uint64 `json:"categoryViews"`
}

func (TopCategoryPayload) isCardPayload() {}

type ActiveMonthPayload struct {
	Month uint32 `json:"month"`
}

func (ActiveMonthPayload) isCardPayload() {}

type BehaviorPayload struct {
	Code     BehaviorCode       `json:"code"`
	Evidence []BehaviorEvidence `json:"evidence,omitempty"`
}

func (BehaviorPayload) isCardPayload() {}

type AchievementPayload struct {
	Codes []AchievementCode `json:"codes"`
}

func (AchievementPayload) isCardPayload() {}

type ActionPayload struct {
	Code   ActionCode   `json:"code"`
	Target ActionTarget `json:"target"`
}

func (ActionPayload) isCardPayload() {}

type Card struct {
	ID          string      `json:"id"`
	Type        CardType    `json:"type"`
	Position    uint32      `json:"position"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Explanation string      `json:"explanation,omitempty"`
	Shareable   bool        `json:"shareable"`
	Payload     CardPayload `json:"payload,omitempty"`
}

// MarshalJSON persists the payload together with its discriminant. A corrupt
// type/payload pair cannot be serialized as a valid Card.
func (card Card) MarshalJSON() ([]byte, error) {
	if err := validateCardPayloadShape(card.Type, card.Payload); err != nil {
		return nil, fmt.Errorf("marshal card %q: %w", card.ID, err)
	}
	type cardEnvelope struct {
		ID          string      `json:"id"`
		Type        CardType    `json:"type"`
		Position    uint32      `json:"position"`
		Title       string      `json:"title"`
		Description string      `json:"description"`
		Explanation string      `json:"explanation,omitempty"`
		Shareable   bool        `json:"shareable"`
		Payload     CardPayload `json:"payload,omitempty"`
	}
	return json.Marshal(cardEnvelope(card))
}

// UnmarshalJSON reconstructs the closed union from Card.Type and rejects
// missing, extra, unknown, or mismatched payloads at the storage boundary.
func (card *Card) UnmarshalJSON(data []byte) error {
	if card == nil {
		return fmt.Errorf("unmarshal card into nil receiver")
	}
	var envelope struct {
		ID          string          `json:"id"`
		Type        CardType        `json:"type"`
		Position    uint32          `json:"position"`
		Title       string          `json:"title"`
		Description string          `json:"description"`
		Explanation string          `json:"explanation"`
		Shareable   bool            `json:"shareable"`
		Payload     json.RawMessage `json:"payload"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		return fmt.Errorf("decode card envelope: %w", err)
	}

	var payload CardPayload
	decodePayload := func(target CardPayload) error {
		if len(envelope.Payload) == 0 || bytes.Equal(bytes.TrimSpace(envelope.Payload), []byte("null")) {
			return fmt.Errorf("card type %s requires payload", envelope.Type)
		}
		decoder := json.NewDecoder(bytes.NewReader(envelope.Payload))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(target); err != nil {
			return err
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			if err == nil {
				return fmt.Errorf("card payload contains trailing data")
			}
			return fmt.Errorf("card payload contains trailing data: %w", err)
		}
		switch value := target.(type) {
		case *YearActivityPayload:
			payload = *value
		case *TopCategoryPayload:
			payload = *value
		case *ActiveMonthPayload:
			payload = *value
		case *BehaviorPayload:
			payload = *value
		case *AchievementPayload:
			payload = *value
		case *ActionPayload:
			payload = *value
		case *ShareCard:
			payload = *value
		default:
			return fmt.Errorf("unsupported card payload target %T", target)
		}
		return nil
	}

	switch envelope.Type {
	case CardIntro:
		if len(envelope.Payload) != 0 && !bytes.Equal(bytes.TrimSpace(envelope.Payload), []byte("null")) {
			return fmt.Errorf("intro card must not contain payload")
		}
	case CardYearActivity:
		if err := decodePayload(&YearActivityPayload{}); err != nil {
			return err
		}
	case CardTopCategory:
		if err := decodePayload(&TopCategoryPayload{}); err != nil {
			return err
		}
	case CardActiveMonth:
		if err := decodePayload(&ActiveMonthPayload{}); err != nil {
			return err
		}
	case CardBehavior:
		if err := decodePayload(&BehaviorPayload{}); err != nil {
			return err
		}
	case CardAchievement:
		if err := decodePayload(&AchievementPayload{}); err != nil {
			return err
		}
	case CardMissedOpportunity, CardNextAction:
		if err := decodePayload(&ActionPayload{}); err != nil {
			return err
		}
	case CardShare:
		if err := decodePayload(&ShareCard{}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown card type %q", envelope.Type)
	}

	result := Card{
		ID: envelope.ID, Type: envelope.Type, Position: envelope.Position,
		Title: envelope.Title, Description: envelope.Description,
		Explanation: envelope.Explanation, Shareable: envelope.Shareable, Payload: payload,
	}
	if err := validateCardPayloadShape(result.Type, result.Payload); err != nil {
		return fmt.Errorf("decode card %q: %w", result.ID, err)
	}
	*card = result
	return nil
}

func validateCardPayloadShape(cardType CardType, payload CardPayload) error {
	switch cardType {
	case CardIntro:
		if payload != nil {
			return fmt.Errorf("intro card must not contain payload")
		}
	case CardYearActivity:
		if _, ok := payload.(YearActivityPayload); !ok {
			return fmt.Errorf("year-activity payload is required")
		}
	case CardTopCategory:
		if _, ok := payload.(TopCategoryPayload); !ok {
			return fmt.Errorf("top-category payload is required")
		}
	case CardActiveMonth:
		if _, ok := payload.(ActiveMonthPayload); !ok {
			return fmt.Errorf("active-month payload is required")
		}
	case CardBehavior:
		if _, ok := payload.(BehaviorPayload); !ok {
			return fmt.Errorf("behavior payload is required")
		}
	case CardAchievement:
		if _, ok := payload.(AchievementPayload); !ok {
			return fmt.Errorf("achievement payload is required")
		}
	case CardMissedOpportunity, CardNextAction:
		if _, ok := payload.(ActionPayload); !ok {
			return fmt.Errorf("action payload is required")
		}
	case CardShare:
		if _, ok := payload.(ShareCard); !ok {
			return fmt.Errorf("share-card payload is required")
		}
	default:
		return fmt.Errorf("unknown card type %q", cardType)
	}
	return nil
}
