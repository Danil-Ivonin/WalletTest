package httpadapter

import "github.com/Danil-Ivonin/WalletTest/internal/domain"

type walletOperationRequest struct {
	// Is there a mistake in the technical specification? A field with the name valletID
	ValletID      string               `json:"valletId"`
	WalletID      string               `json:"walletId"`
	OperationType domain.OperationType `json:"operationType"`
	Amount        int64                `json:"amount"`
}

type walletResponse struct {
	WalletID string `json:"walletId"`
	Balance  int64  `json:"balance"`
}
