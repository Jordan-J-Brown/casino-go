package walletmodels

import "errors"

var (
	ErrInsufficientFunds        = errors.New("insufficient funds")
	ErrInvalidAmount            = errors.New("invalid amount")
	ErrSourceWalletDoesNotExist = errors.New("source wallet does not exist")
	ErrDestWalletDoesNotExist   = errors.New("destination wallet does not exist")
	ErrWalletExists             = errors.New("wallet already exists")
)
