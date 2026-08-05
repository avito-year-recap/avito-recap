package recap

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// MarshalJSON persists the payload together with its discriminant. A corrupt
// type/payload pair cannot be serialized as a valid Card.
func (card Card) MarshalJSON() ([]byte, error) {
	if err := validateCardPayload(card.Type, card.Payload); err != nil {
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
	return json.Marshal(cardEnvelope{
		ID: card.ID, Type: card.Type, Position: card.Position, Title: card.Title,
		Description: card.Description, Explanation: card.Explanation,
		Shareable: card.Shareable, Payload: card.Payload,
	})
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
		if err := json.Unmarshal(envelope.Payload, target); err != nil {
			return err
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
	if err := validateCardPayload(result.Type, result.Payload); err != nil {
		return fmt.Errorf("decode card %q: %w", result.ID, err)
	}
	*card = result
	return nil
}
