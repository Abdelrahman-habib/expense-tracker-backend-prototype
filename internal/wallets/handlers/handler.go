package handlers

import (
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/handlers"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/service"
	"go.uber.org/zap"
)

type WalletHandler struct {
	handlers.BaseHandler
	service service.WalletService
}

func NewWalletHandler(service service.WalletService, logger *zap.Logger) *WalletHandler {
	return &WalletHandler{
		BaseHandler: handlers.NewBaseHandler(logger),
		service:     service,
	}
}
