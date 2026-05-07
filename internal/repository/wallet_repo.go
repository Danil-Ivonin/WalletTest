package repository

import (
	"context"

	"github.com/Danil-Ivonin/WalletTest/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepository struct {
	pool *pgxpool.Pool
}

func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{pool: pool}
}

func (w WalletRepository) ApplyOperation(ctx context.Context, op domain.WalletOperation) (domain.Wallet, error) {
	//TODO implement me
	panic("implement me")
}

func (w WalletRepository) GetWallet(ctx context.Context, id uuid.UUID) (domain.Wallet, error) {
	//TODO implement me
	panic("implement me")
}
