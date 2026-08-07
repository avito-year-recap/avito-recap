package behavior

import (
	"strings"

	"github.com/year-recap/internal/recap/model"
)

func evidenceCount(metric string, actual, threshold uint64, detail string) model.BehaviorEvidence {
	return model.BehaviorEvidence{
		Metric: metric, Actual: float64(actual), Threshold: float64(threshold), Detail: detail,
	}
}

func evidenceRate(metric string, actual, threshold float64, detail string) model.BehaviorEvidence {
	return model.BehaviorEvidence{Metric: metric, Actual: actual, Threshold: threshold, Detail: detail}
}

func evidenceInverseCount(metric string, actual, maximum uint64, detail string) model.BehaviorEvidence {
	return model.BehaviorEvidence{
		Metric: metric, Actual: float64(actual), Threshold: float64(maximum), Detail: detail,
	}
}

func evidenceReason(evidence []model.BehaviorEvidence) string {
	parts := make([]string, 0, len(evidence))
	for _, item := range evidence {
		parts = append(parts, item.Detail)
	}
	return strings.Join(parts, " ")
}
