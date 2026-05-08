package service

import (
	"context"
	"errors"
	"sync"

	"github.com/Danil-Ivonin/WalletTest/internal/domain"
	"github.com/Danil-Ivonin/WalletTest/internal/repository"
	"github.com/google/uuid"
)

var ErrWalletQueueFull = errors.New("wallet operation queue is full")

type walletJob struct {
	ctx    context.Context
	op     domain.WalletOperation
	result chan walletJobResult
}

type walletJobResult struct {
	wallet domain.Wallet
	err    error
}

type walletQueue struct {
	jobs chan walletJob
}

type WalletProcessor struct {
	repo      repository.Wallet
	queueSize int
	mu        sync.Mutex
	queues    map[uuid.UUID]*walletQueue
}

func NewWalletProcessor(repo repository.Wallet, queueSize int) *WalletProcessor {
	if queueSize <= 0 {
		queueSize = 2048
	}
	return &WalletProcessor{
		repo:      repo,
		queueSize: queueSize,
		queues:    make(map[uuid.UUID]*walletQueue),
	}
}

func (p *WalletProcessor) ApplyOperation(ctx context.Context, op domain.WalletOperation) (domain.Wallet, error) {
	q := p.queueFor(op.WalletID)
	job := walletJob{
		ctx:    ctx,
		op:     op,
		result: make(chan walletJobResult, 1),
	}

	select {
	case q.jobs <- job:
	case <-ctx.Done():
		return domain.Wallet{}, ctx.Err()
	default:
		return domain.Wallet{}, ErrWalletQueueFull
	}

	select {
	case result := <-job.result:
		return result.wallet, result.err
	case <-ctx.Done():
		return domain.Wallet{}, ctx.Err()
	}
}

func (p *WalletProcessor) queueFor(walletID uuid.UUID) *walletQueue {
	p.mu.Lock()
	defer p.mu.Unlock()

	if q, ok := p.queues[walletID]; ok {
		return q
	}

	q := &walletQueue{jobs: make(chan walletJob, p.queueSize)}
	p.queues[walletID] = q
	go p.runWalletQueue(q)
	return q
}

func (p *WalletProcessor) runWalletQueue(q *walletQueue) {
	for job := range q.jobs {
		if err := job.ctx.Err(); err != nil {
			job.result <- walletJobResult{err: err}
			continue
		}
		wallet, err := p.repo.ApplyOperation(job.ctx, job.op)
		job.result <- walletJobResult{wallet: wallet, err: err}
	}
}
