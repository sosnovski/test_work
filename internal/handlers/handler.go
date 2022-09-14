package handlers

import (
	"context"

	"mascot/internal/domain"
	"mascot/internal/services"
)

type Handler struct {
	walletService *services.Wallet
}

func NewHandler(walletService *services.Wallet) *Handler {
	return &Handler{walletService: walletService}
}

func (h *Handler) GetBalance(ctx context.Context, req *GetBalanceRequest) (*GetBalanceResponse, error) {
	balance, err := h.walletService.GetBalance(ctx, req.PlayerName, req.Currency)
	if err != nil {
		return nil, MapDomainToTransportError(err)
	}

	return &GetBalanceResponse{
		Balance: balance,
	}, nil
}

func (h *Handler) WithdrawAndDeposit(ctx context.Context, req *WithdrawAndDepositRequest) (*WithdrawAndDepositResponse, error) {
	tx := &domain.Transaction{
		PlayerName: req.PlayerName,
		Withdraw:   req.Withdraw,
		Deposit:    req.Deposit,
		Currency:   req.Currency,
		ExternalID: req.TransactionRef,
	}

	if err := h.walletService.WithdrawAndDeposit(ctx, tx); err != nil {
		return nil, MapDomainToTransportError(err)
	}

	return &WithdrawAndDepositResponse{
		NewBalance:    *tx.BalanceAfterCommit,
		TransactionID: tx.ID,
	}, nil
}

func (h *Handler) RollbackTransaction(ctx context.Context, req *RollbackTransactionRequest) error {
	tx := &domain.Transaction{
		PlayerName: req.PlayerName,
		ExternalID: req.TransactionRef,
		RolledBack: true,
	}

	return h.walletService.RollbackTransaction(ctx, tx)
}
