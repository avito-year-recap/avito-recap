package connect

import (
	"fmt"
	"time"

	recapv1 "github.com/year-recap/gen/go/recap/v1"
	"github.com/year-recap/internal/recap/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func profileToProto(value model.Profile) *recapv1.Profile {
	return &recapv1.Profile{
		Id:          value.ID.String(),
		Code:        value.Code,
		DisplayName: value.DisplayName,
		Description: value.Description,
		AvatarUrl:   optionalString(value.AvatarURL),
	}
}

func recapToProto(value model.Recap) (*recapv1.Recap, error) {
	period, err := periodToProto(value.Period)
	if err != nil {
		return nil, err
	}
	generatedAt, err := timestampToProto("generated_at", value.GeneratedAt)
	if err != nil {
		return nil, err
	}
	behavior, err := behaviorToProto(value.Behavior)
	if err != nil {
		return nil, err
	}
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
	cards := make([]*recapv1.Card, 0, len(value.Cards))
	for index, card := range value.Cards {
		mapped, err := cardToProto(card)
		if err != nil {
			return nil, fmt.Errorf("card %d: %w", index, err)
		}
		cards = append(cards, mapped)
	}
	return &recapv1.Recap{
		InternalId:   value.ID.String(),
		Profile:      profileToProto(value.Profile),
		Year:         value.Year,
		RulesVersion: value.RulesVersion,
		Metrics:      metricsToProto(value.Metrics),
		Behavior:     behavior,
		Achievements: achievements,
		Cards:        cards,
		NextAction:   nextAction,
		GeneratedAt:  generatedAt,
		ShareId:      value.ShareID.String(),
		Period:       period,
		RulesDigest:  value.RulesDigest,
	}, nil
}

func periodToProto(value model.RecapPeriod) (*recapv1.RecapPeriod, error) {
	startAt, err := timestampToProto("period.start_at", value.StartAt)
	if err != nil {
		return nil, err
	}
	endAt, err := timestampToProto("period.end_at", value.EndAt)
	if err != nil {
		return nil, err
	}
	return &recapv1.RecapPeriod{
		Year:    value.Year,
		StartAt: startAt,
		EndAt:   endAt,
		Final:   value.Final,
	}, nil
}

func metricsToProto(value model.Metrics) *recapv1.Metrics {
	activities := make([]*recapv1.CategoryActivity, 0, len(value.CategoryActivities))
	for _, activity := range value.CategoryActivities {
		activities = append(activities, &recapv1.CategoryActivity{
			CategoryCode:       activity.CategoryCode,
			Category:           activity.Category,
			Shareable:          activity.Shareable,
			Views:              activity.Views,
			FavoritesAdded:     activity.FavoritesAdded,
			PurchasesCompleted: activity.PurchasesCompleted,
		})
	}
	return &recapv1.Metrics{
		TotalEvents:          value.TotalEvents,
		Searches:             value.Searches,
		TotalViews:           value.TotalViews,
		UniqueListings:       value.UniqueListings,
		RepeatedViews:        value.RepeatedViews,
		FavoritesAdded:       value.FavoritesAdded,
		ChatsStarted:         value.ChatsStarted,
		ListingsCreated:      value.ListingsCreated,
		ListingsPublished:    value.ListingsPublished,
		PurchasesCompleted:   value.PurchasesCompleted,
		SalesCompleted:       value.SalesCompleted,
		ActiveDays:           value.ActiveDays,
		CategoriesCount:      value.CategoriesCount,
		TopCategoryCode:      optionalString(value.TopCategoryCode),
		TopCategory:          optionalString(value.TopCategory),
		TopCategoryViews:     value.TopCategoryViews,
		TopCategoryShareable: value.TopCategoryShareable,
		MostActiveMonth:      value.MostActiveMonth,
		RepeatRate:           value.RepeatRate,
		PurchaseRate:         value.PurchaseRate,
		ChatsWithPurchase:    value.ChatsWithPurchase,
		CategoryActivities:   activities,
	}
}

func behaviorToProto(value model.Behavior) (*recapv1.Behavior, error) {
	code, err := behaviorCodeToProto(value.Code)
	if err != nil {
		return nil, err
	}
	return &recapv1.Behavior{
		Code:        code,
		Title:       value.Title,
		Description: value.Description,
		Reason:      value.Reason,
		Evidence:    behaviorEvidenceToProto(value.Evidence),
	}, nil
}

