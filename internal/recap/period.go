package recap

import (
	"fmt"
	"time"
)

func completedYearPeriod(year uint32, now time.Time) (RecapPeriod, error) {
	now = now.UTC()
	if year == 0 || year > uint32(now.Year()) {
		return RecapPeriod{}, ErrInvalidYear
	}
	if year == uint32(now.Year()) {
		return RecapPeriod{}, fmt.Errorf("%w: %d is still in progress", ErrYearNotComplete, year)
	}

	start := time.Date(int(year), time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(int(year)+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	return RecapPeriod{Year: year, StartAt: start, EndAt: end, Final: true}, nil
}
