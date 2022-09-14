package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"mascot/internal/db"
	"mascot/internal/domain"
	"mascot/internal/repositories"
)

type Wallet struct {
	transactor *db.Transactor
	walletRepo *repositories.Wallet
}

func NewWallet(transactor *db.Transactor, walletRepo *repositories.Wallet) *Wallet {
	return &Wallet{transactor: transactor, walletRepo: walletRepo}
}

func (w *Wallet) GetBalance(ctx context.Context, playerName, currency string) (int64, error) {
	wallet, err := w.walletRepo.GetWallet(ctx, playerName)
	if err != nil {
		return 0, err
	}

	if err := validateWallet(wallet, currency); err != nil {
		return 0, err
	}

	return wallet.Balance, nil
}

func (w *Wallet) WithdrawAndDeposit(ctx context.Context, transaction *domain.Transaction) error {
	handledTx, err := w.walletRepo.GetTransactionByExternalID(ctx, transaction.ExternalID)
	if err != nil {
		return err
	}

	if handledTx != nil && handledTx.RolledBack {
		return domain.ErrTransactionIsRolledBack
	}

	if handledTx != nil {
		*transaction = *handledTx
		return nil
	}

	var wallet *domain.Wallet
	err = w.transactor.WithTx(ctx, func(tCtx context.Context) error {
		var err error
		wallet, err = w.walletRepo.GetWallet(tCtx, transaction.PlayerName)
		if err != nil {
			return err
		}

		if err := validateTransaction(transaction); err != nil {
			return err
		}

		if err := validateWallet(wallet, transaction.Currency); err != nil {
			return err
		}

		if err := wallet.WithdrawAndDeposit(*transaction.Deposit, *transaction.Withdraw); err != nil {
			return err
		}

		if transaction.ID, err = generateTxID(); err != nil {
			return err
		}

		if err := w.walletRepo.InsertTransaction(tCtx, transaction); err != nil {
			return err
		}

		if err := w.walletRepo.UpdateBalance(tCtx, wallet); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (w *Wallet) RollbackTransaction(ctx context.Context, transaction *domain.Transaction) error {
	err := w.transactor.WithTx(ctx, func(tCtx context.Context) error {
		wallet, err := w.walletRepo.GetWallet(tCtx, transaction.PlayerName)
		if err != nil {
			return err
		}

		handledTx, err := w.walletRepo.GetTransactionByExternalID(tCtx, transaction.ExternalID)
		if err != nil {
			return err
		}

		if handledTx == nil {
			transaction.Currency = wallet.Currency
			transaction.ID, err = generateTxID()
			if err != nil {
				return err
			}
			return w.walletRepo.InsertTransaction(tCtx, transaction)
		}

		if handledTx.RolledBack {
			return nil
		}

		if err := wallet.Rollback(*handledTx.Deposit, *handledTx.Withdraw); err != nil {
			return err
		}

		if err := w.walletRepo.UpdateBalance(tCtx, wallet); err != nil {
			return err
		}

		return w.walletRepo.SetTransactionRolledBack(tCtx, handledTx.ID)
	})

	return err
}

func validateTransaction(transaction *domain.Transaction) error {
	if *transaction.Withdraw < 0 {
		return domain.ErrNegativeWithdrawal
	}

	if *transaction.Deposit < 0 {
		return domain.ErrNegativeDeposit
	}

	return nil
}

func validateWallet(wallet *domain.Wallet, currency string) error {
	if !strings.EqualFold(wallet.Currency, currency) {
		return domain.ErrIllegalCurrency
	}

	return nil
}

func generateTxID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("reading bytes slice: %w", err)
	}
	return hex.EncodeToString(b), nil
}
