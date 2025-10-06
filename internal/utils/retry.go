package utils

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryConfig defines configuration for retry operations
type RetryConfig struct {
	MaxAttempts   int           // Maximum number of retry attempts
	InitialDelay  time.Duration // Initial delay before first retry
	MaxDelay      time.Duration // Maximum delay between retries
	BackoffFactor float64       // Multiplier for exponential backoff (e.g., 2.0 for doubling)
	Jitter        bool          // Add randomness to prevent thundering herd
}

// DefaultRetryConfig returns a sensible default configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
	}
}

// RetryableError indicates an error that can be retried
type RetryableError interface {
	error
	IsRetryable() bool
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check if error implements RetryableError interface
	if retryErr, ok := err.(RetryableError); ok {
		return retryErr.IsRetryable()
	}

	// Default: treat timeout and temporary errors as retryable
	errMsg := err.Error()
	return containsAny(errMsg, []string{
		"timeout",
		"temporary",
		"connection refused",
		"connection reset",
		"broken pipe",
		"service unavailable",
		"too many requests",
	})
}

// RetryWithBackoff executes an operation with exponential backoff retry logic
func RetryWithBackoff(ctx context.Context, config RetryConfig, operation func() error) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Execute operation
		err := operation()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryableError(err) {
			return fmt.Errorf("non-retryable error on attempt %d: %w", attempt+1, err)
		}

		// Check if we've exhausted retries
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Calculate delay with exponential backoff
		currentDelay := delay
		if config.Jitter {
			// Add jitter (±10% randomness)
			jitter := time.Duration(rand.Int63n(int64(float64(delay) * 0.2)))
			currentDelay = delay + jitter - time.Duration(float64(delay)*0.1)
		}

		// Check context before sleeping
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled by context: %w", ctx.Err())
		case <-time.After(currentDelay):
			// Continue to next attempt
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * config.BackoffFactor)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", config.MaxAttempts, lastErr)
}

// RetryWithBackoffFunc is a convenience function that uses default config
func RetryWithBackoffFunc(ctx context.Context, operation func() error) error {
	return RetryWithBackoff(ctx, DefaultRetryConfig(), operation)
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitClosed CircuitBreakerState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	failureCount    int
	lastFailureTime time.Time
	state           CircuitBreakerState
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitClosed,
	}
}

// Execute runs an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(operation func() error) error {
	// Check if circuit should be reset
	if cb.state == CircuitOpen && time.Since(cb.lastFailureTime) > cb.resetTimeout {
		cb.state = CircuitHalfOpen
		cb.failureCount = 0
	}

	// Reject if circuit is open
	if cb.state == CircuitOpen {
		return fmt.Errorf("circuit breaker is open, rejecting operation")
	}

	// Execute operation
	err := operation()
	if err != nil {
		cb.recordFailure()
		return err
	}

	// Success - reset if we were in half-open state
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
		cb.failureCount = 0
	}

	return nil
}

// recordFailure increments failure count and opens circuit if threshold exceeded
func (cb *CircuitBreaker) recordFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.failureCount >= cb.maxFailures {
		cb.state = CircuitOpen
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	return cb.state
}

// Helper functions

func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	// Simple case-insensitive contains check
	sLower := ""
	substrLower := ""
	for _, r := range s {
		sLower += string(toLower(r))
	}
	for _, r := range substr {
		substrLower += string(toLower(r))
	}

	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}

// CalculateExponentialBackoff calculates delay for a given attempt
func CalculateExponentialBackoff(attempt int, baseDelay time.Duration, maxDelay time.Duration) time.Duration {
	delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt)))
	if delay > maxDelay {
		return maxDelay
	}
	return delay
}
