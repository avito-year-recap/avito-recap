package analytics

import (
	"errors"
	"fmt"
	"time"

	"github.com/year-recap/internal/recap/model"
)

var (
	ErrInvalidYear     = errors.New("invalid recap year")
	ErrYearNotComplete = errors.New("recap year is not complete")
)

func CompletedYearPeriod(year uint32, now time.Time) (model.RecapPeriod, error) {
	now = now.UTC()
	if year == 0 || year > uint32(now.Year()) {
		return model.RecapPeriod{}, ErrInvalidYear
	}
	if year == uint32(now.Year()) {
		return model.RecapPeriod{}, fmt.Errorf("%w: %d is still in progress", ErrYearNotComplete, year)
	}
	start := time.Date(int(year), time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(int(year)+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	return model.RecapPeriod{Year: year, StartAt: start, EndAt: end, Final: true}, nil
}
