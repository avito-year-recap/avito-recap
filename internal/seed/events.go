package seed

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/model"
)

// EventsFromScenario is the canonical seed expansion path used by both the
// in-memory demo adapter and ClickHouse bootstrap. Keeping one implementation
// guarantees that the same profile/year produces the same Metrics regardless
// of the selected storage backend.
func EventsFromScenario(profileID uuid.UUID, scenario Scenario) ([]model.Event, error) {
	if err := validateScenario(scenario); err != nil {
		return nil, err
	}

	totalEvents, err := analytics.SumUint64(
		scenario.Activity.Searches,
		scenario.Activity.ListingViews,
		scenario.Activity.FavoritesAdded,
		scenario.Activity.ChatsStarted,
		scenario.Activity.ListingsCreated,
		scenario.Activity.ListingsPublished,
		scenario.Activity.PurchasesCompleted,
		scenario.Activity.SalesCompleted,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: total event count: %w", ErrInvalidScenario, err)
	}
	if totalEvents == 0 {
		return nil, fmt.Errorf("%w: scenario must contain at least one event", ErrInvalidScenario)
	}
	if scenario.Activity.ActiveDays > totalEvents {
		return nil, fmt.Errorf(
			"%w: active days %d exceed total events %d",
			ErrInvalidScenario,
			scenario.Activity.ActiveDays,
			totalEvents,
		)
	}

	schedule, err := eventSchedule(int(scenario.Year), totalEvents, scenario.Activity.ActiveDays, scenario.Months)
	if err != nil {
		return nil, fmt.Errorf("%w: build event schedule: %w", ErrInvalidScenario, err)
	}
	nextTime := scheduleCycler(schedule)

	events := make([]model.Event, 0, totalEvents)
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

	for i := uint64(0); i < scenario.Activity.Searches; i++ {
		emit(model.ActivitySearch, "", nil, nil)
	}

	views, err := splitByCategory(
		scenario.Activity.ListingViews,
		scenario.Categories,
		func(category WeightedCategory) *uint64 { return category.Views },
		"views",
	)
	if err != nil {
		return nil, err
	}
	adIDs := adIDPool(scenario.Activity.ListingViews, scenario.Activity.UniqueListings)
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

	favorites, err := splitByCategory(
		scenario.Activity.FavoritesAdded,
		scenario.Categories,
		func(category WeightedCategory) *uint64 { return category.FavoritesAdded },
		"favoritesAdded",
	)
	if err != nil {
		return nil, err
	}
	for _, code := range sortedCategoryCodes(favorites) {
		for i := uint64(0); i < favorites[code]; i++ {
			emit(model.ActivityFavoriteAdded, code, nil, nil)
		}
	}

	dialogIDs := make([]uint64, scenario.Activity.ChatsStarted)
	for i := range dialogIDs {
		dialogIDs[i] = uint64(i + 1)
	}
	for _, dialogID := range dialogIDs {
		dialogID := dialogID
		emit(model.ActivityChatStarted, "", nil, &dialogID)
	}

	purchases, err := splitByCategory(
		scenario.Activity.PurchasesCompleted,
		scenario.Categories,
		func(category WeightedCategory) *uint64 { return category.PurchasesCompleted },
		"purchasesCompleted",
	)
	if err != nil {
		return nil, err
	}
	linkedDialogs := uint64(0)
	for _, code := range sortedCategoryCodes(purchases) {
		for i := uint64(0); i < purchases[code]; i++ {
			var dialogID *uint64
			if linkedDialogs < scenario.Activity.ChatsWithPurchase && linkedDialogs < uint64(len(dialogIDs)) {
				id := dialogIDs[linkedDialogs]
				dialogID = &id
				linkedDialogs++
			}
			emit(model.ActivityPurchaseCompleted, code, nil, dialogID)
		}
	}

	for i := uint64(0); i < scenario.Activity.ListingsCreated; i++ {
		emit(model.ActivityListingCreated, "", nil, nil)
	}
	for i := uint64(0); i < scenario.Activity.ListingsPublished; i++ {
		emit(model.ActivityListingPublished, "", nil, nil)
	}
	for i := uint64(0); i < scenario.Activity.SalesCompleted; i++ {
		emit(model.ActivitySaleCompleted, "", nil, nil)
	}

	if uint64(len(events)) != totalEvents {
		return nil, fmt.Errorf(
			"%w: generated %d events, expected %d",
			ErrInvalidScenario,
			len(events),
			totalEvents,
		)
	}
	return events, nil
}

