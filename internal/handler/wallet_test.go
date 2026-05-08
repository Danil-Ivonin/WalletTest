package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Danil-Ivonin/WalletTest/internal/domain"
	"github.com/Danil-Ivonin/WalletTest/internal/service"
	"github.com/google/uuid"
)

var errUnexpected = errors.New("unexpected")

func TestStatusFromDomainError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "invalid amount", err: domain.ErrInvalidAmount, want: http.StatusBadRequest},
		{name: "invalid operation type", err: domain.ErrInvalidOperationType, want: http.StatusBadRequest},
		{name: "wallet not found", err: domain.ErrWalletNotFound, want: http.StatusNotFound},
		{name: "insufficient funds", err: domain.ErrInsufficientFunds, want: http.StatusConflict},
		{name: "balance overflow", err: domain.ErrBalanceOverflow, want: http.StatusConflict},
		{name: "wallet queue full", err: service.ErrWalletQueueFull, want: http.StatusTooManyRequests},
		{name: "request canceled", err: context.Canceled, want: http.StatusRequestTimeout},
		{name: "request deadline exceeded", err: context.DeadlineExceeded, want: http.StatusTooManyRequests},
		{name: "unknown error", err: errUnexpected, want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusFromDomainError(tt.err); got != tt.want {
				t.Fatalf("statusFromDomainError() = %d, want %d", got, tt.want)
			}
		})
	}
}

type fakeWalletService struct {
	applyWallet domain.Wallet
	applyErr    error
	getWallet   domain.Wallet
	getErr      error

	lastApplyWalletID uuid.UUID
	lastApplyOpType   domain.OperationType
	lastApplyAmount   int64
	lastGetWalletID   uuid.UUID
}

func (f *fakeWalletService) ApplyOperation(ctx context.Context, walletID uuid.UUID, opType domain.OperationType, amount int64) (domain.Wallet, error) {
	f.lastApplyWalletID = walletID
	f.lastApplyOpType = opType
	f.lastApplyAmount = amount
	if f.applyErr != nil {
		return domain.Wallet{}, f.applyErr
	}
	if f.applyWallet.ID == uuid.Nil {
		f.applyWallet.ID = walletID
	}
	return f.applyWallet, nil
}

func (f *fakeWalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (domain.Wallet, error) {
	f.lastGetWalletID = walletID
	if f.getErr != nil {
		return domain.Wallet{}, f.getErr
	}
	if f.getWallet.ID == uuid.Nil {
		f.getWallet.ID = walletID
	}
	return f.getWallet, nil
}

func TestPostWalletAppliesOperation(t *testing.T) {
	walletID := uuid.New()
	fake := &fakeWalletService{applyWallet: domain.Wallet{ID: walletID, Balance: 1000}}
	h := NewHandler(&service.Service{Wallet: fake})
	router := h.InitRoutes()

	body := []byte(`{"valletId":"` + walletID.String() + `","operationType":"DEPOSIT","amount":1000}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if fake.lastApplyWalletID != walletID || fake.lastApplyOpType != domain.OperationTypeDeposit || fake.lastApplyAmount != 1000 {
		t.Fatalf("captured operation = id:%s type:%s amount:%d", fake.lastApplyWalletID, fake.lastApplyOpType, fake.lastApplyAmount)
	}

	var response walletResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.WalletID != walletID.String() || response.Balance != 1000 {
		t.Fatalf("response = %#v", response)
	}
}

func TestPostWalletValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{name: "invalid json", body: `{`, want: http.StatusBadRequest},
		{name: "invalid wallet id", body: `{"valletId":"bad","operationType":"DEPOSIT","amount":1000}`, want: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&service.Service{Wallet: &fakeWalletService{}})
			router := h.InitRoutes()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestGetWalletBalanceReturnsBalance(t *testing.T) {
	walletID := uuid.New()
	fake := &fakeWalletService{getWallet: domain.Wallet{ID: walletID, Balance: 777}}
	h := NewHandler(&service.Service{Wallet: fake})
	router := h.InitRoutes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+walletID.String(), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if fake.lastGetWalletID != walletID {
		t.Fatalf("lastGetWalletID = %s, want %s", fake.lastGetWalletID, walletID)
	}
}

func TestGetWalletBalanceRejectsInvalidID(t *testing.T) {
	h := NewHandler(&service.Service{Wallet: &fakeWalletService{}})
	router := h.InitRoutes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/bad-id", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
