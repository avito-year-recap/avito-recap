package connect

import (
	"fmt"

	recapv1 "github.com/year-recap/gen/go/recap/v1"
	"github.com/year-recap/internal/recap/model"
)

func profileToProto(value model.Profile) *recapv1.Profile {
	return &recapv1.Profile{
		Name:        value.DisplayName,
		Description: value.Description,
		AvatarUrl:   value.AvatarURL,
		ProfileCode: value.Code,
	}
}

func recapToProto(value model.Recap) (*recapv1.Recap, error) {
	nextAction, err := nextActionToProto(value.NextAction)
	if err != nil {
		return nil, err
	}
	achievements := make([]*recapv1.Achievement, 0, len(value.Achievements))
	for index, achievement := range value.Achievements {
		mapped, err := achievementToProto(achievement)
		if err != nil {
			return nil, fmt.Errorf("achievement %d: %w", index, err)
		}
		achievements = append(achievements, mapped)
	}
	cards := make([]*recapv1.RecapCard, 0, len(value.Cards))
	for index, card := range value.Cards {
		mapped, err := cardToProto(card)
		if err != nil {
			return nil, fmt.Errorf("card %d: %w", index, err)
		}
		cards = append(cards, mapped)
	}
	return &recapv1.Recap{
		Id:           value.ID.String(),
		Year:         value.Year,
		RuleVersion:  value.RulesVersion,
		Cards:        cards,
		Achievements: achievements,
		NextAction:   nextAction,
	}, nil
}

func behaviorEvidenceToProto(values []model.BehaviorEvidence) []*recapv1.BehaviorEvidence {
	mapped := make([]*recapv1.BehaviorEvidence, 0, len(values))
	for _, value := range values {
		mapped = append(mapped, &recapv1.BehaviorEvidence{
			Metric: value.Metric,
			// Label, Comparison and Points are part of the wire contract but
			// not yet produced by the domain model (internal/recap/model.
			// BehaviorEvidence only has Metric/Actual/Threshold/Detail) —
			// left at zero value until the service layer computes them.
			ActualValue: value.Actual,
			Threshold:   value.Threshold,
			Explanation: value.Detail,
		})
	}
	return mapped
}

func achievementToProto(value model.Achievement) (*recapv1.Achievement, error) {
	code, err := achievementCodeToProto(value.Code)
	if err != nil {
		return nil, err
	}
	return &recapv1.Achievement{
		Code:      code,
		Title:     value.Title,
		Reason:    value.Reason,
		Shareable: value.Shareable,
	}, nil
}

func nextActionToProto(value model.NextAction) (*recapv1.NextAction, error) {
	code, err := actionCodeToProto(value.Code)
	if err != nil {
		return nil, err
	}
	target, err := actionTargetToProto(value.Target)
	if err != nil {
		return nil, err
	}
	return &recapv1.NextAction{
		Code:        code,
		Title:       value.Title,
		Description: value.Description,
		ButtonText:  value.ButtonText,
		Explanation: value.Reason,
		Target:      target,
	}, nil
}

func actionTargetToProto(value model.ActionTarget) (*recapv1.ActionTarget, error) {
	mapped := &recapv1.ActionTarget{}
	destinations := 0
	if value.Route != nil {
		destinations++
		mapped.Target = &recapv1.ActionTarget_Route{
			Route: &recapv1.RouteTarget{Path: value.Route.Route},
		}
	}
	if value.Category != nil {
		destinations++
		mapped.Target = &recapv1.ActionTarget_Category{
			Category: &recapv1.CategoryTarget{CategoryCode: value.Category.CategoryCode},
		}
	}
	if value.Listing != nil {
		destinations++
		mapped.Target = &recapv1.ActionTarget_Listing{
			Listing: &recapv1.ListingTarget{ListingId: value.Listing.ListingID.String()},
		}
	}
	if value.Dialog != nil {
		destinations++
		mapped.Target = &recapv1.ActionTarget_Dialog{
			Dialog: &recapv1.DialogTarget{DialogId: value.Dialog.DialogID.String()},
		}
	}
	if value.Search != nil {
		destinations++
		mapped.Target = &recapv1.ActionTarget_Search{
			Search: &recapv1.SearchTarget{CategoryCode: value.Search.CategoryCode},
		}
	}
	if destinations != 1 {
		return nil, fmt.Errorf("%w: action target has %d destinations", errInvalidProjection, destinations)
	}
	return mapped, nil
}

