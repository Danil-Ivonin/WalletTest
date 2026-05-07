package handler

import (
	"errors"
	"net/http"

	"github.com/Danil-Ivonin/WalletTest/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func (h *Handler) PostWallet(c *gin.Context) {
	var req walletOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	//Is there a mistake in the technical specification? A field with the name valletID
	rawWalletID := req.WalletID
	if rawWalletID == "" {
		rawWalletID = req.ValletID
	}
	walletID, err := uuid.Parse(rawWalletID)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid wallet id")
		return
	}

	wallet, err := h.service.Wallet.ApplyOperation(c.Request.Context(), walletID, req.OperationType, req.Amount)
	if err != nil {
		writeDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, walletResponse{
		WalletID: wallet.ID.String(),
		Balance:  wallet.Balance,
	})
}

func (h *Handler) GetWalletBalance(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("walletUUID"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid wallet id")
		return
	}

	wallet, err := h.service.Wallet.GetBalance(c.Request.Context(), walletID)
	if err != nil {
		writeDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, walletResponse{
		WalletID: wallet.ID.String(),
		Balance:  wallet.Balance,
	})
}

func writeDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidAmount), errors.Is(err, domain.ErrInvalidOperationType):
		writeError(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrWalletNotFound):
		writeError(c, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrInsufficientFunds):
		writeError(c, http.StatusConflict, err.Error())
	default:
		writeError(c, http.StatusInternalServerError, "internal server error")
	}
}

func writeError(c *gin.Context, status int, message string) {
	logrus.Error(message)
	c.AbortWithStatusJSON(status, gin.H{
		"success": false,
		"message": message,
	})
}
