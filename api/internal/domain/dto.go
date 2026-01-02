package domain

import "time"

type CreateURLRequest struct {
	LongURL       string     `json:"long_url"`
	UrlExpiration *time.Time `json:"url_expiration"`
}

type CreateURLResponse struct {
	ID            string     `json:"id"`
	Code          string     `json:"code"`
	ShortURL      string     `json:"short_url"`
	LongURL       string     `json:"long_url"`
	UrlExpiration *time.Time `json:"url_expiration"`
	CreatedAt     time.Time  `json:"created_at"`
}