func cardToProto(value model.Card) (*recapv1.RecapCard, error) {
	cardType, err := cardTypeToProto(value.Type)
	if err != nil {
		return nil, err
	}
	card := &recapv1.RecapCard{
		Id:          value.ID,
		Type:        cardType,
		Position:    value.Position,
		Title:       value.Title,
		Description: value.Description,
		Explanation: value.Explanation,
		Shareable:   value.Shareable,
	}
	if err := setCardPayload(card, value.Type, value.Payload); err != nil {
		return nil, err
	}
	return card, nil
}

func setCardPayload(card *recapv1.RecapCard, cardType model.CardType, payload model.CardPayload) error {
	switch cardType {
	case model.CardIntro:
		if payload != nil {
			return fmt.Errorf("%w: INTRO card must not have a payload", errInvalidProjection)
		}
		card.Payload = &recapv1.RecapCard_Intro{Intro: &recapv1.IntroPayload{}}
	case model.CardYearActivity:
		value, ok := payload.(model.YearActivityPayload)
		if !ok {
			return payloadTypeError(cardType, payload)
		}
		card.Payload = &recapv1.RecapCard_YearActivity{
			YearActivity: &recapv1.YearActivityPayload{
				TotalEvents:        value.TotalEvents,
				Searches:           value.Searches,
				TotalViews:         value.TotalViews,
				FavoritesAdded:     value.FavoritesAdded,
				ChatsStarted:       value.ChatsStarted,
				ListingsPublished:  value.ListingsPublished,
				PurchasesCompleted: value.PurchasesCompleted,
				SalesCompleted:     value.SalesCompleted,
			},
		}
	case model.CardTopCategory:
		value, ok := payload.(model.TopCategoryPayload)
		if !ok {
			return payloadTypeError(cardType, payload)
		}
		card.Payload = &recapv1.RecapCard_TopCategory{
			TopCategory: &recapv1.TopCategoryPayload{
				CategoryCode:  value.CategoryCode,
				Category:      value.Category,
				CategoryViews: value.CategoryViews,
			},
		}
	case model.CardActiveMonth:
		value, ok := payload.(model.ActiveMonthPayload)
		if !ok {
			return payloadTypeError(cardType, payload)
		}
		card.Payload = &recapv1.RecapCard_ActiveMonth{
			ActiveMonth: &recapv1.ActiveMonthPayload{Month: value.Month},
		}
	case model.CardBehavior:
		value, ok := payload.(model.BehaviorPayload)
		if !ok {
			return payloadTypeError(cardType, payload)
		}
		code, err := behaviorCodeToProto(value.Code)
		if err != nil {
			return err
		}
		card.Payload = &recapv1.RecapCard_Behavior{
			Behavior: &recapv1.BehaviorPayload{
				Code:     code,
				Evidence: behaviorEvidenceToProto(value.Evidence),
			},
		}
	case model.CardAchievement:
		value, ok := payload.(model.AchievementPayload)
		if !ok {
			return payloadTypeError(cardType, payload)
		}
		codes := make([]recapv1.AchievementCode, 0, len(value.Codes))
		for _, code := range value.Codes {
			mapped, err := achievementCodeToProto(code)
			if err != nil {
				return err
			}
			codes = append(codes, mapped)
		}
		card.Payload = &recapv1.RecapCard_Achievement{
			Achievement: &recapv1.AchievementPayload{Codes: codes},
		}
	case model.CardMissedOpportunity:
		value, ok := payload.(model.ActionPayload)
		if !ok {
			return payloadTypeError(cardType, payload)
		}
		code, err := actionCodeToProto(value.Code)
		if err != nil {
			return err
		}
		target, err := actionTargetToProto(value.Target)
		if err != nil {
			return err
		}
		card.Payload = &recapv1.RecapCard_MissedOpportunity{
			MissedOpportunity: &recapv1.MissedOpportunityPayload{Code: code, Target: target},
		}
	case model.CardNextAction:
		value, ok := payload.(model.ActionPayload)
		if !ok {
			return payloadTypeError(cardType, payload)
		}
		code, err := actionCodeToProto(value.Code)
		if err != nil {
			return err
		}
		target, err := actionTargetToProto(value.Target)
		if err != nil {
			return err
		}
		card.Payload = &recapv1.RecapCard_NextAction{
			NextAction: &recapv1.NextActionPayload{Code: code, Target: target},
		}
	case model.CardShare:
		value, ok := payload.(model.ShareCard)
		if !ok {
			return payloadTypeError(cardType, payload)
		}
		card.Payload = &recapv1.RecapCard_Share{Share: sharePayloadToProto(value)}
	default:
		return fmt.Errorf("%w: unknown card type %q", errInvalidProjection, cardType)
	}
	return nil
}

