package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/biyonik/goshort/internal/domain"
	"github.com/redis/go-redis/v9"
)

// Cache'te saklanan URL yapısı
// Sadece gerekli alanları tutuyoruz (memory tasarrufu)
type cachedURL struct {
	LongURL   string     `json:"long_url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// CacheRepository Redis cache işlemleri
type CacheRepository struct {
	client     *redis.Client
	defaultTTL time.Duration
}

// ErrCacheMiss cache'te bulunamadı
var ErrCacheMiss = errors.New("cache miss")

// NewCacheRepository yeni cache repository oluşturur
func NewCacheRepository(client *redis.Client, defaultTTL time.Duration) *CacheRepository {
	return &CacheRepository{
		client:     client,
		defaultTTL: defaultTTL, // Varsayılan: 24 saat
	}
}

// keyFor Redis key'ini oluşturur
// Örnek: "url:abc123"
func (r *CacheRepository) keyFor(code string) string {
	return "url:" + code
}

// SetURL URL'i cache'e yazar
func (r *CacheRepository) SetURL(ctx context.Context, url *domain.URL) error {
	// 1. Cache'te tutulacak veriyi hazırla
	cached := cachedURL{
		LongURL:   url.LongURL,
		ExpiresAt: url.UrlExpiration,
	}

	// 2. JSON'a çevir
	data, err := json.Marshal(cached)
	if err != nil {
		return fmt.Errorf("json marshal hatası: %w", err)
	}

	// 3. TTL hesapla
	ttl := r.calculateTTL(url.UrlExpiration)

	// 4. Redis'e yaz
	key := r.keyFor(url.Code)
	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("redis set hatası: %w", err)
	}

	return nil
}

// GetURL URL'i cache'ten okur
func (r *CacheRepository) GetURL(ctx context.Context, code string) (*domain.URL, error) {
	// 1. Redis'ten oku
	key := r.keyFor(code)
	data, err := r.client.Get(ctx, key).Result()

	// 2. Bulunamadıysa cache miss
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("redis get hatası: %w", err)
	}

	// 3. JSON'dan parse et
	var cached cachedURL
	if err := json.Unmarshal([]byte(data), &cached); err != nil {
		return nil, fmt.Errorf("json unmarshal hatası: %w", err)
	}

	// 4. domain.URL'e dönüştür
	url := &domain.URL{
		Code:          code,
		LongURL:       cached.LongURL,
		UrlExpiration: cached.ExpiresAt,
	}

	return url, nil
}

// DeleteURL URL'i cache'ten siler
func (r *CacheRepository) DeleteURL(ctx context.Context, code string) error {
	key := r.keyFor(code)
	return r.client.Del(ctx, key).Err()
}

// calculateTTL cache TTL'ini hesaplar
// Production yaklaşımı: min(defaultTTL, timeUntilExpiration)
func (r *CacheRepository) calculateTTL(expiresAt *time.Time) time.Duration {
	// URL'in expiration'ı yoksa varsayılan TTL kullan
	if expiresAt == nil {
		return r.defaultTTL
	}

	// URL ne kadar süre sonra expire olacak?
	timeUntilExpiry := time.Until(*expiresAt)

	// Zaten expire olmuşsa kısa bir TTL ver (1 dakika)
	// Bu sayede expire kontrolü yapılabilir
	if timeUntilExpiry <= 0 {
		return time.Minute
	}

	// min(defaultTTL, timeUntilExpiry)
	if timeUntilExpiry < r.defaultTTL {
		return timeUntilExpiry
	}

	return r.defaultTTL
}
