package casinostorepg

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/usermodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/walletmodels"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/jackc/pgx/v5"
)

type TransactionType = string

const (
	transactionTypeDebit  TransactionType = "DEBIT"
	transactionTypeCredit TransactionType = "CREDIT"
)

type ErrUnknownTransactionType string

func (err ErrUnknownTransactionType) Error() string {
	return "unknown transaction type: " + string(err)
}

func (s *Store) GetUserByID(ctx context.Context, id string) (user *usermodels.User, err error) {
	userExists, err := s.userExists(ctx, id)
	if err != nil {
		return nil, err
	}
	if !userExists {
		return nil, usermodels.ErrUserNotFound
	}

	wallet, err := s.getWalletForUserID(ctx, id)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, errorx.Newf("wallet does not exists for user with ID: %s", id, errorx.WithChildren(err))
	}

	user = &usermodels.User{
		ID:       id,
		WalletID: wallet.ID,
	}

	return user, nil
}

func (s *Store) GetUserFromSessionID(ctx context.Context, sessionID string) (*usermodels.User, error) {
	user := &usermodels.User{}
	err := s.tx.QueryRow(ctx, `
	 SELECT user_id, wallet_id
	 FROM service.user_wallets
	 WHERE user_id=(
	 	SELECT user_id 
		FROM service.user_sessions 
		WHERE session_id=$1 AND expires_at > now()
	 );
	`, sessionID).Scan(&user.ID, &user.WalletID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, errorx.Newf("failed to get user for session ID: %s", sessionID, errorx.WithChildren(err))
	}

	return user, nil
}

func (s *Store) CreateUserWithBalance(ctx context.Context, username string, balance int64) (user *usermodels.User, err error) {
	userID := uuid.New().String()
	walletID := uuid.New().String()

	usernameExists, err := s.usernameExists(ctx, username)
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, usermodels.ErrUsernameExists
	}

	_, err = s.tx.Exec(ctx, `INSERT INTO service.users (user_id, username) VALUES ($1, $2);`, userID, username)
	if err != nil {
		fmt.Printf("Error type: %T\n", err)
		return nil, errorx.New("failed to create user", errorx.WithChildren(err))
	}

	err = s.createWalletWithBalance(ctx, walletID, balance)
	if err != nil {
		return nil, errorx.New("failed to create wallet for new user", errorx.WithChildren(err))
	}

	_, err = s.tx.Exec(
		ctx,
		`INSERT INTO service.user_wallets (user_id, wallet_id) VALUES ($1, $2);`,
		userID, walletID,
	)

	user = &usermodels.User{
		ID:       userID,
		WalletID: walletID,
	}

	return user, nil
}

func (s *Store) CreateSessionForUser(
	ctx context.Context,
	sessionID string,
	user *usermodels.User,
	expiresAt time.Time,
) error {
	exists, err := s.sessionExists(ctx, sessionID)
	if err != nil {
		return err
	}
	if exists {
		return usermodels.ErrSessionExists
	}

	_, err = s.tx.Exec(ctx, `
		INSERT INTO 
			service.user_sessions (session_id, user_id, expires_at)
		VALUES
			($1, $2, $3)
		;`,
		sessionID, user.ID, formatTimestamp(expiresAt),
	)
	if err != nil {
		return errorx.New("failed to insert session", errorx.WithChildren(err))
	}

	return nil
}

func (s *Store) getWalletForUserID(ctx context.Context, userID string) (*walletmodels.Wallet, error) {
	var (
		id           string
		balanceMinor int64
	)
	err := s.tx.QueryRow(
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

func (s *Store) usernameExists(ctx context.Context, name string) (exists bool, err error) {
	err = s.tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM service.users WHERE username=$1);`, name).Scan(&exists)
	if err != nil {
		return false, errorx.New("failed to check if username exists", errorx.WithChildren(err))
	}
	return
}

func (s *Store) userExists(ctx context.Context, id string) (exists bool, err error) {
	err = s.tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM service.users WHERE user_id=$1);`, id).Scan(&exists)
	if err != nil {
		return false, errorx.New("failed to check user exists", errorx.WithChildren(err))
	}
	return exists, err
}

func (s *Store) sessionExists(ctx context.Context, sessionID string) (exists bool, err error) {
	err = s.tx.QueryRow(
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