func sharePayloadToProto(value model.ShareCard) *recapv1.SharePayload {
	return &recapv1.SharePayload{
		ShareId:          value.ShareID.String(),
		Year:             value.Year,
		BehaviorTitle:    value.BehaviorTitle,
		AchievementTitle: optionalString(value.AchievementTitle),
		TopCategory:      optionalString(value.TopCategory),
	}
}

func publicShareToProto(value model.ShareCard) *recapv1.PublicShare {
	return &recapv1.PublicShare{
		ShareId:          value.ShareID.String(),
		Year:             value.Year,
		BehaviorTitle:    value.BehaviorTitle,
		AchievementTitle: optionalString(value.AchievementTitle),
		TopCategory:      optionalString(value.TopCategory),
	}
}

func behaviorCodeToProto(value model.BehaviorCode) (recapv1.BehaviorCode, error) {
	switch value {
	case model.BehaviorActiveSeller:
		return recapv1.BehaviorCode_BEHAVIOR_CODE_ACTIVE_SELLER, nil
	case model.BehaviorStartingSeller:
		return recapv1.BehaviorCode_BEHAVIOR_CODE_STARTING_SELLER, nil
	case model.BehaviorDecisiveBuyer:
		return recapv1.BehaviorCode_BEHAVIOR_CODE_DECISIVE_BUYER, nil
	case model.BehaviorFindHunter:
		return recapv1.BehaviorCode_BEHAVIOR_CODE_FIND_HUNTER, nil
	case model.BehaviorResearcher:
		return recapv1.BehaviorCode_BEHAVIOR_CODE_RESEARCHER, nil
	case model.BehaviorUniversal:
		return recapv1.BehaviorCode_BEHAVIOR_CODE_UNIVERSAL_USER, nil
	default:
		return recapv1.BehaviorCode_BEHAVIOR_CODE_UNSPECIFIED, fmt.Errorf(
			"%w: unknown behavior code %q",
			errInvalidProjection,
			value,
		)
	}
}

func achievementCodeToProto(value model.AchievementCode) (recapv1.AchievementCode, error) {
	switch value {
	case model.AchievementSuccessfulSeller:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_SUCCESSFUL_SELLER, nil
	case model.AchievementConsistentPublisher:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_CONSISTENT_PUBLISHER, nil
	case model.AchievementAttentiveResearcher:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_ATTENTIVE_RESEARCHER, nil
	case model.AchievementMasterOfFavorites:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_MASTER_OF_FAVORITES, nil
	case model.AchievementBroadInterests:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_BROAD_INTERESTS, nil
	case model.AchievementAllRounder:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_ALL_ROUNDER, nil
	case model.AchievementFirstSellingSteps:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_FIRST_SELLING_STEPS, nil
	case model.AchievementDealCloser:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_DEAL_CLOSER, nil
	case model.AchievementQuickDecision:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_QUICK_DECISION, nil
	case model.AchievementStyleIcon:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_STYLE_ICON, nil
	case model.AchievementFashionableMan:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_FASHIONABLE_MAN, nil
	case model.AchievementTraveler:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_TRAVELER, nil
	case model.AchievementForTheSoul:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_FOR_THE_SOUL, nil
	case model.AchievementBookworm:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_BOOKWORM, nil
	case model.AchievementBeautyConnoisseur:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_BEAUTY_CONNOISSEUR, nil
	case model.AchievementInTheRhythmOfMusic:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_IN_THE_RHYTHM_OF_MUSIC, nil
	case model.AchievementWorldOfPlay:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_WORLD_OF_PLAY, nil
	case model.AchievementMasterCraft:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_MASTER_CRAFT, nil
	case model.AchievementCaringOwner:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_CARING_OWNER, nil
	case model.AchievementLittleDiscoveries:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_LITTLE_DISCOVERIES, nil
	default:
		return recapv1.AchievementCode_ACHIEVEMENT_CODE_UNSPECIFIED, fmt.Errorf(
			"%w: unknown achievement code %q",
			errInvalidProjection,
			value,
		)
	}
}

