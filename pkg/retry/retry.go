package retry

import (
	"fmt"
	"time"
)

func Do(attempts int, duration time.Duration, fn func() error) error {
	var err error
	for range attempts {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(duration)
	}

	if err != nil {
		return fmt.Errorf("%d reties failed: %w", attempts, err)
	}
	return nil
}
