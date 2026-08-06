package ruleset

import "testing"

func TestDefaultRulesetIsValidAndStable(t *testing.T) {
	value := DefaultRuleset()
	if err := value.Validate(); err != nil {
		t.Fatalf("default ruleset: %v", err)
	}
	if value.Version != CurrentRulesVersion {
		t.Fatalf("version=%q", value.Version)
	}
	if len(value.Digest()) != 64 {
		t.Fatalf("digest=%q", value.Digest())
	}
}

func TestDigestChangesWithMaterialRule(t *testing.T) {
	left := DefaultRuleset()
	right := DefaultRuleset()
	right.Thresholds.FindHunterMinViews++
	if left.Digest() == right.Digest() {
		t.Fatal("digest must bind material configuration")
	}
}
