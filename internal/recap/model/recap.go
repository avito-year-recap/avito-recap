package model

import (
	"github.com/google/uuid"
	"time"
)

type RecapKey struct {
	ProfileID    uuid.UUID `json:"profileId"`
	Year         uint32    `json:"year"`
	RulesVersion string    `json:"rulesVersion"`
	RulesDigest  string    `json:"rulesDigest"`
}

type Recap struct {
	ID              uuid.UUID       `json:"id"`
	ShareID         uuid.UUID       `json:"shareId"`
	Profile         Profile         `json:"profile"`
	Year            uint32          `json:"year"`
	Period          RecapPeriod     `json:"period"`
	RulesVersion    string          `json:"rulesVersion"`
	RulesDigest     string          `json:"rulesDigest"`
	Metrics         Metrics         `json:"metrics"`
	ActionableState ActionableState `json:"actionableState"`
	Behavior        Behavior        `json:"behavior"`
	Achievements    []Achievement   `json:"achievements"`
	Cards           []Card          `json:"cards"`
	NextAction      NextAction      `json:"nextAction"`
	GeneratedAt     time.Time       `json:"generatedAt"`
}

func (r Recap) Key() RecapKey {
	return RecapKey{ProfileID: r.Profile.ID, Year: r.Year, RulesVersion: r.RulesVersion, RulesDigest: r.RulesDigest}
}

type ShareCard struct {
	ShareID          uuid.UUID `json:"shareId"`
	Year             uint32    `json:"year"`
	PrivacyVersion   string    `json:"privacyVersion"`
	BehaviorTitle    string    `json:"behaviorTitle"`
	AchievementTitle string    `json:"achievementTitle,omitempty"`
	TopCategory      string    `json:"topCategory,omitempty"`
}

// ShareCard is both the strict public DTO and the payload of the final story
// card. Keeping one type guarantees that the user sees exactly the same safe
// data that will be available through the public share link.
func (ShareCard) isCardPayload() {}
