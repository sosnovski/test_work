package domain

import "errors"

var (
	ErrWalletNotFound          = errors.New("wallet not found")
	ErrNotEnoughMoney          = errors.New("not enough money")
	ErrIllegalCurrency         = errors.New("illegal currency")
	ErrNegativeWithdrawal      = errors.New("negative withdrawal")
	ErrNegativeDeposit         = errors.New("negative deposit")
	ErrTransactionIsRolledBack = errors.New("transaction is rolled back")
)
