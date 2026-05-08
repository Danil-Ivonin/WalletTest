package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Danil-Ivonin/WalletTest/internal/domain"
	"github.com/google/uuid"
)

type walletServiceRepo struct {
	applied []domain.WalletOperation
	wallet  domain.Wallet
	err     error
}

func (r *walletServiceRepo) ApplyOperation(ctx context.Context, op domain.WalletOperation) (domain.Wallet, error) {
	r.applied = append(r.applied, op)
	if r.err != nil {
		return domain.Wallet{}, r.err
	}
	return domain.Wallet{ID: op.WalletID, Balance: op.Amount}, nil
}

func (r *walletServiceRepo) GetWallet(ctx context.Context, id uuid.UUID) (domain.Wallet, error) {
	if r.err != nil {
		return domain.Wallet{}, r.err
	}
	if r.wallet.ID == uuid.Nil {
		r.wallet.ID = id
	}
	return r.wallet, nil
}

func TestWalletServiceApplyOperationValidatesInput(t *testing.T) {
	walletID := uuid.New()
	tests := []struct {
		name   string
		opType domain.OperationType
		amount int64
		want   error
	}{
		{name: "zero amount", opType: domain.OperationTypeDeposit, amount: 0, want: domain.ErrInvalidAmount},
		{name: "negative amount", opType: domain.OperationTypeDeposit, amount: -1, want: domain.ErrInvalidAmount},
		{name: "invalid operation", opType: domain.OperationType("TRANSFER"), amount: 1, want: domain.ErrInvalidOperationType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &walletServiceRepo{}
			svc := NewWalletService(repo, 4)

			_, got := svc.ApplyOperation(context.Background(), walletID, tt.opType, tt.amount)
			if !errors.Is(got, tt.want) {
				t.Fatalf("ApplyOperation() error = %v, want %v", got, tt.want)
			}
			if len(repo.applied) != 0 {
				t.Fatalf("repository was called for invalid input: %#v", repo.applied)
			}
		})
	}
}

func TestWalletServiceApplyOperationDelegatesValidInput(t *testing.T) {
	repo := &walletServiceRepo{}
	svc := NewWalletService(repo, 4)
	walletID := uuid.New()

	wallet, err := svc.ApplyOperation(context.Background(), walletID, domain.OperationTypeDeposit, 100)
	if err != nil {
		t.Fatalf("ApplyOperation() error = %v", err)
	}
	if wallet.ID != walletID || wallet.Balance != 100 {
		t.Fatalf("wallet = %#v, want id %s balance 100", wallet, walletID)
	}
	if len(repo.applied) != 1 {
		t.Fatalf("repository calls = %d, want 1", len(repo.applied))
	}
	if repo.applied[0].WalletID != walletID || repo.applied[0].Type != domain.OperationTypeDeposit || repo.applied[0].Amount != 100 {
		t.Fatalf("operation = %#v", repo.applied[0])
	}
}

func TestWalletServiceGetBalanceDelegatesToRepository(t *testing.T) {
	walletID := uuid.New()
	repo := &walletServiceRepo{wallet: domain.Wallet{ID: walletID, Balance: 55}}
	svc := NewWalletService(repo, 4)

	wallet, err := svc.GetBalance(context.Background(), walletID)
	if err != nil {
		t.Fatalf("GetBalance() error = %v", err)
	}
	if wallet.ID != walletID || wallet.Balance != 55 {
		t.Fatalf("wallet = %#v, want id %s balance 55", wallet, walletID)
	}
}
