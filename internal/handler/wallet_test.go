package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/Danil-Ivonin/WalletTest/internal/domain"
	"github.com/Danil-Ivonin/WalletTest/internal/service"
)

var errUnexpected = errors.New("unexpected")

func TestStatusFromDomainError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "invalid amount", err: domain.ErrInvalidAmount, want: http.StatusBadRequest},
		{name: "invalid operation type", err: domain.ErrInvalidOperationType, want: http.StatusBadRequest},
		{name: "wallet not found", err: domain.ErrWalletNotFound, want: http.StatusNotFound},
		{name: "insufficient funds", err: domain.ErrInsufficientFunds, want: http.StatusConflict},
		{name: "balance overflow", err: domain.ErrBalanceOverflow, want: http.StatusConflict},
		{name: "wallet queue full", err: service.ErrWalletQueueFull, want: http.StatusTooManyRequests},
		{name: "request canceled", err: context.Canceled, want: http.StatusRequestTimeout},
		{name: "request deadline exceeded", err: context.DeadlineExceeded, want: http.StatusTooManyRequests},
		{name: "unknown error", err: errUnexpected, want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusFromDomainError(tt.err); got != tt.want {
				t.Fatalf("statusFromDomainError() = %d, want %d", got, tt.want)
			}
		})
	}
}
