package casinostorepg

import (
	"context"
	"time"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/walletmodels"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/jackc/pgx/v5"
)

func (s *Store) WalletExists(ctx context.Context, walletID string) (exists bool, err error) {
	err = s.tx.
		QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM service.wallets WHERE wallet_id=$1);`, walletID).
		Scan(&exists)
	return
}

func (s *Store) GetWallet(ctx context.Context, walletID string) (*walletmodels.Wallet, error) {
	wallet := &walletmodels.Wallet{}
	err := s.tx.QueryRow(ctx, `SELECT wallet_id, balance_minor FROM service.wallets WHERE wallet_id=$1`, walletID).Scan(
		&wallet.ID,
		&wallet.BalanceMinor,
	)
	if err != nil {
		return nil, errorx.New("failed to retrieve wallet", errorx.WithChildren(err))
	}

	return wallet, nil
}

func (store *Store) ApplyWalletTransaction(
	ctx context.Context,
	walletID string,
	applyFn func(wallet *walletmodels.Wallet) (*walletmodels.Transaction, error),
) error {
	wallet, err := store.getLockedWallet(ctx, walletID)
	if err != nil {
		return err
	}
	if wallet == nil {
		return walletmodels.ErrWalletDoesNotExist
	}

	walletTransaction, err := applyFn(wallet)
	if err != nil {
		return err
	}
	if walletTransaction == nil {
		return errorx.New("nil wallet transaction returned")
	}

	err = store.updateBalance(ctx, wallet)
	if err != nil {
		return errorx.New("failed to update source wallet balance", errorx.WithChildren(err))
	}

	err = store.insertTransaction(ctx, *walletTransaction)
	if err != nil {
		return errorx.New("failed to insert transaction", errorx.WithChildren(err))
	}

	return nil
}

func (s *Store) createWallet(ctx context.Context, walletID string) error {
	_, err := s.tx.Exec(ctx, `INSERT INTO service.wallets (wallet_id) VALUES ($1);`, walletID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) createWalletWithBalance(ctx context.Context, walletID string, balance int64) error {
	_, err := s.tx.Exec(ctx, `INSERT INTO service.wallets (wallet_id, balance_minor) VALUES ($1, $2);`, walletID, balance)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) getLockedWallet(ctx context.Context, walletID string) (*walletmodels.Wallet, error) {
	var (
		id           string
		balanceMinor int64
	)
	err := s.tx.QueryRow(
		ctx,
		`SELECT wallet_id, balance_minor FROM service.wallets WHERE wallet_id = $1 FOR UPDATE;`,
		walletID,
	).Scan(&id, &balanceMinor)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &walletmodels.Wallet{
		ID:           id,
		BalanceMinor: balanceMinor,
	}, nil
}

func (s *Store) updateBalance(ctx context.Context, wallet *walletmodels.Wallet) error {
	_, err := s.tx.Exec(
		ctx,
		`UPDATE service.wallets SET balance_minor=$1 WHERE wallet_id=$2;`,
		wallet.BalanceMinor, wallet.ID,
	)

	return err
}

func (s *Store) insertTransaction(ctx context.Context, transaction walletmodels.Transaction) error {

	timestamp := transaction.Timestamp.Format(time.RFC3339)
	var transactionType TransactionType
	switch transaction.Type {
	case walletmodels.TransactionTypeCredit:
		transactionType = transactionTypeCredit
	case walletmodels.TransactionTypeDebit:
		transactionType = transactionTypeDebit
	default:
		return ErrUnknownTransactionType(transaction.Type)
	}

	_, err := s.tx.Exec(
		ctx,
		`INSERT INTO 
			service.transactions (transaction_id, wallet_id, transaction_type, amount_minor, time_stamp)
		VALUES
			($1, $2, $3, $4, $5)
		;`,
		transaction.ID, transaction.WalletID, transactionType, transaction.AmountMinor, timestamp,
	)

	return err
}
