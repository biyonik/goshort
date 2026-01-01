package domain

import "time"

type URL struct {
	ID            string     `json:"id"`
	LongURL       string     `json:"long_url"`
	Code          string     `json:"code"`
	UrlExpiration *time.Time `json:"url_expiration"`
	ClickCount    int64      `json:"click_count"`
	UserID        string     `json:"user_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
}
