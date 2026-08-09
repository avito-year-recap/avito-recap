package bootstrap

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
)

// eventsFromScenario deterministically expands a weighted seed scenario into
// the raw events a real AnalyticsStorage.CalculateMetrics aggregation would
// see. Event order carries no meaning; only the resulting aggregate counts
// do, and those are what seeds_test.go checks against the scenario's
// declared activity numbers.
func eventsFromScenario(profileID uuid.UUID, item scenario) ([]model.Event, error) {
	if item.Year == 0 || strings.TrimSpace(item.ProfileCode) == "" {
		return nil, fmt.Errorf("profileCode and year are required")
	}
	if item.Activity.UniqueListings > item.Activity.ListingViews {
		return nil, fmt.Errorf("unique listings exceed listing views")
	}
	if item.Activity.ChatsWithPurchase > item.Activity.ChatsStarted {
		return nil, fmt.Errorf("chats with purchase exceed chats started")
	}
	if item.Activity.ChatsWithPurchase > item.Activity.PurchasesCompleted {
		return nil, fmt.Errorf("chats with purchase exceed purchases completed")
	}
	if err := validateCategoryWeights(item.Categories); err != nil {
		return nil, err
	}
	if err := validateMonthWeights(item.Months); err != nil {
		return nil, err
	}

	schedule, err := dateSchedule(int(item.Year), item.Activity.ActiveDays, item.Months)
	if err != nil {
		return nil, err
	}
	nextTime := scheduleCycler(schedule)

	var events []model.Event
	emit := func(kind model.ActivityType, category string, adID, dialogID *uint64) {
		events = append(events, model.Event{
			ID:         uuid.New(),
			ProfileID:  profileID,
			Type:       kind,
			OccurredAt: nextTime(),
			Category:   category,
			AdID:       adID,
			DialogID:   dialogID,
		})
	}

	for i := uint64(0); i < item.Activity.Searches; i++ {
		emit(model.ActivitySearch, "", nil, nil)
	}

	views := splitByCategory(item.Activity.ListingViews, item.Categories, func(c weightedCategory) uint64 { return c.Views })
	adIDs := adIDPool(item.Activity.ListingViews, item.Activity.UniqueListings)
	adIndex := 0
	for _, code := range sortedCategoryCodes(views) {
		for i := uint64(0); i < views[code]; i++ {
			var adID *uint64
			if len(adIDs) > 0 {
				id := adIDs[adIndex%len(adIDs)]
				adIndex++
				adID = &id
			}
			emit(model.ActivityListingView, code, adID, nil)
		}
	}

	favorites := splitByCategory(item.Activity.FavoritesAdded, item.Categories, func(c weightedCategory) uint64 { return c.FavoritesAdded })
	for _, code := range sortedCategoryCodes(favorites) {
		for i := uint64(0); i < favorites[code]; i++ {
			emit(model.ActivityFavoriteAdded, code, nil, nil)
		}
	}

	dialogIDs := make([]uint64, item.Activity.ChatsStarted)
	for i := range dialogIDs {
		dialogIDs[i] = uint64(i + 1)
	}
	for _, dialogID := range dialogIDs {
		dialogID := dialogID
		emit(model.ActivityChatStarted, "", nil, &dialogID)
	}

	purchases := splitByCategory(item.Activity.PurchasesCompleted, item.Categories, func(c weightedCategory) uint64 { return c.PurchasesCompleted })
	linkedDialogs := uint64(0)
	for _, code := range sortedCategoryCodes(purchases) {
		for i := uint64(0); i < purchases[code]; i++ {
			var dialogID *uint64
			if linkedDialogs < item.Activity.ChatsWithPurchase && linkedDialogs < uint64(len(dialogIDs)) {
				id := dialogIDs[linkedDialogs]
				dialogID = &id
				linkedDialogs++
			}
			emit(model.ActivityPurchaseCompleted, code, nil, dialogID)
		}
	}

	for i := uint64(0); i < item.Activity.ListingsCreated; i++ {
		emit(model.ActivityListingCreated, "", nil, nil)
	}
	for i := uint64(0); i < item.Activity.ListingsPublished; i++ {
		emit(model.ActivityListingPublished, "", nil, nil)
	}
	for i := uint64(0); i < item.Activity.SalesCompleted; i++ {
		emit(model.ActivitySaleCompleted, "", nil, nil)
	}

	return events, nil
}

