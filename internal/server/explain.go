package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/year-recap/internal/recap/application"
	transportconnect "github.com/year-recap/internal/transport/connect"
)

func explainRecap(applicationAPI transportconnect.Application) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			response.Header().Set("Allow", "GET, HEAD")
			http.Error(response, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		profileCode := strings.TrimSpace(request.URL.Query().Get("profile_code"))
		if profileCode == "" {
			writeExplainError(response, http.StatusBadRequest, "profile_code is required")
			return
		}
		yearRaw := strings.TrimSpace(request.URL.Query().Get("year"))
		parsedYear, err := strconv.ParseUint(yearRaw, 10, 32)
		if err != nil || parsedYear == 0 {
			writeExplainError(response, http.StatusBadRequest, "year must be a positive integer")
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
		explanation, err := applicationAPI.ExplainRecap(request.Context(), profile.ID, uint32(parsedYear))
		if err != nil {
			status := http.StatusInternalServerError
			switch {
			case errors.Is(err, application.ErrInvalidYear):
				status = http.StatusBadRequest
			case errors.Is(err, application.ErrYearNotComplete), errors.Is(err, application.ErrNotEnoughActivity):
				status = http.StatusUnprocessableEntity
			case errors.Is(err, application.ErrProfileNotFound):
				status = http.StatusNotFound
			}
			writeExplainError(response, status, err.Error())
			return
		}

		response.Header().Set("Content-Type", "application/json; charset=utf-8")
		response.WriteHeader(http.StatusOK)
		if request.Method == http.MethodHead {
			return
		}
		_ = json.NewEncoder(response).Encode(explanation)
	}
}

func writeExplainError(response http.ResponseWriter, status int, message string) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(map[string]string{"error": message})
}
