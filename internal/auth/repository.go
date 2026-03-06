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