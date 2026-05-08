package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Danil-Ivonin/WalletTest/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepository struct {
	pool *pgxpool.Pool
}

func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{pool: pool}
}

func (r *WalletRepository) ApplyOperation(ctx context.Context, op domain.WalletOperation) (domain.Wallet, error) {
	var wallet domain.Wallet
	err := withRetry(ctx, func() error {
		got, err := r.applyOperationOnce(ctx, op)
		if err != nil {
			return err
		}
		wallet = got
		return nil
	})
	if err != nil {
		return domain.Wallet{}, err
	}
	return wallet, nil
}

func (r *WalletRepository) GetWallet(ctx context.Context, id uuid.UUID) (domain.Wallet, error) {
	var wallet domain.Wallet
	err := r.pool.QueryRow(ctx, `
SELECT id, balance, created_at, updated_at
FROM wallets
WHERE id = $1
`, id).Scan(&wallet.ID, &wallet.Balance, &wallet.CreatedAt, &wallet.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Wallet{}, domain.ErrWalletNotFound
	}
	if err != nil {
		return domain.Wallet{}, fmt.Errorf("get wallet %s: %w", id, err)
	}
	return wallet, nil
}

func (r *WalletRepository) applyOperationOnce(ctx context.Context, op domain.WalletOperation) (domain.Wallet, error) {
	//Open transaction
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Wallet{}, fmt.Errorf("begin wallet transaction: %w", err)
	}
	//Rollback on error
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	//Switch on operation type
	var wallet domain.Wallet
	switch op.Type {
	case domain.OperationTypeDeposit:
		wallet, err = applyDeposit(ctx, tx, op)
	case domain.OperationTypeWithdraw:
		wallet, err = applyWithdraw(ctx, tx, op)
	default:
		return domain.Wallet{}, domain.ErrInvalidOperationType
	}
	if err != nil {
		return domain.Wallet{}, err
	}

	//transactions audit log
	if err := insertTransaction(ctx, tx, op, wallet.Balance); err != nil {
		return domain.Wallet{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Wallet{}, fmt.Errorf("commit wallet transaction: %w", err)
	}
	return wallet, nil
}

func applyDeposit(ctx context.Context, tx pgx.Tx, op domain.WalletOperation) (domain.Wallet, error) {
	//Create new wallet if not exist
	if _, err := tx.Exec(ctx, `
INSERT INTO wallets (id, balance)
VALUES ($1, 0)
ON CONFLICT (id) DO NOTHING
`, op.WalletID); err != nil {
		return domain.Wallet{}, fmt.Errorf("ensure wallet %s exists: %w", op.WalletID, err)
	}

	//Apply deposit
	var wallet domain.Wallet
	err := tx.QueryRow(ctx, `
UPDATE wallets
SET balance = balance + $2,
    updated_at = now()
WHERE id = $1
  AND balance <= 9223372036854775807 - $2
RETURNING id, balance, created_at, updated_at
`, op.WalletID, op.Amount).Scan(&wallet.ID, &wallet.Balance, &wallet.CreatedAt, &wallet.UpdatedAt)
	if err == nil {
		return wallet, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Wallet{}, domain.ErrBalanceOverflow
	}
	return domain.Wallet{}, fmt.Errorf("deposit wallet %s: %w", op.WalletID, err)
}

func applyWithdraw(ctx context.Context, tx pgx.Tx, op domain.WalletOperation) (domain.Wallet, error) {
	var wallet domain.Wallet
	err := tx.QueryRow(ctx, `
UPDATE wallets
SET balance = balance - $2,
    updated_at = now()
WHERE id = $1
  AND balance >= $2
RETURNING id, balance, created_at, updated_at
`, op.WalletID, op.Amount).Scan(&wallet.ID, &wallet.Balance, &wallet.CreatedAt, &wallet.UpdatedAt)
	if err == nil {
		return wallet, nil
	}
	//if error is not ErrNoRows
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Wallet{}, fmt.Errorf("withdraw wallet %s: %w", op.WalletID, err)
	}

	//Wallet not exist, or not enough money
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM wallets WHERE id = $1)`, op.WalletID).Scan(&exists); err != nil {
		return domain.Wallet{}, fmt.Errorf("check wallet %s exists: %w", op.WalletID, err)
	}
	if !exists {
		return domain.Wallet{}, domain.ErrWalletNotFound
	}
	return domain.Wallet{}, domain.ErrInsufficientFunds
}

func insertTransaction(ctx context.Context, tx pgx.Tx, op domain.WalletOperation, balanceAfter int64) error {
	_, err := tx.Exec(ctx, `
INSERT INTO wallet_transactions (id, wallet_id, operation_type, amount, balance_after)
VALUES ($1, $2, $3, $4, $5)
`, uuid.New(), op.WalletID, string(op.Type), op.Amount, balanceAfter)
	if err != nil {
		return fmt.Errorf("insert wallet transaction for %s: %w", op.WalletID, err)
	}
	return nil
}

func withRetry(ctx context.Context, fn func() error) error {
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = fn()
		if !isRetryablePostgresError(err) {
			return err
		}
	}
	return err
}

func isRetryablePostgresError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "40001" || pgErr.Code == "40P01"
}
