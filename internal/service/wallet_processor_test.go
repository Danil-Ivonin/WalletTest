package service

import (
	"context"
	"errors"
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

type blockingWalletRepo struct {
	started chan struct{}
	release chan struct{}
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

func (r *blockingWalletRepo) ApplyOperation(ctx context.Context, op domain.WalletOperation) (domain.Wallet, error) {
	r.started <- struct{}{}
	select {
	case <-r.release:
		return domain.Wallet{ID: op.WalletID, Balance: op.Amount}, nil
	case <-ctx.Done():
		return domain.Wallet{}, ctx.Err()
	}
}

func (r *blockingWalletRepo) GetWallet(ctx context.Context, id uuid.UUID) (domain.Wallet, error) {
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

func TestWalletProcessorReturnsQueueFull(t *testing.T) {
	repo := &blockingWalletRepo{
		started: make(chan struct{}, 2),
		release: make(chan struct{}, 2),
	}
	processor := NewWalletProcessor(repo, 1)
	walletID := uuid.New()

	errs := make(chan error, 2)
	go func() {
		_, err := processor.ApplyOperation(context.Background(), domain.WalletOperation{
			WalletID: walletID,
			Type:     domain.OperationTypeDeposit,
			Amount:   1,
		})
		errs <- err
	}()

	<-repo.started

	go func() {
		_, err := processor.ApplyOperation(context.Background(), domain.WalletOperation{
			WalletID: walletID,
			Type:     domain.OperationTypeDeposit,
			Amount:   1,
		})
		errs <- err
	}()

	waitUntilQueued(t, processor, walletID, 1)

	_, got := processor.ApplyOperation(context.Background(), domain.WalletOperation{
		WalletID: walletID,
		Type:     domain.OperationTypeDeposit,
		Amount:   1,
	})
	if !errors.Is(got, ErrWalletQueueFull) {
		t.Fatalf("ApplyOperation() error = %v, want %v", got, ErrWalletQueueFull)
	}

	repo.release <- struct{}{}
	<-repo.started
	repo.release <- struct{}{}

	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("queued operation error = %v", err)
		}
	}
}

func waitUntilQueued(t *testing.T, processor *WalletProcessor, walletID uuid.UUID, want int) {
	t.Helper()

	deadline := time.After(time.Second)
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for queue length %d", want)
		case <-ticker.C:
			processor.mu.Lock()
			q := processor.queues[walletID]
			got := 0
			if q != nil {
				got = len(q.jobs)
			}
			processor.mu.Unlock()
			if got == want {
				return
			}
		}
	}
}

func TestWalletProcessorAllowsDifferentWalletsInParallel(t *testing.T) {
	repo := &recordingWalletRepo{delay: 10 * time.Millisecond}
	processor := NewWalletProcessor(repo, 16)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := processor.ApplyOperation(context.Background(), domain.WalletOperation{
				WalletID: uuid.New(),
				Type:     domain.OperationTypeDeposit,
				Amount:   1,
			})
			if err != nil {
				t.Errorf("ApplyOperation() error = %v", err)
			}
		}()
	}
	wg.Wait()

	if repo.maxActive < 2 {
		t.Fatalf("maxActive=%d, want at least 2 for different wallets", repo.maxActive)
	}
}

func waitUntilActive(t *testing.T, repo *recordingWalletRepo) {
	t.Helper()

	deadline := time.After(time.Second)
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for active repository call")
		case <-ticker.C:
			repo.mu.Lock()
			active := repo.active
			repo.mu.Unlock()
			if active > 0 {
				return
			}
		}
	}
}
