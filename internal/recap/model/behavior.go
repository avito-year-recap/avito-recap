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

type BehaviorEvidence struct {
	Metric    string  `json:"metric"`
	Actual    float64 `json:"actual"`
	Threshold float64 `json:"threshold"`
	Points    uint32  `json:"points"`
	Detail    string  `json:"detail"`
}

type Behavior struct {
	Code        BehaviorCode       `json:"code"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Reason      string             `json:"reason"`
	Score       uint32             `json:"score"`
	Evidence    []BehaviorEvidence `json:"evidence,omitempty"`
}