func behaviorEvidenceToProto(values []model.BehaviorEvidence) []*recapv1.BehaviorEvidence {
	mapped := make([]*recapv1.BehaviorEvidence, 0, len(values))
	for _, value := range values {
		mapped = append(mapped, &recapv1.BehaviorEvidence{
			Metric:    value.Metric,
			Actual:    value.Actual,
			Threshold: value.Threshold,
			Detail:    value.Detail,
		})
	}
	return mapped
}

func achievementToProto(value model.Achievement) (*recapv1.Achievement, error) {
	code, err := achievementCodeToProto(value.Code)
	if err != nil {
		return nil, err
	}
	category, err := achievementCategoryToProto(value.Category)
	if err != nil {
		return nil, err
	}
	return &recapv1.Achievement{
		Code:        code,
		Title:       value.Title,
		Description: value.Description,
		Reason:      value.Reason,
		Shareable:   value.Shareable,
		Category:    category,
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
		Reason:      value.Reason,
		Target:      target,
	}, nil
}

func actionTargetToProto(value model.ActionTarget) (*recapv1.ActionTarget, error) {
	mapped := &recapv1.ActionTarget{}
	destinations := 0
	if value.Route != nil {
		destinations++
		mapped.Destination = &recapv1.ActionTarget_Route{
			Route: &recapv1.RouteTarget{Route: value.Route.Route},
		}
	}
	if value.Category != nil {
		destinations++
		mapped.Destination = &recapv1.ActionTarget_Category{
			Category: &recapv1.CategoryTarget{CategoryCode: value.Category.CategoryCode},
		}
	}
	if value.Listing != nil {
		destinations++
		mapped.Destination = &recapv1.ActionTarget_Listing{
			Listing: &recapv1.ListingTarget{ListingId: value.Listing.ListingID.String()},
		}
	}
	if value.Dialog != nil {
		destinations++
		mapped.Destination = &recapv1.ActionTarget_Dialog{
			Dialog: &recapv1.DialogTarget{DialogId: value.Dialog.DialogID.String()},
		}
	}
	if value.Search != nil {
		destinations++
		mapped.Destination = &recapv1.ActionTarget_Search{
			Search: &recapv1.SearchTarget{CategoryCode: value.Search.CategoryCode},
		}
	}
	if destinations != 1 {
		return nil, fmt.Errorf("%w: action target has %d destinations", errInvalidProjection, destinations)
	}
	return mapped, nil
}

func cardToProto(value model.Card) (*recapv1.Card, error) {
	cardType, err := cardTypeToProto(value.Type)
	if err != nil {
		return nil, err
	}
	payload, err := cardPayloadToProto(value.Type, value.Payload)
	if err != nil {
		return nil, err
	}
	return &recapv1.Card{
		Id:          value.ID,
		Type:        cardType,
		Position:    value.Position,
		Title:       value.Title,
		Description: value.Description,
		Explanation: optionalString(value.Explanation),
		Shareable:   value.Shareable,
		Payload:     payload,
	}, nil
}

