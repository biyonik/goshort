package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/biyonik/goshort/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrUserNotFound kullanıcı bulunamadı
var ErrUserNotFound = errors.New("user not found")

// ErrEmailExists email zaten kayıtlı
var ErrEmailExists = errors.New("email already exists")

// UserRepository kullanıcı veritabanı işlemleri
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository yeni user repository oluşturur
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create yeni kullanıcı oluşturur
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, tier, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Tier,
		user.CreatedAt,
	)

	if err != nil {
		// Duplicate email kontrolü
		if isDuplicateKeyError(err) {
			return ErrEmailExists
		}
		return fmt.Errorf("user create hatası: %w", err)
	}

	return nil
}

// FindByEmail email ile kullanıcı bulur
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, tier, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Tier,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user find hatası: %w", err)
	}

	return &user, nil
}

// FindByID ID ile kullanıcı bulur
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, tier, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Tier,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user find hatası: %w", err)
	}

	return &user, nil
}

// isDuplicateKeyError PostgreSQL duplicate key hatası kontrolü
func isDuplicateKeyError(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate key") || contains(err.Error(), "unique constraint"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
