package structural

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
	"strings"
)

func ValidateCards(cards []model.Card) error {
	if len(cards) == 0 {
		return errors.New("cards are required")
	}
	seen := make(map[string]struct{}, len(cards))
	shareCards := 0
	for index, card := range cards {
		if !model.IsKnownCardType(card.Type) {
			return fmt.Errorf("card %d has unknown type %q", index, card.Type)
		}
		if card.ID == "" || card.Title == "" || card.Description == "" {
			return fmt.Errorf("card %d is incomplete", index)
		}
		if card.Position != uint32(index+1) {
			return fmt.Errorf("card %q has position %d, want %d", card.ID, card.Position, index+1)
		}
		if _, ok := seen[card.ID]; ok {
			return fmt.Errorf("duplicate card id %q", card.ID)
		}
		seen[card.ID] = struct{}{}
		if card.Type == model.CardShare {
			shareCards++
			if index != len(cards)-1 {
				return errors.New("share card must be the final story card")
			}
			if !card.Shareable {
				return errors.New("share card must be marked shareable")
			}
		} else if card.Shareable {
			return fmt.Errorf("only the final share card may be marked shareable, got %q", card.ID)
		}
		if err := ValidateCardPayload(card.Type, card.Payload); err != nil {
			return fmt.Errorf("card %q: %w", card.ID, err)
		}
	}
	if shareCards != 1 {
		return fmt.Errorf("exactly one share card is required, got %d", shareCards)
	}
	return nil
}

func ValidateCardPayload(cardType model.CardType, payload model.CardPayload) error {
	switch cardType {
	case model.CardIntro:
		if payload != nil {
			return errors.New("card must not have a payload")
		}
	case model.CardShare:
		value, ok := payload.(model.ShareCard)
		if !ok {
			return errors.New("requires share-card payload")
		}
		if err := ValidateShareCard(value); err != nil {
			return err
		}
	case model.CardYearActivity:
		if _, ok := payload.(model.YearActivityPayload); !ok {
			return errors.New("requires year-activity payload")
		}
	case model.CardTopCategory:
		value, ok := payload.(model.TopCategoryPayload)
		if !ok || value.CategoryCode == "" || value.Category == "" || value.CategoryViews == 0 {
			return errors.New("requires complete top-category payload")
		}
	case model.CardActiveMonth:
		value, ok := payload.(model.ActiveMonthPayload)
		if !ok || value.Month < 1 || value.Month > 12 {
			return errors.New("requires valid active-month payload")
		}
	case model.CardBehavior:
		value, ok := payload.(model.BehaviorPayload)
		if !ok || !model.IsKnownBehaviorCode(value.Code) {
			return errors.New("requires valid behavior payload")
		}
	case model.CardAchievement:
		value, ok := payload.(model.AchievementPayload)
		if !ok || len(value.Codes) == 0 || len(value.Codes) > ruleset.MaxAchievements {
			return errors.New("requires one to three achievement codes")
		}
		seen := make(map[model.AchievementCode]struct{}, len(value.Codes))
		for _, code := range value.Codes {
			if !model.IsKnownAchievementCode(code) {
				return errors.New("achievement payload has unknown code")
			}
			if _, exists := seen[code]; exists {
				return errors.New("achievement payload has duplicate code")
			}
			seen[code] = struct{}{}
		}
	case model.CardMissedOpportunity, model.CardNextAction:
		value, ok := payload.(model.ActionPayload)
		if !ok || !model.IsKnownActionCode(value.Code) {
			return errors.New("requires action payload")
		}
		if err := ValidateActionTarget(value.Target); err != nil {
			return err
		}
		if err := ValidateTargetForAction(value.Code, value.Target); err != nil {
			return err
		}
	}
	return nil
}

func ValidateShareCard(value model.ShareCard) error {
	if value.ShareID == uuid.Nil {
		return errors.New("share id is required")
	}
	if value.Year == 0 {
		return errors.New("share year is required")
	}
	if strings.TrimSpace(value.PrivacyVersion) == "" {
		return errors.New("privacy version is required")
	}
	if strings.TrimSpace(value.BehaviorTitle) == "" {
		return errors.New("behavior title is required")
	}
	return nil
}