func cardPayloadToProto(cardType model.CardType, payload model.CardPayload) (*recapv1.CardPayload, error) {
	switch cardType {
	case model.CardIntro:
		if payload != nil {
			return nil, fmt.Errorf("%w: INTRO card must not have a payload", errInvalidProjection)
		}
		return nil, nil
	case model.CardYearActivity:
		value, ok := payload.(model.YearActivityPayload)
		if !ok {
			return nil, payloadTypeError(cardType, payload)
		}
		return &recapv1.CardPayload{Value: &recapv1.CardPayload_YearActivity{
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
		}}, nil
	case model.CardTopCategory:
		value, ok := payload.(model.TopCategoryPayload)
		if !ok {
			return nil, payloadTypeError(cardType, payload)
		}
		return &recapv1.CardPayload{Value: &recapv1.CardPayload_TopCategory{
			TopCategory: &recapv1.TopCategoryPayload{
				CategoryCode:  value.CategoryCode,
				Category:      value.Category,
				CategoryViews: value.CategoryViews,
			},
		}}, nil
	case model.CardActiveMonth:
		value, ok := payload.(model.ActiveMonthPayload)
		if !ok {
			return nil, payloadTypeError(cardType, payload)
		}
		return &recapv1.CardPayload{Value: &recapv1.CardPayload_ActiveMonth{
			ActiveMonth: &recapv1.ActiveMonthPayload{Month: value.Month},
		}}, nil
	case model.CardBehavior:
		value, ok := payload.(model.BehaviorPayload)
		if !ok {
			return nil, payloadTypeError(cardType, payload)
		}
		code, err := behaviorCodeToProto(value.Code)
		if err != nil {
			return nil, err
		}
		return &recapv1.CardPayload{Value: &recapv1.CardPayload_Behavior{
			Behavior: &recapv1.BehaviorPayload{
				Code:     code,
				Evidence: behaviorEvidenceToProto(value.Evidence),
			},
		}}, nil
	case model.CardAchievement:
		value, ok := payload.(model.AchievementPayload)
		if !ok {
			return nil, payloadTypeError(cardType, payload)
		}
		codes := make([]recapv1.AchievementCode, 0, len(value.Codes))
		for _, code := range value.Codes {
			mapped, err := achievementCodeToProto(code)
			if err != nil {
				return nil, err
			}
			codes = append(codes, mapped)
		}
		return &recapv1.CardPayload{Value: &recapv1.CardPayload_Achievement{
			Achievement: &recapv1.AchievementPayload{Codes: codes},
		}}, nil
	case model.CardMissedOpportunity, model.CardNextAction:
		value, ok := payload.(model.ActionPayload)
		if !ok {
			return nil, payloadTypeError(cardType, payload)
		}
		code, err := actionCodeToProto(value.Code)
		if err != nil {
			return nil, err
		}
		target, err := actionTargetToProto(value.Target)
		if err != nil {
			return nil, err
		}
		return &recapv1.CardPayload{Value: &recapv1.CardPayload_Action{
			Action: &recapv1.ActionPayload{Code: code, Target: target},
		}}, nil
	case model.CardShare:
		value, ok := payload.(model.ShareCard)
		if !ok {
			return nil, payloadTypeError(cardType, payload)
		}
		return &recapv1.CardPayload{Value: &recapv1.CardPayload_Share{
			Share: shareCardToProto(value),
		}}, nil
	default:
		return nil, fmt.Errorf("%w: unknown card type %q", errInvalidProjection, cardType)
	}
}

func shareCardToProto(value model.ShareCard) *recapv1.ShareCard {
	return &recapv1.ShareCard{
		ShareId:          value.ShareID.String(),
		Year:             value.Year,
		BehaviorTitle:    value.BehaviorTitle,
		AchievementTitle: optionalString(value.AchievementTitle),
		TopCategory:      optionalString(value.TopCategory),
		PrivacyVersion:   value.PrivacyVersion,
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

func achievementCategoryToProto(
	value model.AchievementCategory,
) (recapv1.AchievementCategory, error) {
	switch value {
	case model.AchievementCategorySelling:
		return recapv1.AchievementCategory_ACHIEVEMENT_CATEGORY_SELLING, nil
	case model.AchievementCategoryBuying:
		return recapv1.AchievementCategory_ACHIEVEMENT_CATEGORY_BUYING, nil
	case model.AchievementCategoryDiscovery:
		return recapv1.AchievementCategory_ACHIEVEMENT_CATEGORY_DISCOVERY, nil
	case model.AchievementCategoryCollection:
		return recapv1.AchievementCategory_ACHIEVEMENT_CATEGORY_COLLECTION, nil
	case model.AchievementCategoryVersatility:
		return recapv1.AchievementCategory_ACHIEVEMENT_CATEGORY_VERSATILITY, nil
	case model.AchievementCategoryInterest:
		return recapv1.AchievementCategory_ACHIEVEMENT_CATEGORY_INTEREST, nil
	default:
		return recapv1.AchievementCategory_ACHIEVEMENT_CATEGORY_UNSPECIFIED, fmt.Errorf(
			"%w: unknown achievement category %q",
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

func timestampToProto(field string, value time.Time) (*timestamppb.Timestamp, error) {
	timestamp := timestamppb.New(value.UTC())
	if err := timestamp.CheckValid(); err != nil {
		return nil, fmt.Errorf("%w: %s: %w", errInvalidProjection, field, err)
	}
	return timestamp, nil
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
