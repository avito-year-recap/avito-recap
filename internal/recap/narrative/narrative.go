package narrative

import (
	"context"
	"fmt"
	"strings"

	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/presentation/cards"
	"github.com/year-recap/internal/recap/validation/structural"
)

const MaxDescriptionRunes = structural.MaxNarrativeDescriptionRunes

// Generator turns already-derived, privacy-safe recap facts into short copy.
// It must never decide behavior, achievements, thresholds, or next actions.
type Generator interface {
	Generate(ctx context.Context, facts Facts) (Story, error)
}

// Enricher is the application-facing narrative boundary. Implementations may
// improve presentation copy, but must preserve the deterministic recap when AI
// is unavailable or returns invalid output.
type Enricher interface {
	Enrich(ctx context.Context, recap model.Recap) (model.Recap, error)
}

// Facts is the only data an AI narrative provider receives. It deliberately
// excludes profile UUIDs, listing/dialog IDs, raw events, messages, and exact
// purchase objects. The rule engine remains the source of truth.
type Facts struct {
	Year            uint32            `json:"year"`
	Metrics         MetricFacts       `json:"metrics"`
	TopCategory     string            `json:"topCategory,omitempty"`
	MostActiveMonth string            `json:"mostActiveMonth,omitempty"`
	Behavior        BehaviorFacts     `json:"behavior"`
	Achievements    []AchievementFact `json:"achievements,omitempty"`
	NextAction      ActionFacts       `json:"nextAction"`
	EditableCardIDs []string          `json:"editableCardIds"`
}

type MetricFacts struct {
	TotalEvents        uint64  `json:"totalEvents"`
	Searches           uint64  `json:"searches"`
	TotalViews         uint64  `json:"totalViews"`
	FavoritesAdded     uint64  `json:"favoritesAdded"`
	ChatsStarted       uint64  `json:"chatsStarted"`
	ListingsPublished  uint64  `json:"listingsPublished"`
	PurchasesCompleted uint64  `json:"purchasesCompleted"`
	SalesCompleted     uint64  `json:"salesCompleted"`
	ActiveDays         uint64  `json:"activeDays"`
	CategoriesCount    uint64  `json:"categoriesCount"`
	RepeatRatePercent  float64 `json:"repeatRatePercent"`
	PurchaseRatePct    float64 `json:"purchaseRatePercent"`
}

type BehaviorFacts struct {
	Code        model.BehaviorCode `json:"code"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
}

type AchievementFact struct {
	Code   model.AchievementCode `json:"code"`
	Title  string                `json:"title"`
	Reason string                `json:"reason"`
}

type ActionFacts struct {
	Code        model.ActionCode `json:"code"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Reason      string           `json:"reason"`
}

// Story contains copy overrides only. IDs/types/payloads/explanations are not
// writable by AI, so the generated text cannot alter executable behavior.
type Story struct {
	Cards []CardNarrative `json:"cards"`
}

type CardNarrative struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

func editableCardIDs(recap model.Recap) []string {
	ids := make([]string, 0, len(recap.Cards))
	for _, card := range recap.Cards {
		if model.IsNarrativeEditableCardType(card.Type) {
			ids = append(ids, card.ID)
		}
	}
	return ids
}

func FactsFromRecap(recap model.Recap) Facts {
	achievementFacts := make([]AchievementFact, 0, len(recap.Achievements))
	for _, achievement := range recap.Achievements {
		achievementFacts = append(achievementFacts, AchievementFact{
			Code: achievement.Code, Title: achievement.Title, Reason: achievement.Reason,
		})
	}

	month := ""
	if recap.Metrics.MostActiveMonth >= 1 && recap.Metrics.MostActiveMonth <= 12 {
		month = cards.MonthName(recap.Metrics.MostActiveMonth)
	}

	return Facts{
		Year: recap.Year,
		Metrics: MetricFacts{
			TotalEvents: recap.Metrics.TotalEvents, Searches: recap.Metrics.Searches,
			TotalViews: recap.Metrics.TotalViews, FavoritesAdded: recap.Metrics.FavoritesAdded,
			ChatsStarted: recap.Metrics.ChatsStarted, ListingsPublished: recap.Metrics.ListingsPublished,
			PurchasesCompleted: recap.Metrics.PurchasesCompleted, SalesCompleted: recap.Metrics.SalesCompleted,
			ActiveDays: recap.Metrics.ActiveDays, CategoriesCount: recap.Metrics.CategoriesCount,
			RepeatRatePercent: recap.Metrics.RepeatRate * 100,
			PurchaseRatePct:   recap.Metrics.PurchaseRate * 100,
		},
		TopCategory: recap.Metrics.TopCategory, MostActiveMonth: month,
		Behavior:     BehaviorFacts{Code: recap.Behavior.Code, Title: recap.Behavior.Title, Description: recap.Behavior.Description},
		Achievements: achievementFacts,
		NextAction: ActionFacts{
			Code: recap.NextAction.Code, Title: recap.NextAction.Title,
			Description: recap.NextAction.Description, Reason: recap.NextAction.Reason,
		},
		EditableCardIDs: editableCardIDs(recap),
	}
}

// Apply validates the complete AI copy projection and only replaces
// descriptions of explicitly AI-editable cards. A non-empty Story must contain
// exactly one entry for every editable card; partial/duplicate/SHARE overrides
// are rejected atomically and the deterministic recap remains untouched.
func Apply(recap model.Recap, story Story) (model.Recap, error) {
	allowed := make(map[string]int, len(recap.Cards))
	for i, card := range recap.Cards {
		if model.IsNarrativeEditableCardType(card.Type) {
			allowed[card.ID] = i
		}
	}
	if len(story.Cards) != len(allowed) {
		return recap, fmt.Errorf("narrative returned %d cards, expected exactly %d editable cards", len(story.Cards), len(allowed))
	}
	if len(allowed) == 0 {
		return recap, nil
	}

	seen := make(map[string]struct{}, len(story.Cards))
	result := recap
	result.Cards = append([]model.Card(nil), recap.Cards...)

	for _, generated := range story.Cards {
		generated.ID = strings.TrimSpace(generated.ID)
		generated.Description = strings.TrimSpace(generated.Description)

		index, ok := allowed[generated.ID]
		if !ok {
			return recap, fmt.Errorf("narrative returned non-editable or unknown card id %q", generated.ID)
		}
		if _, exists := seen[generated.ID]; exists {
			return recap, fmt.Errorf("narrative returned duplicate card id %q", generated.ID)
		}
		seen[generated.ID] = struct{}{}

		if err := structural.ValidateNarrativeDescription(generated.Description); err != nil {
			return recap, fmt.Errorf("narrative description for %q: %w", generated.ID, err)
		}
		result.Cards[index].Description = generated.Description
	}

	return result, nil
}
