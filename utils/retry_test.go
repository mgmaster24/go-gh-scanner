package utils

import (
	"errors"
	"testing"
	"time"
)

func TestRetry_SucceedsFirstAttempt(t *testing.T) {
	calls := 0
	err := Retry(3, time.Millisecond, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetry_SucceedsAfterRetries(t *testing.T) {
	calls := 0
	err := Retry(3, time.Millisecond, func() error {
		calls++
		if calls < 3 {
			return errors.New("transient")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil after retries, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetry_ExhaustsAttempts(t *testing.T) {
	sentinel := errors.New("permanent failure")
	calls := 0
	err := Retry(3, time.Millisecond, func() error {
		calls++
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected exactly 3 calls, got %d", calls)
	}
}
