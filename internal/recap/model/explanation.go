package model

// RuleComparison describes how a metric is compared with a rule boundary.
type RuleComparison string

const (
	RuleComparisonGTE RuleComparison = "GTE"
	RuleComparisonLTE RuleComparison = "LTE"
	RuleComparisonGT  RuleComparison = "GT"
	RuleComparisonEQ  RuleComparison = "EQ"
)

// RuleCheck is a single transparent condition evaluated by the personalization engine.
type RuleCheck struct {
	Metric      string         `json:"metric"`
	Actual      float64        `json:"actual"`
	Threshold   float64        `json:"threshold"`
	Comparison  RuleComparison `json:"comparison"`
	Passed      bool           `json:"passed"`
	Explanation string         `json:"explanation"`
}

// BehaviorRuleEvaluation shows why a behavior candidate matched or failed.
type BehaviorRuleEvaluation struct {
	Code     BehaviorCode `json:"code"`
	Priority uint32       `json:"priority"`
	Matched  bool         `json:"matched"`
	Selected bool         `json:"selected"`
	Checks   []RuleCheck  `json:"checks"`
}

// ActionRuleEvaluation exposes the recommendation table without leaking private targets.
type ActionRuleEvaluation struct {
	Name     string     `json:"name"`
	Code     ActionCode `json:"code"`
	Priority int        `json:"priority"`
	Matched  bool       `json:"matched"`
	Selected bool       `json:"selected"`
}

// AchievementExplanation contains only the safe reasoning behind an awarded achievement.
type AchievementExplanation struct {
	Code      AchievementCode `json:"code"`
	Title     string          `json:"title"`
	Reason    string          `json:"reason"`
	Shareable bool            `json:"shareable"`
}

// NextActionExplanation contains the product reasoning without the executable target.
// This keeps the debug endpoint useful without exposing listing/dialog identifiers.
type NextActionExplanation struct {
	Code        ActionCode `json:"code"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ButtonText  string     `json:"buttonText"`
	Reason      string     `json:"reason"`
}

// RecapExplanation is a read-only decision trace for a generated recap.
// It deliberately contains no internal profile UUIDs, listing IDs or dialog IDs.
type RecapExplanation struct {
	ProfileCode        string                   `json:"profileCode"`
	Year               uint32                   `json:"year"`
	RulesVersion       string                   `json:"rulesVersion"`
	RulesDigest        string                   `json:"rulesDigest"`
	Behavior           Behavior                 `json:"behavior"`
	BehaviorCandidates []BehaviorRuleEvaluation `json:"behaviorCandidates"`
	Achievements       []AchievementExplanation `json:"achievements"`
	NextAction         NextActionExplanation    `json:"nextAction"`
	ActionCandidates   []ActionRuleEvaluation   `json:"actionCandidates"`
}
