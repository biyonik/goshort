package repository

import (
	"context"

	"github.com/biyonik/goshort/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type URLRepository struct {
	db *pgxpool.Pool
}

func NewURLRepository(db *pgxpool.Pool) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) Create(ctx context.Context, url *domain.URL) error {
	query := `
		INSERT INTO urls (id, long_url, code, url_expiration, click_count, user_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query,
		url.ID,
		url.LongURL,
		url.Code,
		url.UrlExpiration,
		url.ClickCount,
		url.UserID,
		url.CreatedAt,
	)

	return err
}

func (r *URLRepository) FindByShortCode(ctx context.Context, code string) (*domain.URL, error) {
	query := `
		SELECT id, long_url, code, url_expiration, click_count, user_id, created_at, updated_at
		FROM urls 
		WHERE code = $1 AND deleted_at IS NULL
	`

	var url domain.URL
	err := r.db.QueryRow(ctx, query, code).Scan(
		&url.ID,
		&url.LongURL,
		&url.Code,
		&url.UrlExpiration,
		&url.ClickCount,
		&url.UserID,
		&url.CreatedAt,
		&url.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (r *URLRepository) GetNextCounter(ctx context.Context) (uint64, error) {
	var counter uint64
	err := r.db.QueryRow(ctx, "SELECT nextval('url_counter_seq')").Scan(&counter)
	return counter, err
}

func (r *URLRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE code = $1 AND deleted_at IS NULL)`
	err := r.db.QueryRow(ctx, query, code).Scan(&exists)
	return exists, err
}
