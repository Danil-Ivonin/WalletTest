package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestWithRetryRetriesRetryablePostgresErrors(t *testing.T) {
	attempts := 0
	err := withRetry(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			return &pgconn.PgError{Code: "40001"}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("withRetry() error = %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
}

func TestWithRetryStopsOnNonRetryableError(t *testing.T) {
	want := errors.New("not retryable")
	attempts := 0
	err := withRetry(context.Background(), func() error {
		attempts++
		return want
	})

	if !errors.Is(err, want) {
		t.Fatalf("withRetry() error = %v, want %v", err, want)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestWithRetryStopsWhenContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	attempts := 0
	err := withRetry(ctx, func() error {
		attempts++
		return nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("withRetry() error = %v, want context.Canceled", err)
	}
	if attempts != 0 {
		t.Fatalf("attempts = %d, want 0", attempts)
	}
}

func TestIsRetryablePostgresError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "serialization failure", err: &pgconn.PgError{Code: "40001"}, want: true},
		{name: "deadlock detected", err: &pgconn.PgError{Code: "40P01"}, want: true},
		{name: "unique violation", err: &pgconn.PgError{Code: "23505"}, want: false},
		{name: "plain error", err: errors.New("plain"), want: false},
		{name: "nil", err: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryablePostgresError(tt.err); got != tt.want {
				t.Fatalf("isRetryablePostgresError() = %v, want %v", got, tt.want)
			}
		})
	}
}
