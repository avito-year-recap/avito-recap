package model

type BehaviorCode string

const (
	BehaviorActiveSeller   BehaviorCode = "ACTIVE_SELLER"
	BehaviorStartingSeller BehaviorCode = "STARTING_SELLER"
	BehaviorDecisiveBuyer  BehaviorCode = "DECISIVE_BUYER"
	BehaviorFindHunter     BehaviorCode = "FIND_HUNTER"
	BehaviorResearcher     BehaviorCode = "RESEARCHER"
	BehaviorUniversal      BehaviorCode = "UNIVERSAL_USER"
)

// BehaviorEvidence explains why a behavior rule matched. Thresholds are
// eligibility boundaries; the model intentionally does not expose a pseudo
// confidence score because this MVP is rule-based, not probabilistic.
type BehaviorEvidence struct {
	Metric    string  `json:"metric"`
	Actual    float64 `json:"actual"`
	Threshold float64 `json:"threshold"`
	Detail    string  `json:"detail"`
}

type Behavior struct {
	Code        BehaviorCode       `json:"code"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Reason      string             `json:"reason"`
	Evidence    []BehaviorEvidence `json:"evidence,omitempty"`
}
