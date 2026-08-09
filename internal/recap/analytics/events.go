package analytics

import (
	"sort"

	"github.com/year-recap/internal/recap/model"
)

type categoryTotals struct {
	views     uint64
	favorites uint64
	purchases uint64
}

// AggregateEvents turns a profile's raw yearly events into the same Metrics
// contract CalculateMetrics returns. It is the single place that decides what
// a "top category" or "most active month" means, so the storage adapter only
// has to fetch rows and this stays testable without a database.
func AggregateEvents(events []model.Event) model.Metrics {
	var metrics model.Metrics
	metrics.TotalEvents = uint64(len(events))

	categories := make(map[string]*categoryTotals)
	viewedListings := make(map[uint64]struct{})
	chatDialogs := make(map[uint64]struct{})
	purchaseDialogs := make(map[uint64]struct{})
	activeDates := make(map[string]struct{})
	monthCounts := [13]uint64{} // index 1..12

	categoryTotal := func(code string) *categoryTotals {
		totals, ok := categories[code]
		if !ok {
			totals = &categoryTotals{}
			categories[code] = totals
		}
		return totals
	}

	for _, event := range events {
		occurredAt := event.OccurredAt.UTC()
		activeDates[occurredAt.Format("2006-01-02")] = struct{}{}
		monthCounts[occurredAt.Month()]++

		switch event.Type {
		case model.ActivitySearch:
			metrics.Searches++
		case model.ActivityListingView:
			metrics.TotalViews++
			if event.AdID != nil {
				viewedListings[*event.AdID] = struct{}{}
			}
			if event.Category != "" {
				categoryTotal(event.Category).views++
			}
		case model.ActivityFavoriteAdded:
			metrics.FavoritesAdded++
			if event.Category != "" {
				categoryTotal(event.Category).favorites++
			}
		case model.ActivityChatStarted:
			metrics.ChatsStarted++
			if event.DialogID != nil {
				chatDialogs[*event.DialogID] = struct{}{}
			}
		case model.ActivityListingCreated:
			metrics.ListingsCreated++
		case model.ActivityListingPublished:
			metrics.ListingsPublished++
		case model.ActivityPurchaseCompleted:
			metrics.PurchasesCompleted++
			if event.Category != "" {
				categoryTotal(event.Category).purchases++
			}
			if event.DialogID != nil {
				purchaseDialogs[*event.DialogID] = struct{}{}
			}
		case model.ActivitySaleCompleted:
			metrics.SalesCompleted++
		}
	}

	metrics.UniqueListings = uint64(len(viewedListings))
	if metrics.TotalViews >= metrics.UniqueListings {
		metrics.RepeatedViews = metrics.TotalViews - metrics.UniqueListings
	}
	metrics.ActiveDays = uint64(len(activeDates))
	metrics.CategoriesCount = uint64(len(categories))

	for dialogID := range chatDialogs {
		if _, purchased := purchaseDialogs[dialogID]; purchased {
			metrics.ChatsWithPurchase++
		}
	}

	var bestMonth uint32
	var bestMonthCount uint64
	for month := uint32(1); month <= 12; month++ {
		if count := monthCounts[month]; count > bestMonthCount {
			bestMonthCount = count
			bestMonth = month
		}
	}
	metrics.MostActiveMonth = bestMonth

	codes := make([]string, 0, len(categories))
	for code := range categories {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	var topCode string
	var topViews uint64
	metrics.CategoryActivities = make([]model.CategoryActivity, 0, len(codes))
	for _, code := range codes {
		totals := categories[code]
		definition, _ := CategoryByCode(code)
		metrics.CategoryActivities = append(metrics.CategoryActivities, model.CategoryActivity{
			CategoryCode:       code,
			Category:           definition.Title,
			Shareable:          definition.Shareable,
			Views:              totals.views,
			FavoritesAdded:     totals.favorites,
			PurchasesCompleted: totals.purchases,
		})
		if totals.views > topViews {
			topViews = totals.views
			topCode = code
		}
	}
	if topCode != "" {
		definition, _ := CategoryByCode(topCode)
		metrics.TopCategoryCode = topCode
		metrics.TopCategory = definition.Title
		metrics.TopCategoryShareable = definition.Shareable
		metrics.TopCategoryViews = topViews
	}

	return EnrichMetrics(metrics)
}
