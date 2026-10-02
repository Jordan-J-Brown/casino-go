package walletmodels

import "errors"

var (
	ErrInsufficientFunds  = errors.New("INSUFFICIENT_FUNDS")
	ErrInvalidAmount      = errors.New("INVALID_AMOUNT")
	ErrWalletDoesNotExist = errors.New("WALLET_DOES_NOT_EXIST")
	ErrWalletExists       = errors.New("WALLET_EXISTS")
)
