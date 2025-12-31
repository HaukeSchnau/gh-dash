package data

import (
	"context"
	"errors"
	"math"
	"net"
	"strings"
	"time"
)

type retryableError struct {
	err error
}

func (e retryableError) Error() string {
	return e.err.Error()
}

func markRetryable(err error) error {
	if err == nil {
		return nil
	}
	return retryableError{err: err}
}

func retryRead[T any](fn func() (T, error)) (T, error) {
	const maxAttempts = 3
	const baseDelay = 200 * time.Millisecond
	const maxDelay = 2 * time.Second
	var zero T

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		val, err := fn()
		if err == nil {
			return val, nil
		}
		lastErr = err
		if !shouldRetry(err) || attempt == maxAttempts {
			break
		}
		sleepWithBackoff(attempt, baseDelay, maxDelay)
	}
	return zero, lastErr
}

func shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	var retryable retryableError
	if errors.As(err, &retryable) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") || strings.Contains(msg, "temporary") || strings.Contains(msg, "connection reset")
}

func sleepWithBackoff(attempt int, baseDelay time.Duration, maxDelay time.Duration) {
	backoff := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))
	if backoff > maxDelay {
		backoff = maxDelay
	}
	time.Sleep(backoff)
}
