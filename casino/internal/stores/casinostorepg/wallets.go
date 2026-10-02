package casinostorepg

import (
	"context"
	"time"
	"uuid"

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

func (s *Store) GetOrCreateCasinoWallet(ctx context.Context, walletID string, balance int64) (*walletmodels.Wallet, error) {
	exists, err := s.CasinoWalletExists(ctx, walletID)
	if err != nil {
		return nil, err
	}
	if exists {
		wallet, err := s.GetWallet(ctx, walletID)
		if err != nil {
			return nil, err
		}
		return wallet, nil
	}

	return s.CreateCasinoWallet(ctx, walletID, balance)
}

func (s *Store) CasinoWalletExists(ctx context.Context, walletID string) (exists bool, err error) {
	err = s.tx.QueryRow(
		ctx, `SELECT EXISTS(SELECT 1 FROM service.casino_wallets WHERE wallet_id=$1);`, walletID,
	).Scan(&exists)
	if err != nil && err == pgx.ErrNoRows {
		return false, nil
	}

	return
}

func (s *Store) CreateCasinoWallet(ctx context.Context, walletID string, balance int64) (*walletmodels.Wallet, error) {
	err := s.createWalletWithBalance(ctx, walletID, balance)
	if err != nil {
		return nil, err
	}

	_, err = s.tx.Exec(ctx, `INSERT INTO service.casino_wallets (wallet_id) VALUES ($1);`, walletID)
	if err != nil {
		return nil, err
	}

	return &walletmodels.Wallet{
		ID:           walletID,
		BalanceMinor: balance,
	}, nil
}

func (s *Store) TransferFunds(
	ctx context.Context,
	args walletmodels.TransferArgs,
	transferFn func(src, dst *walletmodels.Wallet) error,
) error {
	srcWallet, err := s.getLockedWallet(ctx, args.SourceWalletID)
	if err != nil {
		return err
	}
	if srcWallet == nil {
		return walletmodels.ErrSourceWalletDoesNotExist
	}

	dstWallet, err := s.getLockedWallet(ctx, args.DestinationWalletID)
	if err != nil {
		return err
	}
	if dstWallet == nil {
		return walletmodels.ErrDestWalletDoesNotExist
	}

	err = transferFn(srcWallet, dstWallet)
	if err != nil {
		return err
	}

	if err = s.updateBalance(ctx, srcWallet); err != nil {
		return errorx.New("failed to update source wallet balance", errorx.WithChildren(err))
	}
	if err = s.insertTransaction(ctx, walletmodels.Transaction{
		ID:          uuid.New().String(),
		Timestamp:   args.Timestamp,
		Type:        walletmodels.TransactionTypeDebit,
		WalletID:    srcWallet.ID,
		AmountMinor: args.AmountMinor,
	}); err != nil {
		return errorx.New("failed to insert source wallet transaction", errorx.WithChildren(err))
	}

	if err = s.updateBalance(ctx, dstWallet); err != nil {
		return errorx.New("failed to update desination wallet balance", errorx.WithChildren(err))
	}
	if err = s.insertTransaction(ctx, walletmodels.Transaction{
		ID:          uuid.New().String(),
		Timestamp:   args.Timestamp,
		Type:        walletmodels.TransactionTypeCredit,
		WalletID:    dstWallet.ID,
		AmountMinor: args.AmountMinor,
	}); err != nil {
		return errorx.New("failed to insert destination wallet transaction", errorx.WithChildren(err))
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
