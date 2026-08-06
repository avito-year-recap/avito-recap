package ruleset

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/year-recap/internal/recap/model"
)

func (r Ruleset) Digest() string {
	copyRules := r
	copyRules.Version = strings.TrimSpace(copyRules.Version)
	copyRules.Algorithm = strings.TrimSpace(copyRules.Algorithm)
	copyRules.SharePolicy.Version = strings.TrimSpace(copyRules.SharePolicy.Version)
	copyRules.AchievementPolicy.Rules = append([]AchievementRuleConfig(nil), copyRules.AchievementPolicy.Rules...)
	sort.Slice(copyRules.AchievementPolicy.Rules, func(i, j int) bool {
		return copyRules.AchievementPolicy.Rules[i].Code < copyRules.AchievementPolicy.Rules[j].Code
	})
	copyRules.SharePolicy.AllowedAchievementCodes = append([]model.AchievementCode(nil), copyRules.SharePolicy.AllowedAchievementCodes...)
	sort.Slice(copyRules.SharePolicy.AllowedAchievementCodes, func(i, j int) bool {
		return copyRules.SharePolicy.AllowedAchievementCodes[i] < copyRules.SharePolicy.AllowedAchievementCodes[j]
	})
	data, err := json.Marshal(copyRules)
	if err != nil {
		panic(fmt.Sprintf("marshal validated ruleset: %v", err))
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (p SharePolicy) AchievementAllowed(code model.AchievementCode) bool {
	for _, allowed := range p.AllowedAchievementCodes {
		if code == allowed {
			return true
		}
	}
	return false
}
