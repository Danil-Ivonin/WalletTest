package handler

import "github.com/Danil-Ivonin/WalletTest/internal/service"
import "github.com/gin-gonic/gin"

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	api := router.Group("/api/v1")
	api.POST("/wallet", h.PostWallet)
	api.GET("/wallets/:walletUUID", h.GetWalletBalance)

	return router
}