// splitByCategory resolves a per-category count for total: a category's own
// explicit count if it has one, otherwise its share of total by weight. Any
// rounding drift against total is reconciled into an uncategorized ("")
// bucket, or trimmed from the largest bucket if rounding overshot.
func splitByCategory(total uint64, categories []weightedCategory, explicit func(weightedCategory) uint64) map[string]uint64 {
	result := make(map[string]uint64, len(categories)+1)
	if total == 0 {
		return result
	}
	var assigned uint64
	for _, category := range categories {
		count := explicit(category)
		if count == 0 {
			count = weightedShare(total, category.Weight)
		}
		result[category.Code] = count
		assigned += count
	}
	reconcileToTotal(result, total, assigned)
	return result
}

func reconcileToTotal(counts map[string]uint64, total, assigned uint64) {
	if assigned == total {
		return
	}
	if assigned < total {
		counts[""] += total - assigned
		return
	}
	excess := assigned - total
	for excess > 0 {
		code, count := largestBucket(counts)
		if count == 0 {
			return
		}
		trim := excess
		if trim > count {
			trim = count
		}
		counts[code] -= trim
		excess -= trim
	}
}

func largestBucket(counts map[string]uint64) (string, uint64) {
	var bestCode string
	var bestCount uint64
	codes := sortedCategoryCodes(counts)
	for _, code := range codes {
		if counts[code] > bestCount {
			bestCount = counts[code]
			bestCode = code
		}
	}
	return bestCode, bestCount
}

