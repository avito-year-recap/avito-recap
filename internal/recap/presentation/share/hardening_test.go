package share_test

import (
	"testing"

	"github.com/year-recap/internal/recap/presentation/share"
	"github.com/year-recap/internal/recap/ruleset"
	"github.com/year-recap/internal/recap/testkit"
)

func TestPrivacyProjectionRejectsUnsafeExternalText(t *testing.T) {
	value := testkit.Recap()
	value.Metrics.TopCategory = "Электроника\nprivate-token"
	value.Metrics.TopCategoryShareable = true
	card := share.BuildWithRuleset(ruleset.DefaultRuleset(), value)
	if card.TopCategory != "" {
		t.Fatalf("unsafe category leaked: %q", card.TopCategory)
	}
	if card.PrivacyVersion != ruleset.DefaultRuleset().SharePolicy.Version {
		t.Fatalf("privacy policy version missing: %+v", card)
	}
}

func TestPrivacyProjectionRejectsSensitiveCatalogueCategory(t *testing.T) {
	value := testkit.Recap()
	value.Metrics.TopCategoryCode = "pets"
	value.Metrics.TopCategory = "Товары для животных"
	value.Metrics.TopCategoryShareable = true // untrusted upstream flag

	card := share.BuildWithRuleset(ruleset.DefaultRuleset(), value)
	if card.TopCategory != "" {
		t.Fatalf("sensitive catalogue category leaked publicly: %q", card.TopCategory)
	}
}
