package bootstrap

import (
	"fmt"
	"strings"

	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/validation/structural"
	"github.com/year-recap/internal/seed"
)

// DemoData is the fully expanded demo dataset used by one-off tools such as
// cmd/eventgen. Profiles come from seeds/profiles.json; Events are generated
// from seeds/scenarios.json with the same canonical seed generator used by the
// rest of the project.
type DemoData struct {
	Profiles []model.Profile
	Events   []model.Event
}

// GenerateDemoData loads demo profiles and scenarios and expands every
// scenario into raw events. It performs no storage writes, which makes it
// suitable for producers such as cmd/eventgen that publish the generated
// events to Kafka instead of inserting them directly into ClickHouse.
func GenerateDemoData(profilesPath, scenariosPath string) (DemoData, error) {
	profiles, err := readJSON[[]model.Profile](profilesPath)
	if err != nil {
		return DemoData{}, fmt.Errorf("load profiles seed: %w", err)
	}

	scenarios, err := readJSON[[]seed.Scenario](scenariosPath)
	if err != nil {
		return DemoData{}, fmt.Errorf("load scenarios seed: %w", err)
	}

	byCode := make(map[string]model.Profile, len(profiles))
	for index := range profiles {
		profiles[index] = model.NormalizeProfile(profiles[index])
		profile := profiles[index]
		if err := structural.ValidateProfile(profile); err != nil {
			return DemoData{}, fmt.Errorf("invalid profile seed at index %d: %w", index, err)
		}
		if _, exists := byCode[profile.Code]; exists {
			return DemoData{}, fmt.Errorf("duplicate profile code %q", profile.Code)
		}
		byCode[profile.Code] = profile
	}

	data := DemoData{
		Profiles: profiles,
		Events:   make([]model.Event, 0),
	}

	for _, scenario := range scenarios {
		code := strings.TrimSpace(scenario.ProfileCode)
		profile, ok := byCode[code]
		if !ok {
			return DemoData{}, fmt.Errorf("scenario references unknown profile code %q", scenario.ProfileCode)
		}

		events, err := seed.EventsFromScenario(profile.ID, scenario)
		if err != nil {
			return DemoData{}, fmt.Errorf("build events for %s/%d: %w", profile.Code, scenario.Year, err)
		}
		data.Events = append(data.Events, events...)
	}

	return data, nil
}
