package analytics_test

import (
	"testing"

	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/model"
)

func FuzzEnrichMetricsRatesStayWithinUnitInterval(f *testing.F) {
	f.Add(uint64(0), uint64(0), uint64(0), uint64(0))
	f.Add(uint64(1), uint64(1), uint64(1), uint64(1))
	f.Add(uint64(100), uint64(40), uint64(10), uint64(3))
	f.Add(^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0))
	f.Fuzz(func(t *testing.T, views, uniqueRaw, chats, linkedChatsRaw uint64) {
		unique := boundUint64(uniqueRaw, views)
		linkedChats := boundUint64(linkedChatsRaw, chats)
		metrics := analytics.EnrichMetrics(model.Metrics{
			TotalViews: views, UniqueListings: unique, RepeatedViews: views - unique,
			ChatsStarted: chats, ChatsWithPurchase: linkedChats,
		})
		if metrics.RepeatRate < 0 || metrics.RepeatRate > 1 {
			t.Fatalf("repeat rate = %v", metrics.RepeatRate)
		}
		if metrics.PurchaseRate < 0 || metrics.PurchaseRate > 1 {
			t.Fatalf("purchase rate = %v", metrics.PurchaseRate)
		}
	})
}

func boundUint64(value, maximum uint64) uint64 {
	if maximum == ^uint64(0) {
		return value
	}
	return value % (maximum + 1)
}
