package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"miyano/internal/adapter/postgres/gen"
	"miyano/internal/domain"
	"miyano/internal/ports"
)

// uniqueViolation is PostgreSQL's error code for unique constraint violations.
const uniqueViolation = "23505"

// UserAuthRepository implements ports.UserAuthRepository on PostgreSQL.
type UserAuthRepository struct {
	pool *pgxpool.Pool
}

// NewUserAuthRepository builds the repository from the connection pool.
// It takes the pool (not a DBTX) because register spans a transaction.
func NewUserAuthRepository(pool *pgxpool.Pool) *UserAuthRepository {
	return &UserAuthRepository{pool: pool}
}

// CreateUser inserts the user and its profile in one transaction.
func (r *UserAuthRepository) CreateUser(ctx context.Context, email, passwordHash, displayName string) (domain.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.User{}, fmt.Errorf("beginning register transaction: %w", err)
	}
	defer tx.Rollback(ctx) // no-op once committed

	q := gen.New(tx)
	row, err := q.CreateUser(ctx, gen.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  &displayName,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domain.User{}, ports.ErrEmailTaken
		}
		return domain.User{}, fmt.Errorf("creating user: %w", err)
	}
	if err := q.CreateProfile(ctx, gen.CreateProfileParams{UserID: row.ID, DisplayName: displayName}); err != nil {
		return domain.User{}, fmt.Errorf("creating profile: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.User{}, fmt.Errorf("committing register transaction: %w", err)
	}

	return domain.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		DisplayName:  displayName,
	}, nil
}

// GetUserByEmail returns the user or ports.ErrNotFound.
func (r *UserAuthRepository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	row, err := gen.New(r.pool).GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ports.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("getting user by email: %w", err)
	}
	return domain.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		DisplayName:  derefString(row.DisplayName),
	}, nil
}

// CreateRefreshToken stores the token hash with its expiry.
func (r *UserAuthRepository) CreateRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	err := gen.New(r.pool).CreateRefreshToken(ctx, gen.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("creating refresh token: %w", err)
	}
	return nil
}

// GetRefreshTokenByHash returns the token row or ports.ErrNotFound.
func (r *UserAuthRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error) {
	row, err := gen.New(r.pool).GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.RefreshToken{}, ports.ErrNotFound
		}
		return domain.RefreshToken{}, fmt.Errorf("getting refresh token: %w", err)
	}
	return domain.RefreshToken{
		UserID:    row.UserID,
		ExpiresAt: row.ExpiresAt.Time,
		RevokedAt: timestamptzToPtr(row.RevokedAt),
	}, nil
}

// RevokeRefreshToken marks a still-active token as revoked. Tokens already
// revoked or unknown return ports.ErrRefreshRejected.
func (r *UserAuthRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	n, err := gen.New(r.pool).RevokeRefreshToken(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("revoking refresh token: %w", err)
	}
	if n == 0 {
		return ports.ErrRefreshRejected
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func timestamptzToPtr(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}