func actionCodeToProto(value model.ActionCode) (recapv1.ActionCode, error) {
	switch value {
	case model.ActionFinishDraft:
		return recapv1.ActionCode_ACTION_CODE_FINISH_DRAFT, nil
	case model.ActionOpenFavorites:
		return recapv1.ActionCode_ACTION_CODE_OPEN_FAVORITES, nil
	case model.ActionImproveListings:
		return recapv1.ActionCode_ACTION_CODE_IMPROVE_LISTINGS, nil
	case model.ActionContinueDialogs:
		return recapv1.ActionCode_ACTION_CODE_CONTINUE_DIALOGS, nil
	case model.ActionOpenTopCategory:
		return recapv1.ActionCode_ACTION_CODE_OPEN_TOP_CATEGORY, nil
	case model.ActionCreateFirstListing:
		return recapv1.ActionCode_ACTION_CODE_CREATE_FIRST_LISTING, nil
	case model.ActionCreateListing:
		return recapv1.ActionCode_ACTION_CODE_CREATE_LISTING, nil
	case model.ActionSaveSearch:
		return recapv1.ActionCode_ACTION_CODE_SAVE_SEARCH, nil
	case model.ActionViewSimilarListings:
		return recapv1.ActionCode_ACTION_CODE_VIEW_SIMILAR_LISTINGS, nil
	case model.ActionExploreRecommendations:
		return recapv1.ActionCode_ACTION_CODE_EXPLORE_RECOMMENDATIONS, nil
	default:
		return recapv1.ActionCode_ACTION_CODE_UNSPECIFIED, fmt.Errorf(
			"%w: unknown action code %q",
			errInvalidProjection,
			value,
		)
	}
}

func cardTypeToProto(value model.CardType) (recapv1.CardType, error) {
	switch value {
	case model.CardIntro:
		return recapv1.CardType_CARD_TYPE_INTRO, nil
	case model.CardYearActivity:
		return recapv1.CardType_CARD_TYPE_YEAR_ACTIVITY, nil
	case model.CardTopCategory:
		return recapv1.CardType_CARD_TYPE_TOP_CATEGORY, nil
	case model.CardActiveMonth:
		return recapv1.CardType_CARD_TYPE_ACTIVE_MONTH, nil
	case model.CardBehavior:
		return recapv1.CardType_CARD_TYPE_BEHAVIOR, nil
	case model.CardAchievement:
		return recapv1.CardType_CARD_TYPE_ACHIEVEMENT, nil
	case model.CardMissedOpportunity:
		return recapv1.CardType_CARD_TYPE_MISSED_OPPORTUNITY, nil
	case model.CardNextAction:
		return recapv1.CardType_CARD_TYPE_NEXT_ACTION, nil
	case model.CardShare:
		return recapv1.CardType_CARD_TYPE_SHARE, nil
	default:
		return recapv1.CardType_CARD_TYPE_UNSPECIFIED, fmt.Errorf(
			"%w: unknown card type %q",
			errInvalidProjection,
			value,
		)
	}
}

func payloadTypeError(cardType model.CardType, payload model.CardPayload) error {
	return fmt.Errorf(
		"%w: card type %s has payload %T",
		errInvalidProjection,
		cardType,
		payload,
	)
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
