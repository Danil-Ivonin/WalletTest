package service

import (
	"context"

	"github.com/Danil-Ivonin/WalletTest/internal/domain"
	"github.com/Danil-Ivonin/WalletTest/internal/repository"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type Wallet interface {
	ApplyOperation(ctx context.Context, walletID uuid.UUID, opType domain.OperationType, amount int64) (domain.Wallet, error)
	GetBalance(ctx context.Context, walletID uuid.UUID) (domain.Wallet, error)
}

type Service struct {
	Wallet
}

func NewService(repo *repository.Repository) *Service {
	return &Service{
		Wallet: NewWalletService(repo.Wallet, viper.GetInt("app.wallet_queue_size")),
	}
}
