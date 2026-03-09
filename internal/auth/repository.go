package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)
type User struct {
	ID                 string
	Email              string
	PasswordHash       string
	IsVerified         bool
	VerificationCode   string
	ExpiresAt          time.Time
}
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}


func (r *Repository) Create(ctx context.Context, user *User) error {
	
	query := `
		INSERT INTO users (email, password_hash, verification_code, verification_expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, query,
		user.Email,
		user.PasswordHash,
		user.VerificationCode,
		user.ExpiresAt,
	)
	return err
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, is_verified 
		FROM users 
		WHERE email = $1
	`
	
	user := &User{}
	
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, 
		&user.Email, 
		&user.PasswordHash, 
		&user.IsVerified,
	)
	
	if err != nil {
		return nil, err 
	}
	
	return user, nil
}

func (r *Repository) GetUserByEmailForVerification(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, is_verified, verification_code, verification_expires_at 
		FROM users 
		WHERE email = $1
	`
	
	user := &User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, 
		&user.Email, 
		&user.IsVerified, 
		&user.VerificationCode, 
		&user.ExpiresAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (r *Repository) MarkUserAsVerified(ctx context.Context, email string) error {
	query := `
		UPDATE users 
		SET is_verified = TRUE, verification_code = '', verification_expires_at = NOW() 
		WHERE email = $1
	`
	_, err := r.db.Exec(ctx, query, email)
	return err
}

func (r *Repository) UpdateVerificationCode(ctx context.Context, email string, code string, expiresAt time.Time) error {
	query := `
		UPDATE users 
		SET verification_code = $1, verification_expires_at = $2 
		WHERE email = $3
	`
	_, err := r.db.Exec(ctx, query, code, expiresAt, email)
	return err
}

func (r *Repository) UpdatePassword(ctx context.Context, email string, newPasswordHash string) error {
	query := `
		UPDATE users 
		SET password_hash = $1, verification_code = '', verification_expires_at = NOW() 
		WHERE email = $2
	`
	_, err := r.db.Exec(ctx, query, newPasswordHash, email)
	return err
}