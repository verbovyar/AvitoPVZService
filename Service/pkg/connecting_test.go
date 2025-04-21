package postgres

import (
	"errors"
	"testing"
	"time"
)

func TestDoWithTries_SucceedsAfterRetries(t *testing.T) {
	attempts := 0
	err := doWithTries(func() error {
		attempts++
		if attempts < 3 {
			return errors.New("fail")
		}
		return nil
	}, 5, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestDoWithTries_FailsWhenExhausted(t *testing.T) {
	count := 0
	err := doWithTries(func() error {
		count++
		return errors.New("always fail")
	}, 2, 1*time.Millisecond)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if count != 2 {
		t.Errorf("expected 2 attempts, got %d", count)
	}
}
