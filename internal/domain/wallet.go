package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type OperationType string

const (
	OperationTypeDeposit  OperationType = "DEPOSIT"
	OperationTypeWithdraw OperationType = "WITHDRAW"
)

var (
	ErrInvalidAmount        = errors.New("amount must be greater than zero")
	ErrInvalidOperationType = errors.New("operation type must be DEPOSIT or WITHDRAW")
	ErrWalletNotFound       = errors.New("wallet not found")
	ErrInsufficientFunds    = errors.New("insufficient funds")
)

type Wallet struct {
	ID        uuid.UUID
	Balance   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WalletOperation struct {
	WalletID uuid.UUID
	Type     OperationType
	Amount   int64
}

func ValidateOperationType(op OperationType) error {
	switch op {
	case OperationTypeDeposit, OperationTypeWithdraw:
		return nil
	default:
		return ErrInvalidOperationType
	}
}
