package behavior

import (
	"testing"

	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

func TestDetectSelectsExpectedPersonas(t *testing.T) {
	thresholds := ruleset.DefaultRuleset().Thresholds
	tests := []struct {
		name    string
		metrics model.Metrics
		want    model.BehaviorCode
	}{
		{"active seller", model.Metrics{ListingsPublished: thresholds.ActiveSellerMinListings, SalesCompleted: thresholds.ActiveSellerMinDeals}, model.BehaviorActiveSeller},
		{"starting seller", model.Metrics{ListingsCreated: thresholds.StartingSellerMinCreated, ListingsPublished: thresholds.StartingSellerMaxPublished}, model.BehaviorStartingSeller},
		{"decisive buyer", model.Metrics{PurchasesCompleted: thresholds.DecisiveBuyerMinPurchases, ChatsStarted: thresholds.DecisiveBuyerMinChats, ChatsWithPurchase: thresholds.DecisiveBuyerMinLinkedChats}, model.BehaviorDecisiveBuyer},
		{"find hunter", model.Metrics{TotalViews: 20, UniqueListings: 16, RepeatedViews: 4, FavoritesAdded: 3}, model.BehaviorFindHunter},
		{"fallback", model.Metrics{}, model.BehaviorUniversal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Detect(tc.metrics)
			if got.Code != tc.want {
				t.Fatalf("got %s want %s", got.Code, tc.want)
			}
			if got.Title == "" || got.Reason == "" {
				t.Fatalf("incomplete: %+v", got)
			}
		})
	}
}
