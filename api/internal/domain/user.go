package domain

import "time"

// UserTier kullanıcı seviyesi
type UserTier string

const (
	TierFree UserTier = "free"
	TierPro  UserTier = "pro"
)

// User kullanıcı modeli
type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // JSON'da gösterme
	Tier         UserTier   `json:"tier"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

// RegisterRequest kayıt isteği
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest giriş isteği
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse auth yanıtı (login/register)
type AuthResponse struct {
	Token     string   `json:"token"`
	ExpiresAt int64    `json:"expires_at"`
	User      UserInfo `json:"user"`
}

// UserInfo token içinde dönen kullanıcı bilgisi
type UserInfo struct {
	ID    string   `json:"id"`
	Email string   `json:"email"`
	Tier  UserTier `json:"tier"`
}
