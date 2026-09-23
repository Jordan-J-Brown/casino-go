package userwalletstorepg

type ErrUnknownTransactionType string

func (err ErrUnknownTransactionType) Error() string {
	return "unknown transaction type: " + string(err)
}
