package domain

type Transaction struct {
	ID                 string
	PlayerName         string
	Withdraw           *int64
	Deposit            *int64
	Currency           string
	ExternalID         string
	BalanceAfterCommit *int64
	RolledBack         bool
}
