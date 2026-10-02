package walletmodels

import (
	"time"
)

type Wallet struct {
	ID           string
	BalanceMinor int64
}

func (w *Wallet) Debit(amountMinor int64) error {
	if amountMinor < 0 {
		return ErrInvalidAmount
	}

	newBalance := w.BalanceMinor - amountMinor
	if newBalance < 0 {
		return ErrInsufficientFunds
	}

	w.BalanceMinor = newBalance

	return nil
}

func (w *Wallet) Credit(amountMinor int64) error {
	if amountMinor < 0 {
		return ErrInvalidAmount
	}

	w.BalanceMinor += amountMinor

	return nil
}

func (w *Wallet) Clone() *Wallet {
	clone := &Wallet{
		ID:           w.ID,
		BalanceMinor: w.BalanceMinor,
	}

	return clone
}

type TransactionType string

const (
	TransactionTypeDebit  TransactionType = "DEBIT"
	TransactionTypeCredit TransactionType = "CREDIT"
)

type Transaction struct {
	ID          string
	Timestamp   time.Time
	Type        TransactionType
	WalletID    string
	AmountMinor int64
}
