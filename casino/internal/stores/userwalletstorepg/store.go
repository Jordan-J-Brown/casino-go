package userwalletstorepg

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/usermodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/walletmodels"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/pgdb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionType = string

const (
	transactionTypeDebit  TransactionType = "DEBIT"
	transactionTypeCredit TransactionType = "CREDIT"
)

type Store struct {
	pool   *pgxpool.Pool
	logger logsx.Logger
}

func New(pool *pgxpool.Pool, logger logsx.Logger) *Store {
	return &Store{
		pool:   pool,
		logger: logger,
	}
}

func (s *Store) GetUserByID(ctx context.Context, id string) (user *usermodels.User, err error) {
	err = pgdb.RunTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		userExists, err := s.userExists(ctx, tx, id)
		if err != nil {
			return err
		}
		if !userExists {
			return usermodels.ErrUserNotFound
		}

		wallet, err := s.getWalletForUserID(ctx, tx, id)
		if err != nil {
			return err
		}
		if wallet == nil {
			return errorx.Newf("wallet does not exists for user with ID: %s", id, errorx.WithChildren(err))
		}

		user = &usermodels.User{
			ID:       id,
			WalletID: wallet.ID,
		}

		return nil
	})

	return
}

func (s *Store) GetUserFromSessionID(ctx context.Context, sessionID string) (*usermodels.User, error) {
	user := &usermodels.User{}
	err := s.pool.QueryRow(ctx, `
		SELECT user_id, wallet_id 
		FROM service.wallets 
		WHERE user_id=(SELECT user_id FROM service.user_sessions WHERE session_id=$1);	
	`, sessionID).Scan(&user.ID, &user.WalletID)
	if err != nil {
		return nil, errorx.New("failed to get user from session ID", errorx.WithChildren(err))
	}

	return user, nil
}

func (s *Store) GetWallet(ctx context.Context, walletID string) (*walletmodels.Wallet, error) {
	wallet := &walletmodels.Wallet{}
	err := s.pool.QueryRow(ctx, `SELECT wallet_id, balance_minor FROM service.wallets WHERE wallet_id=$1`, walletID).Scan(
		&wallet.ID,
		&wallet.BalanceMinor,
	)
	if err != nil {
		return nil, errorx.New("failed to retrieve wallet", errorx.WithChildren(err))
	}

	return wallet, nil
}

func (s *Store) CreateUserWithBalance(ctx context.Context, username string, balance int64) (user *usermodels.User, err error) {
	err = pgdb.RunTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		userID := uuid.New().String()
		walletID := uuid.New().String()

		usernameExists, err := s.usernameExists(ctx, tx, username)
		if err != nil {
			return err
		}
		if usernameExists {
			return usermodels.ErrUsernameExists
		}

		_, err = tx.Exec(ctx, `INSERT INTO service.users (user_id, username) VALUES ($1, $2);`, userID, username)
		if err != nil {
			fmt.Printf("Error type: %T\n", err)
			return errorx.New("failed to create user", errorx.WithChildren(err))
		}

		err = s.createWallet(ctx, tx, userID, walletID)
		if err != nil {
			return errorx.New("failed to create wallet for new user", errorx.WithChildren(err))
		}

		err = s.updateBalance(ctx, tx, &walletmodels.Wallet{
			ID:           walletID,
			BalanceMinor: balance,
		})
		if err != nil {
			return errorx.New("failed to update wallet balance for new user", errorx.WithChildren(err))
		}

		user = &usermodels.User{
			ID:       userID,
			WalletID: walletID,
		}

		return nil
	})

	return
}

func (s *Store) CreateSessionForUser(
	ctx context.Context,
	sessionID string,
	user *usermodels.User,
	expiresAt time.Time,
) error {
	return pgdb.RunTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		exists, err := s.sessionExists(ctx, tx, sessionID)
		if err != nil {
			return err
		}
		if exists {
			return usermodels.ErrSessionExists
		}

		_, err = tx.Exec(ctx, `
		INSERT INTO 
			user_sessions (session_id, user_id, expires_at)
		VALUES
			($1, $2, $3)
		;`,
			sessionID, user.ID, formatTimestamp(expiresAt),
		)
		if err != nil {
			return errorx.New("failed to insert session", errorx.WithChildren(err))
		}

		return nil
	})
}