func validateScenario(scenario Scenario) error {
	if strings.TrimSpace(scenario.ProfileCode) == "" {
		return fmt.Errorf("%w: profile code is required", ErrInvalidScenario)
	}
	if scenario.Year == 0 {
		return fmt.Errorf("%w: year is required", ErrInvalidScenario)
	}
	if scenario.Activity.UniqueListings > scenario.Activity.ListingViews {
		return fmt.Errorf("%w: unique listings exceed views", ErrInvalidScenario)
	}
	if scenario.Activity.ChatsWithPurchase > scenario.Activity.ChatsStarted {
		return fmt.Errorf("%w: chats with purchase exceed started chats", ErrInvalidScenario)
	}
	if scenario.Activity.ChatsWithPurchase > scenario.Activity.PurchasesCompleted {
		return fmt.Errorf("%w: chats with purchase exceed completed purchases", ErrInvalidScenario)
	}
	if scenario.Activity.ActiveDays == 0 {
		return fmt.Errorf("%w: activeDays must be positive", ErrInvalidScenario)
	}
	if err := validateCategoryWeights(scenario.Categories); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidScenario, err)
	}
	if err := validateMonthWeights(scenario.Months); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidScenario, err)
	}
	return nil
}

// splitByCategory has two unambiguous modes:
//   - if no category explicitly specifies the metric, distribute the total by
//     category weights;
//   - if at least one category specifies it, unspecified categories mean zero
//     and the explicit counts must sum exactly to total.
//
// Pointer fields in WeightedCategory are intentional: JSON omission and an
// explicit zero are different product meanings and must not be conflated.
func splitByCategory(
	total uint64,
	categories []WeightedCategory,
	explicit func(WeightedCategory) *uint64,
	metricName string,
) (map[string]uint64, error) {
	result := make(map[string]uint64, len(categories))
	if total == 0 {
		for _, category := range categories {
			if value := explicit(category); value != nil && *value != 0 {
				return nil, fmt.Errorf(
					"%w: category %q has explicit %s=%d while total is zero",
					ErrInvalidScenario,
					category.Code,
					metricName,
					*value,
				)
			}
		}
		return result, nil
	}
	if len(categories) == 0 {
		return map[string]uint64{"": total}, nil
	}

	hasExplicit := false
	for _, category := range categories {
		if explicit(category) != nil {
			hasExplicit = true
			break
		}
	}
	if hasExplicit {
		var assigned uint64
		for _, category := range categories {
			count := uint64(0)
			if value := explicit(category); value != nil {
				count = *value
			}
			result[category.Code] = count
			assigned += count
		}
		if assigned != total {
			return nil, fmt.Errorf(
				"%w: explicit category %s counts sum to %d, want %d",
				ErrInvalidScenario,
				metricName,
				assigned,
				total,
			)
		}
		return result, nil
	}

	weights := make([]weightedAllocation, 0, len(categories))
	for _, category := range categories {
		weights = append(weights, weightedAllocation{key: category.Code, weight: category.Weight})
	}
	return allocateWeighted(total, weights), nil
}

type weightedAllocation struct {
	key    string
	weight uint32
}

// allocateWeighted implements the largest-remainder method, so all integer
// buckets sum exactly to total without the rounding drift of independent
// math.Round calls. Equal remainders prefer the heavier bucket, then key, for reproducibility.
func allocateWeighted(total uint64, weights []weightedAllocation) map[string]uint64 {
	result := make(map[string]uint64, len(weights))
	if total == 0 || len(weights) == 0 {
		return result
	}

	type remainder struct {
		key       string
		weight    uint32
		remainder uint64
	}
	remainders := make([]remainder, 0, len(weights))
	var assigned uint64
	for _, item := range weights {
		// Split before multiplying to avoid uint64 overflow even for very large
		// synthetic totals. Weight is validated as part of a 100% distribution.
		whole := total / 100
		fraction := total % 100
		base := whole*uint64(item.weight) + (fraction*uint64(item.weight))/100
		rem := (fraction * uint64(item.weight)) % 100
		result[item.key] = base
		assigned += base
		remainders = append(remainders, remainder{key: item.key, weight: item.weight, remainder: rem})
	}
	sort.Slice(remainders, func(i, j int) bool {
		if remainders[i].remainder != remainders[j].remainder {
			return remainders[i].remainder > remainders[j].remainder
		}
		if remainders[i].weight != remainders[j].weight {
			return remainders[i].weight > remainders[j].weight
		}
		return remainders[i].key < remainders[j].key
	})
	leftover := total - assigned
	for i := uint64(0); i < leftover; i++ {
		result[remainders[i%uint64(len(remainders))].key]++
	}
	return result
}

