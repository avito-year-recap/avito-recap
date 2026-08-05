package recap

import "errors"

func sumUint64(values ...uint64) (uint64, error) {
	var result uint64

	for _, value := range values {
		if value > ^uint64(0)-result {
			return 0, errors.New("event counters overflow uint64")
		}

		result += value
	}

	return result, nil
}
