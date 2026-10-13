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

func (s *Store) GetUserByName(ctx context.Context, username string) (*usermodels.User, error) {
	var user usermodels.User
	err := s.tx.QueryRow(
		ctx,
		`
		SELECT users.user_id, wallet_id 
		FROM service.users 
		JOIN service.user_wallets 
		ON users.user_id = user_wallets.user_id 
		WHERE username=$1;`,
		username,
	).Scan(&user.ID, &user.WalletID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, usermodels.ErrUserNotFound
		}
	}

	return &user, nil
}

func (s *Store) GetUserFromSessionID(ctx context.Context, sessionIdHash string) (*usermodels.User, error) {
	user := &usermodels.User{}
	err := s.tx.QueryRow(ctx, `
	 SELECT user_id, wallet_id
	 FROM service.user_wallets
	 WHERE user_id=(
	 	SELECT user_id 
		FROM service.user_sessions 
		WHERE session_id_hash=$1 AND expires_at > now() AND revoked=false
	 );
	`, sessionIdHash).Scan(&user.ID, &user.WalletID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, errorx.Newf("failed to get user for session ID: %s", sessionIdHash, errorx.WithChildren(err))
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
	sessionIdHash string,
	user *usermodels.User,
	expiresAt time.Time,
) error {
	revoked, err := s.revokeSessionIfExists(ctx, user.ID)
	if err != nil {
		return err
	}
	if revoked {
		s.logger.DebugContext(ctx, "User session revoked", "user_id", user.ID)
	}

	_, err = s.tx.Exec(ctx, `
		INSERT INTO 
			service.user_sessions (session_id_hash, user_id, expires_at)
		VALUES
			($1, $2, $3)
		;`,
		sessionIdHash, user.ID, formatTimestamp(expiresAt),
	)
	if err != nil {
		return errorx.New("failed to insert session", errorx.WithChildren(err))
	}

	return nil
}

func (s *Store) RevokeSession(ctx context.Context, sessionIdHash string) error {
	_, err := s.tx.Exec(ctx, `UPDATE service.user_sessions SET revoked=true WHERE session_id_hash='$1';`, sessionIdHash)
	if err != nil {
		return errorx.Newf("failed to revoke session. Hash: %s", sessionIdHash, errorx.WithChildren(err))
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

func (s *Store) revokeSessionIfExists(ctx context.Context, userID string) (revoked bool, err error) {
	tag, err := s.tx.Exec(ctx, `UPDATE service.user_sessions SET revoked=true WHERE user_id=$1 AND revoked=false;`, userID)
	if err != nil {
		return false, err
	}

	revoked = tag.RowsAffected() > 0

	return
}

func formatTimestamp(timestamp time.Time) string {
	return timestamp.UTC().Format(time.DateTime)
}
