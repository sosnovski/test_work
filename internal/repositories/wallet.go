package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"

	"mascot/internal/domain"
)

type Wallet struct {
	querier Querier
}

func NewWallet(querier Querier) *Wallet {
	return &Wallet{querier}
}

func (w *Wallet) GetWallet(ctx context.Context, playerName string) (*domain.Wallet, error) {
	row := w.querier.Conn(ctx).QueryRow(ctx,
		"SELECT id, player_name, currency, balance FROM wallets WHERE player_name = $1 FOR UPDATE",
		playerName,
	)

	res := &domain.Wallet{}
	if err := row.Scan(&res.ID, &res.UserName, &res.Currency, &res.Balance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWalletNotFound
		}
		return nil, err
	}

	return res, nil
}

func (w *Wallet) UpdateBalance(ctx context.Context, wallet *domain.Wallet) error {
	_, err := w.querier.Conn(ctx).Exec(ctx,
		"UPDATE wallets SET balance = $1 WHERE player_name = $2",
		wallet.Balance, wallet.UserName,
	)

	return err
}

func (w *Wallet) GetTransactionByExternalID(ctx context.Context, externalID string) (*domain.Transaction, error) {
	row := w.querier.Conn(ctx).QueryRow(ctx, "SELECT id, player_name, withdraw, deposit, "+
		"currency, external_id, balance_after_commit, rolled_back FROM transactions WHERE  external_id = $1",
		externalID,
	)

	tx := &domain.Transaction{}
	err := row.Scan(
		&tx.ID,
		&tx.PlayerName,
		&tx.Withdraw,
		&tx.Deposit,
		&tx.Currency,
		&tx.ExternalID,
		&tx.BalanceAfterCommit,
		&tx.RolledBack,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (w *Wallet) SetTransactionRolledBack(ctx context.Context, txID string) error {
	_, err := w.querier.Conn(ctx).Exec(ctx, "UPDATE transactions SET rolled_back = TRUE WHERE id = $1", txID)
	return err
}

func (w *Wallet) InsertTransaction(ctx context.Context, tx *domain.Transaction) error {
	_, err := w.querier.Conn(ctx).Exec(ctx,
		"INSERT INTO transactions (id, player_name, withdraw, deposit, currency, external_id, rolled_back) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING",
		tx.ID, tx.PlayerName, tx.Withdraw, tx.Deposit, tx.Currency, tx.ExternalID, tx.RolledBack,
	)

	if err != nil {
		return err
	}

	return nil
}
