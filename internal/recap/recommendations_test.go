package recap

import "testing"

func TestBuildNextActionAllBranches(t *testing.T) {
	tests := []struct {
		name     string
		metrics  Metrics
		expected ActionCode
	}{
		{
			name: "finish draft",
			metrics: Metrics{
				ListingsCreated:   4,
				ListingsPublished: 3,
			},
			expected: ActionFinishDraft,
		},
		{
			name: "improve listings",
			metrics: Metrics{
				ListingsCreated:   publishedForImprove,
				ListingsPublished: publishedForImprove,
			},
			expected: ActionImproveListings,
		},
		{
			name: "continue chats",
			metrics: Metrics{
				ChatsStarted: chatsForContinue,
			},
			expected: ActionContinueChats,
		},
		{
			name: "open favorites",
			metrics: Metrics{
				FavoritesAdded: favoritesForAction,
			},
			expected: ActionOpenFavorites,
		},
		{
			name: "open top category",
			metrics: Metrics{
				TopCategory: "Электроника",
			},
			expected: ActionOpenTopCategory,
		},
		{
			name:     "create first listing fallback",
			metrics:  Metrics{},
			expected: ActionCreateFirstListing,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := BuildNextAction(test.metrics)
			if actual.Code != test.expected {
				t.Fatalf("expected %s, got %s", test.expected, actual.Code)
			}
			if actual.Title == "" || actual.Description == "" || actual.ButtonText == "" || actual.Reason == "" {
				t.Fatalf("next action contains empty user-facing text: %+v", actual)
			}
		})
	}
}

func TestBuildNextActionPriority(t *testing.T) {
	metrics := Metrics{
		ListingsCreated:   5,
		ListingsPublished: 4,
		FavoritesAdded:    30,
		ChatsStarted:      10,
		TopCategory:       "Авто",
	}

	if actual := BuildNextAction(metrics); actual.Code != ActionFinishDraft {
		t.Fatalf("draft must have highest priority, got %s", actual.Code)
	}

	metrics.ListingsCreated = metrics.ListingsPublished
	if actual := BuildNextAction(metrics); actual.Code != ActionImproveListings {
		t.Fatalf("listing improvement must win after draft, got %s", actual.Code)
	}

	metrics.ListingsCreated = 0
	metrics.ListingsPublished = 0
	if actual := BuildNextAction(metrics); actual.Code != ActionContinueChats {
		t.Fatalf("chat continuation must win after seller actions, got %s", actual.Code)
	}
}
