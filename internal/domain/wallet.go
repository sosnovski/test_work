package domain

type Wallet struct {
	ID       int64
	UserName string
	Currency string
	Balance  int64
}

func (w *Wallet) WithdrawAndDeposit(deposit, withdraw int64) error {
	if w.Balance < withdraw {
		return ErrNotEnoughMoney
	}

	w.Balance -= withdraw
	w.Balance += deposit
	return nil
}

func (w *Wallet) Rollback(deposit, withdraw int64) error {
	if w.Balance < deposit {
		return ErrNotEnoughMoney
	}

	w.Balance -= deposit
	w.Balance += withdraw
	return nil
}
