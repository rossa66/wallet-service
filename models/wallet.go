package models

import "github.com/google/uuid"

type OperationType string

const (
	Deposit  OperationType = "DEPOSIT"
	Withdraw OperationType = "WITHDRAW"
)

type WalletOperation struct {
	WalletID      uuid.UUID    `json:"walletId"`
	OperationType OperationType `json:"operationType"`
	Amount        int64        `json:"amount"`
}

type Wallet struct {
	ID      uuid.UUID `json:"id"`
	Balance int64     `json:"balance"`
}

type WalletBalanceResponse struct {
	Balance int64 `json:"balance"`
}