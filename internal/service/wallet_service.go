package service

import (
	"context"

	"github.com/Danil-Ivonin/WalletTest/internal/domain"
	"github.com/Danil-Ivonin/WalletTest/internal/repository"
	"github.com/google/uuid"
)

type WalletService struct {
	repo      repository.Wallet
	processor *WalletProcessor
}

func NewWalletService(repo repository.Wallet, queueSize int) *WalletService {
	return &WalletService{
		repo:      repo,
		processor: NewWalletProcessor(repo, queueSize),
	}
}

func (s *WalletService) ApplyOperation(ctx context.Context, walletID uuid.UUID, opType domain.OperationType, amount int64) (domain.Wallet, error) {
	if amount <= 0 {
		return domain.Wallet{}, domain.ErrInvalidAmount
	}
	if err := domain.ValidateOperationType(opType); err != nil {
		return domain.Wallet{}, err
	}

	return s.processor.ApplyOperation(ctx, domain.WalletOperation{
		WalletID: walletID,
		Type:     opType,
		Amount:   amount,
	})
}

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (domain.Wallet, error) {
	return s.repo.GetWallet(ctx, walletID)
}
