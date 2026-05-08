package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Danil-Ivonin/WalletTest/internal/domain"
	"github.com/google/uuid"
)

type recordingWalletRepo struct {
	mu        sync.Mutex
	active    int
	maxActive int
	finished  int
	delay     time.Duration
}

func (r *recordingWalletRepo) ApplyOperation(ctx context.Context, op domain.WalletOperation) (domain.Wallet, error) {
	r.mu.Lock()
	r.active++
	if r.active > r.maxActive {
		r.maxActive = r.active
	}
	r.mu.Unlock()

	select {
	case <-time.After(r.delay):
	case <-ctx.Done():
		return domain.Wallet{}, ctx.Err()
	}

	r.mu.Lock()
	r.active--
	r.finished++
	balance := int64(r.finished)
	r.mu.Unlock()

	return domain.Wallet{ID: op.WalletID, Balance: balance}, nil
}

func (r *recordingWalletRepo) GetWallet(ctx context.Context, id uuid.UUID) (domain.Wallet, error) {
	return domain.Wallet{ID: id}, nil
}

func TestWalletProcessorSerializesSameWallet(t *testing.T) {
	repo := &recordingWalletRepo{delay: time.Millisecond}
	processor := NewWalletProcessor(repo, 128)
	walletID := uuid.New()

	const jobs = 50
	var wg sync.WaitGroup
	errs := make(chan error, jobs)

	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := processor.ApplyOperation(context.Background(), domain.WalletOperation{
				WalletID: walletID,
				Type:     domain.OperationTypeDeposit,
				Amount:   1,
			})
			errs <- err
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if repo.maxActive != 1 {
		t.Fatalf("maxActive=%d, want 1", repo.maxActive)
	}
	if repo.finished != jobs {
		t.Fatalf("finished=%d, want %d", repo.finished, jobs)
	}
}
