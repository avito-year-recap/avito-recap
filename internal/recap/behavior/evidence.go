package behavior

import (
	"strings"

	"github.com/year-recap/internal/recap/model"
)

func evidenceReason(evidence []model.BehaviorEvidence) string {
	parts := make([]string, 0, len(evidence))
	for _, item := range evidence {
		parts = append(parts, item.Detail)
	}
	return strings.Join(parts, " ")
}
