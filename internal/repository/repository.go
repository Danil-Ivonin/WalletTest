package repository

import (
	"context"

	"github.com/Danil-Ivonin/WalletTest/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Wallet interface {
	ApplyOperation(ctx context.Context, op domain.WalletOperation) (domain.Wallet, error)
	GetWallet(ctx context.Context, id uuid.UUID) (domain.Wallet, error)
}

type Repository struct {
	Wallet
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		Wallet: NewWalletRepository(pool),
	}
}