func sortedCategoryCodes(counts map[string]uint64) []string {
	codes := make([]string, 0, len(counts))
	for code := range counts {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

// adIDPool returns uniqueListings distinct synthetic listing ids to round-robin
// view events across, so aggregating distinct AdIDs reproduces uniqueListings.
func adIDPool(totalViews, uniqueListings uint64) []uint64 {
	if uniqueListings == 0 || totalViews == 0 {
		return nil
	}
	pool := make([]uint64, uniqueListings)
	for i := range pool {
		pool[i] = uint64(i + 1)
	}
	return pool
}

func validateCategoryWeights(categories []weightedCategory) error {
	var sum uint64
	seen := make(map[string]struct{}, len(categories))
	for _, category := range categories {
		code := strings.TrimSpace(category.Code)
		if code == "" || category.Weight == 0 {
			return fmt.Errorf("category code and weight are required")
		}
		if _, exists := seen[code]; exists {
			return fmt.Errorf("duplicate category %q", code)
		}
		seen[code] = struct{}{}
		sum += uint64(category.Weight)
	}
	if len(categories) > 0 && sum != 100 {
		return fmt.Errorf("category weights sum to %d, want 100", sum)
	}
	return nil
}

func validateMonthWeights(months []weightedMonth) error {
	if len(months) == 0 {
		return fmt.Errorf("months are required")
	}
	var sum uint64
	seen := make(map[uint32]struct{}, len(months))
	for _, month := range months {
		if month.Month < 1 || month.Month > 12 || month.Weight == 0 {
			return fmt.Errorf("invalid month %d/weight %d", month.Month, month.Weight)
		}
		if _, exists := seen[month.Month]; exists {
			return fmt.Errorf("duplicate month %d", month.Month)
		}
		seen[month.Month] = struct{}{}
		sum += uint64(month.Weight)
	}
	if sum != 100 {
		return fmt.Errorf("month weights sum to %d, want 100", sum)
	}
	return nil
}

// dateSchedule returns exactly activeDays distinct calendar dates, split
// across months proportionally to weight (via weightedShare, the same
// rounding used everywhere else). Events are later assigned dates by cycling
// through this pool round-robin: since a month's share of the pool already
// mirrors its weight, cycling through it with even probability per date also
// gives that month its proportional share of event volume — which is what
// makes MostActiveMonth fall out of a plain aggregation over the events.
// (An earlier version repeated each date `weight` times to bias volume
// directly; with TotalEvents often smaller than that inflated pool, cycling
// never reached the later, more heavily repeated months at all.)
func dateSchedule(year int, activeDays uint64, months []weightedMonth) ([]time.Time, error) {
	if activeDays == 0 {
		return nil, fmt.Errorf("activeDays must be positive")
	}
	daysPerMonth := make(map[uint32]uint64, len(months))
	var assigned uint64
	for _, month := range months {
		count := weightedShare(activeDays, month.Weight)
		daysPerMonth[month.Month] = count
		assigned += count
	}
	reconcileDaysToTotal(daysPerMonth, activeDays, assigned)

	capacity := make(map[uint32]uint64, len(months))
	for _, month := range months {
		capacity[month.Month] = uint64(time.Date(year, time.Month(month.Month)+1, 0, 0, 0, 0, 0, time.UTC).Day())
	}
	clampToCapacity(daysPerMonth, capacity, months)

	var pool []time.Time
	for _, month := range months {
		pool = append(pool, pickDaysInMonth(year, time.Month(month.Month), daysPerMonth[month.Month])...)
	}
	if len(pool) == 0 {
		return nil, fmt.Errorf("date schedule produced no active days")
	}
	return pool, nil
}

// clampToCapacity caps each month at its real day count (weightedShare has no
// notion of "December only has 31 days") and moves any shortfall into months
// that still have room, so the total stays exactly activeDays whenever the
// scenario's months have enough combined capacity for it.
func clampToCapacity(daysPerMonth, capacity map[uint32]uint64, months []weightedMonth) {
	var shortfall uint64
	for _, month := range months {
		if daysPerMonth[month.Month] > capacity[month.Month] {
			shortfall += daysPerMonth[month.Month] - capacity[month.Month]
			daysPerMonth[month.Month] = capacity[month.Month]
		}
	}
	for _, month := range months {
		if shortfall == 0 {
			break
		}
		room := capacity[month.Month] - daysPerMonth[month.Month]
		if room == 0 {
			continue
		}
		add := shortfall
		if add > room {
			add = room
		}
		daysPerMonth[month.Month] += add
		shortfall -= add
	}
}

func reconcileDaysToTotal(daysPerMonth map[uint32]uint64, total, assigned uint64) {
	if assigned == total {
		return
	}
	months := make([]uint32, 0, len(daysPerMonth))
	for month := range daysPerMonth {
		months = append(months, month)
	}
	sort.Slice(months, func(i, j int) bool { return months[i] < months[j] })
	if assigned < total {
		daysPerMonth[months[0]] += total - assigned
		return
	}
	excess := assigned - total
	for _, month := range months {
		if excess == 0 {
			break
		}
		trim := excess
		if trim > daysPerMonth[month] {
			trim = daysPerMonth[month]
		}
		daysPerMonth[month] -= trim
		excess -= trim
	}
}

func pickDaysInMonth(year int, month time.Month, count uint64) []time.Time {
	if count == 0 {
		return nil
	}
	daysInMonth := uint64(time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day())
	if count > daysInMonth {
		count = daysInMonth
	}
	step := float64(daysInMonth) / float64(count)
	result := make([]time.Time, 0, count)
	seen := make(map[int]struct{}, count)
	for i := uint64(0); i < count; i++ {
		day := int(float64(i)*step) + 1
		if day > int(daysInMonth) {
			day = int(daysInMonth)
		}
		if _, exists := seen[day]; exists {
			continue
		}
		seen[day] = struct{}{}
		result = append(result, time.Date(year, month, day, 12, 0, 0, 0, time.UTC))
	}
	return result
}

func scheduleCycler(schedule []time.Time) func() time.Time {
	index := 0
	return func() time.Time {
		value := schedule[index%len(schedule)]
		index++
		return value
	}
}

func weightedShare(total uint64, weight uint32) uint64 {
	if total == 0 || weight == 0 {
		return 0
	}
	value := math.Round(float64(total) * float64(weight) / 100)
	if value < 1 {
		return 1
	}
	return uint64(value)
}
