package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/biyonik/goshort/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ClickRepository click event'leri kaydeder
type ClickRepository struct {
	db *pgxpool.Pool
}

// NewClickRepository yeni click repository oluşturur
func NewClickRepository(db *pgxpool.Pool) *ClickRepository {
	return &ClickRepository{db: db}
}

// SaveBatch birden fazla click'i tek sorguda kaydeder
func (r *ClickRepository) SaveBatch(ctx context.Context, events []domain.ClickEvent) error {
	if len(events) == 0 {
		return nil
	}

	// Batch INSERT sorgusu oluştur
	// INSERT INTO clicks (id, url_code, ...) VALUES ($1, $2, ...), ($3, $4, ...), ...

	valueStrings := make([]string, 0, len(events))
	valueArgs := make([]interface{}, 0, len(events)*6)

	for i, event := range events {
		// Her event için ($1, $2, $3, $4, $5, $6) placeholder
		base := i * 6
		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6,
		))

		// ID yoksa oluştur
		id := event.ID
		if id == "" {
			id = uuid.Must(uuid.NewV7()).String()
		}

		valueArgs = append(valueArgs,
			id,
			event.URLCode,
			event.Timestamp,
			event.IP,
			event.UserAgent,
			event.Referer,
		)
	}

	query := fmt.Sprintf(
		`INSERT INTO clicks (id, url_code, timestamp, ip, user_agent, referer) VALUES %s`,
		strings.Join(valueStrings, ", "),
	)

	_, err := r.db.Exec(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("batch insert hatası: %w", err)
	}

	return nil
}
