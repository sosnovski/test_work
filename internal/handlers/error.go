package handlers

import (
	"errors"
	"fmt"

	"mascot/internal/domain"
)

const (
	ErrParse              = -32700
	ErrInvalidParams      = -32602
	ErrMethodNotFound     = -32601
	ErrInvalidRequest     = -32600
	ErrInternalError      = -32603
	ErrDefaultServerError = -32000

	ErrNotEnoughMoneyCode          = 1
	ErrIllegalCurrencyCode         = 2
	ErrNegativeDepositCode         = 3
	ErrNegativeWithdrawalCode      = 4
	ErrSpendingBudgetExceeded      = 5
	ErrTransactionIsRolledBackCode = 6
)

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewError(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

func (e *Error) Error() string {
	return fmt.Sprintf("error: %s with code %d", e.Message, e.Code)
}

func (e *Error) WithData(data interface{}) *Error {
	e.Data = data
	return e
}

func MapDomainToTransportError(err error) error {
	switch {
	case errors.Is(err, domain.ErrIllegalCurrency):
		return NewError(ErrIllegalCurrencyCode, err.Error())
	case errors.Is(err, domain.ErrNotEnoughMoney):
		return NewError(ErrNotEnoughMoneyCode, err.Error())
	case errors.Is(err, domain.ErrNegativeWithdrawal):
		return NewError(ErrNegativeWithdrawalCode, err.Error())
	case errors.Is(err, domain.ErrNegativeDeposit):
		return NewError(ErrNegativeDepositCode, err.Error())
	case errors.Is(err, domain.ErrTransactionIsRolledBack):
		return NewError(ErrTransactionIsRolledBackCode, err.Error())
	default:
		return NewError(ErrDefaultServerError, err.Error())
	}
}
