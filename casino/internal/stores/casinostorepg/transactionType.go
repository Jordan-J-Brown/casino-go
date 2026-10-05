package casinostorepg

type TransactionType = string

const (
	transactionTypeDebit  TransactionType = "DEBIT"
	transactionTypeCredit TransactionType = "CREDIT"
)

type ErrUnknownTransactionType string

func (err ErrUnknownTransactionType) Error() string {
	return "unknown transaction type: " + string(err)
}
