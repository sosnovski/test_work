package handlers

type GetBalanceRequest struct {
	PlayerName string `json:"playerName" validate:"required"`
	Currency   string `json:"currency" validate:"required"`
	GameID     string `json:"gameId"`
}

type GetBalanceResponse struct {
	Balance int64 `json:"balance"`
}

type WithdrawAndDepositRequest struct {
	PlayerName     string `json:"playerName" validate:"required"`
	Withdraw       *int64 `json:"withdraw" validate:"required"`
	Deposit        *int64 `json:"deposit" validate:"required"`
	Currency       string `json:"currency" validate:"required"`
	TransactionRef string `json:"transactionRef" validate:"required"`
}

type WithdrawAndDepositResponse struct {
	NewBalance    int64  `json:"newBalance"`
	TransactionID string `json:"transactionId"`
}

type RollbackTransactionRequest struct {
	PlayerName     string `json:"playerName" validate:"required"`
	TransactionRef string `json:"transactionRef" validate:"required"`
}
