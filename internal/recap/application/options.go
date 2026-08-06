package application

import (
	"github.com/year-recap/internal/recap/ruleset"
	"time"
)

func WithClock(clock func() time.Time) Option {
	return func(service *Service) {
		if clock != nil {
			service.now = clock
		}
	}
}

func WithIDGenerator(generator IDGenerator) Option {
	return func(service *Service) {
		if generator != nil {
			service.newID = generator
		}
	}
}

func WithRuleset(configured ruleset.Ruleset) Option {
	return func(service *Service) { service.ruleset = configured }
}
