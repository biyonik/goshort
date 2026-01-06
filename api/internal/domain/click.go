package domain

import "time"

// ClickEvent bir tıklama olayını temsil eder
type ClickEvent struct {
	ID        string    `json:"id"`
	URLCode   string    `json:"url_code"`
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Referer   string    `json:"referer"`
}