func (s *Store) TransferFunds(
	ctx context.Context,
	tx pgx.Tx,
	args walletmodels.TransferArgs,
	transferFn func(src, dst *walletmodels.Wallet) error,
) error {
	srcWallet, err := s.getLockedWallet(ctx, tx, args.SourceWalletID)
	if err != nil {
		return err
	}
	if srcWallet == nil {
		return walletmodels.ErrSourceWalletDoesNotExist
	}

	dstWallet, err := s.getLockedWallet(ctx, tx, args.DestinationWalletID)
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

	if err = s.updateBalance(ctx, tx, srcWallet); err != nil {
		return errorx.New("failed to update source wallet balance", errorx.WithChildren(err))
	}
	if err = s.insertTransaction(ctx, tx, walletmodels.Transaction{
		ID:          uuid.New().String(),
		Timestamp:   args.Timestamp,
		Type:        walletmodels.TransactionTypeDebit,
		WalletID:    srcWallet.ID,
		AmountMinor: args.AmountMinor,
	}); err != nil {
		return errorx.New("failed to insert source wallet transaction", errorx.WithChildren(err))
	}

	if err = s.updateBalance(ctx, tx, dstWallet); err != nil {
		return errorx.New("failed to update desination wallet balance", errorx.WithChildren(err))
	}
	if err = s.insertTransaction(ctx, tx, walletmodels.Transaction{
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

func (s *Store) WalletExists(ctx context.Context, walletID string) (exists bool, err error) {
	err = s.pool.
		QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM service.wallets WHERE wallet_id=$1);`, walletID).
		Scan(&exists)
	return
}

func (s *Store) insertTransaction(
	ctx context.Context,
	tx pgx.Tx,
	transaction walletmodels.Transaction,
) error {

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

	_, err := tx.Exec(
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

func (s *Store) updateBalance(ctx context.Context, tx pgx.Tx, wallet *walletmodels.Wallet) error {
	_, err := tx.Exec(
		ctx,
		`UPDATE service.wallets SET balance_minor=$1 WHERE wallet_id=$2;`,
		wallet.BalanceMinor, wallet.ID,
	)

	return err
}

func (s *Store) createWallet(ctx context.Context, tx pgx.Tx, userID, walletID string) error {
	_, err := tx.Exec(ctx, `INSERT INTO service.wallets (wallet_id, user_id) VALUES ($1, $2);`, walletID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) deleteWallet(ctx context.Context, tx pgx.Tx, walletID string) error {
	_, err := tx.Exec(ctx, `UPDATE service.wallets SET soft_deleted = true WHERE wallet_id = $1`, walletID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) getLockedWallet(ctx context.Context, tx pgx.Tx, walletID string) (*walletmodels.Wallet, error) {
	var (
		id           string
		balanceMinor int64
	)
	err := tx.QueryRow(
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

func (s *Store) getWalletForUserID(ctx context.Context, tx pgx.Tx, userID string) (*walletmodels.Wallet, error) {
	var (
		id           string
		balanceMinor int64
	)
	err := tx.QueryRow(
		ctx,
		`SELECT wallet_id, balance_minor FROM service.wallets WHERE user_id = $1 AND soft_deleted=false;`,
		userID,
	).Scan(&id, &balanceMinor)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, errorx.Newf("failed to get wallet for user ID: '%s'", userID, errorx.WithChildren(err))
	}

	return &walletmodels.Wallet{
		ID:           id,
		BalanceMinor: balanceMinor,
	}, nil
}

func (s *Store) usernameExists(ctx context.Context, tx pgx.Tx, name string) (exists bool, err error) {
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM service.users WHERE username=$1);`, name).Scan(&exists)
	if err != nil {
		return false, errorx.New("failed to check if username exists", errorx.WithChildren(err))
	}
	return
}

func (s *Store) userExists(ctx context.Context, tx pgx.Tx, id string) (exists bool, err error) {
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM service.users WHERE user_id=$1);`, id).Scan(&exists)
	if err != nil {
		return false, errorx.New("failed to check user exists", errorx.WithChildren(err))
	}
	return exists, err
}

func (s *Store) sessionExists(ctx context.Context, tx pgx.Tx, sessionID string) (exists bool, err error) {
	err = tx.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM service.user_sessions WHERE session_id=$1 AND expires_at > now());`,
		sessionID,
	).Scan(&exists)
	if err != nil {
		err = errorx.New("failed to check if session exists", errorx.WithChildren(err))
	}
	return
}

func formatTimestamp(timestamp time.Time) string {
	return timestamp.UTC().Format(time.DateTime)
}
