package service

import (
	"context"

	"github.com/Danil-Ivonin/WalletTest/internal/domain"
	"github.com/Danil-Ivonin/WalletTest/internal/repository"
	"github.com/google/uuid"
)

type WalletService struct {
	repo repository.Wallet
}

func NewWalletService(repo repository.Wallet) *WalletService {
	return &WalletService{repo: repo}
}

func (w WalletService) ApplyOperation(ctx context.Context, walletID uuid.UUID, opType domain.OperationType, amount int64) (domain.Wallet, error) {
	//TODO implement me
	panic("implement me")
}

func (w WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (domain.Wallet, error) {
	//TODO implement me
	panic("implement me")
}