func eventSchedule(year int, totalEvents, activeDays uint64, months []WeightedMonth) ([]time.Time, error) {
	if totalEvents == 0 || activeDays == 0 {
		return nil, fmt.Errorf("totalEvents and activeDays must be positive")
	}
	if activeDays > totalEvents {
		return nil, fmt.Errorf("activeDays %d exceed totalEvents %d", activeDays, totalEvents)
	}

	monthWeights := make([]weightedAllocation, 0, len(months))
	monthByKey := make(map[string]WeightedMonth, len(months))
	for _, month := range months {
		key := fmt.Sprintf("%02d", month.Month)
		monthWeights = append(monthWeights, weightedAllocation{key: key, weight: month.Weight})
		monthByKey[key] = month
	}
	eventCounts := allocateWeighted(totalEvents, monthWeights)
	dayCounts := allocateWeighted(activeDays, monthWeights)

	capacities := make(map[string]uint64, len(months))
	var capacityTotal uint64
	for _, item := range monthWeights {
		month := monthByKey[item.key]
		calendarDays := uint64(time.Date(year, time.Month(month.Month)+1, 0, 0, 0, 0, 0, time.UTC).Day())
		capacity := minUint64(calendarDays, eventCounts[item.key])
		capacities[item.key] = capacity
		capacityTotal += capacity
	}
	if capacityTotal < activeDays {
		return nil, fmt.Errorf("selected months can represent at most %d active days, want %d", capacityTotal, activeDays)
	}

	var overflow uint64
	for key, count := range dayCounts {
		if count > capacities[key] {
			overflow += count - capacities[key]
			dayCounts[key] = capacities[key]
		}
	}
	if overflow > 0 {
		// Prefer heavier months when redistributing capped active days; month is
		// the deterministic tie-breaker.
		sorted := append([]weightedAllocation(nil), monthWeights...)
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].weight != sorted[j].weight {
				return sorted[i].weight > sorted[j].weight
			}
			return sorted[i].key < sorted[j].key
		})
		for overflow > 0 {
			progress := false
			for _, item := range sorted {
				if overflow == 0 {
					break
				}
				if dayCounts[item.key] >= capacities[item.key] {
					continue
				}
				dayCounts[item.key]++
				overflow--
				progress = true
			}
			if !progress {
				return nil, fmt.Errorf("could not redistribute active-day overflow")
			}
		}
	}

	keys := make([]string, 0, len(monthByKey))
	for key := range monthByKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	schedule := make([]time.Time, 0, totalEvents)
	for _, key := range keys {
		month := monthByKey[key]
		dates := pickDaysInMonth(year, time.Month(month.Month), dayCounts[key])
		count := eventCounts[key]
		if count == 0 {
			continue
		}
		if len(dates) == 0 {
			return nil, fmt.Errorf("month %d has %d events but no active date", month.Month, count)
		}
		for i := uint64(0); i < count; i++ {
			schedule = append(schedule, dates[i%uint64(len(dates))])
		}
	}
	if uint64(len(schedule)) != totalEvents {
		return nil, fmt.Errorf("schedule has %d events, want %d", len(schedule), totalEvents)
	}
	return schedule, nil
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
		day := int(math.Floor(float64(i)*step)) + 1
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

func sortedCategoryCodes(counts map[string]uint64) []string {
	codes := make([]string, 0, len(counts))
	for code := range counts {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

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

func validateCategoryWeights(categories []WeightedCategory) error {
	if len(categories) == 0 {
		return nil
	}
	var sum uint64
	seen := make(map[string]struct{}, len(categories))
	for _, category := range categories {
		code := strings.TrimSpace(category.Code)
		if code == "" || strings.TrimSpace(category.Title) == "" || category.Weight == 0 {
			return fmt.Errorf("category code, title and positive weight are required")
		}
		if _, exists := seen[code]; exists {
			return fmt.Errorf("duplicate category %q", code)
		}
		seen[code] = struct{}{}
		sum += uint64(category.Weight)
	}
	if sum != 100 {
		return fmt.Errorf("category weights sum to %d, want 100", sum)
	}
	return nil
}

func validateMonthWeights(months []WeightedMonth) error {
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

func minUint64(left, right uint64) uint64 {
	if left < right {
		return left
	}
	return right
}
