package retry

import (
	"context"
	"errors"
	"math"
	"time"
)

//TODO: add tests

var (
	ErrDelayZero       = errors.New("delay cannot be zero")
	ErrMaxDelayZero    = errors.New("maxDelay cannot be zero")
	ErrRetriesExceeded = errors.New("retries exceeded")
)

func Backoff(
	ctx context.Context,
	maxRetries int,
	delay time.Duration,
	maxDelay time.Duration,
	work func(ctx context.Context) error,
) (err error) {
	if delay == 0 {
		return ErrDelayZero
	}
	if maxDelay == 0 {
		return ErrMaxDelayZero
	}

	err = work(ctx)
	if err == nil {
		return nil
	}

	for i := 1; i <= maxRetries; i++ {
		if i > 1 {
			delay = time.Duration(math.Pow(float64(delay), 2))
			if delay > maxDelay {
				delay = maxDelay
			}
		}

		time.Sleep(delay)

		err = work(ctx)
		if err == nil {
			return
		}
	}

	return errors.Join(ErrRetriesExceeded, err)
}
