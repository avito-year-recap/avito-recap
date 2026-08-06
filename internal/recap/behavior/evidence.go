package behavior

import (
	"strings"

	"github.com/year-recap/internal/recap/model"
)

func evidenceCount(metric string, actual, threshold uint64, maxPoints uint32, detail string) model.BehaviorEvidence {
	return model.BehaviorEvidence{Metric: metric, Actual: float64(actual), Threshold: float64(threshold), Points: scaledPoints(float64(actual), float64(threshold), maxPoints), Detail: detail}
}

func evidenceRate(metric string, actual, threshold float64, maxPoints uint32, detail string) model.BehaviorEvidence {
	return model.BehaviorEvidence{Metric: metric, Actual: actual, Threshold: threshold, Points: scaledPoints(actual, threshold, maxPoints), Detail: detail}
}

func evidenceInverseCount(metric string, actual, maximum uint64, maxPoints uint32, detail string) model.BehaviorEvidence {
	points := uint32(0)
	if actual <= maximum {
		points = maxPoints
	}
	return model.BehaviorEvidence{Metric: metric, Actual: float64(actual), Threshold: float64(maximum), Points: points, Detail: detail}
}

// Thresholds are eligibility boundaries, not conversion coefficients. Once a
// required condition is met it contributes its full declared weight; activity
// above the threshold does not distort comparisons between different cohorts.
func scaledPoints(actual, threshold float64, maxPoints uint32) uint32 {
	if threshold > 0 && actual >= threshold {
		return maxPoints
	}
	return 0
}

func evidenceScore(evidence []model.BehaviorEvidence) uint32 {
	var score uint32
	for _, item := range evidence {
		score += item.Points
	}
	return score
}

func evidenceReason(evidence []model.BehaviorEvidence) string {
	parts := make([]string, 0, len(evidence))
	for _, item := range evidence {
		parts = append(parts, item.Detail)
	}
	return strings.Join(parts, " ")
}
