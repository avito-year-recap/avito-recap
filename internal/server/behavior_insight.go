package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/year-recap/internal/recap/application"
	transportconnect "github.com/year-recap/internal/transport/connect"
)

const maxBehaviorInsightBody = 1 << 10

type behaviorInsightRequest struct {
	ProfileCode string `json:"profileCode"`
	StartAt     string `json:"startAt"`
	EndAt       string `json:"endAt"`
}

func behaviorInsight(applicationAPI transportconnect.Application) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			response.Header().Set("Allow", "POST")
			writeExplainError(response, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
			return
		}

		var body behaviorInsightRequest
		decoder := json.NewDecoder(io.LimitReader(request.Body, maxBehaviorInsightBody))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			writeExplainError(response, http.StatusBadRequest, "invalid JSON body: "+err.Error())
			return
		}

		profileCode := strings.TrimSpace(body.ProfileCode)
		if profileCode == "" {
			writeExplainError(response, http.StatusBadRequest, "profileCode is required")
			return
		}
		startAt, err := time.Parse(time.RFC3339, strings.TrimSpace(body.StartAt))
		if err != nil {
			writeExplainError(response, http.StatusBadRequest, "startAt must be an RFC3339 timestamp")
			return
		}
		endAt, err := time.Parse(time.RFC3339, strings.TrimSpace(body.EndAt))
		if err != nil {
			writeExplainError(response, http.StatusBadRequest, "endAt must be an RFC3339 timestamp")
			return
		}

		profile, err := applicationAPI.GetProfileByCode(request.Context(), profileCode)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, application.ErrProfileNotFound) {
				status = http.StatusNotFound
			}
			writeExplainError(response, status, err.Error())
			return
		}

		result, err := applicationAPI.AnalyzeBehavior(request.Context(), profile.ID, startAt, endAt)
		if err != nil {
			status := http.StatusInternalServerError
			switch {
			case errors.Is(err, application.ErrInvalidPeriod):
				status = http.StatusBadRequest
			case errors.Is(err, application.ErrPeriodTooLong):
				status = http.StatusRequestEntityTooLarge
			case errors.Is(err, application.ErrNoActivityInPeriod):
				status = http.StatusUnprocessableEntity
			case errors.Is(err, application.ErrProfileNotFound):
				status = http.StatusNotFound
			case errors.Is(err, application.ErrEventRangeUnsupported), errors.Is(err, application.ErrInsightUnavailable):
				status = http.StatusServiceUnavailable
			}
			writeExplainError(response, status, err.Error())
			return
		}

		response.Header().Set("Content-Type", "application/json; charset=utf-8")
		response.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(response).Encode(result)
	}
}
