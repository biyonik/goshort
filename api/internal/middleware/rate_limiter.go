package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/biyonik/goshort/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter Redis tabanlı sliding window rate limiter
type RateLimiter struct {
	client *redis.Client
	prefix string
}

// RateLimitConfig rate limit ayarları
type RateLimitConfig struct {
	Limit  int           // Maksimum istek sayısı
	Window time.Duration // Zaman penceresi
}

// TieredRateLimitConfig tier bazlı rate limit ayarları
type TieredRateLimitConfig struct {
	Free RateLimitConfig
	Pro  RateLimitConfig
}

// NewRateLimiter yeni rate limiter oluşturur
func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{
		client: client,
		prefix: "rate:",
	}
}

// keyFor Redis key'ini oluşturur
// Örnek: "rate:192.168.1.1:/api/urls"
func (r *RateLimiter) keyFor(ip, endpoint string) string {
	return fmt.Sprintf("%s%s:%s", r.prefix, ip, endpoint)
}

// Allow istek yapılabilir mi kontrol eder
// Returns: (izin verildi mi, kalan limit, error)
func (r *RateLimiter) Allow(ctx context.Context, ip, endpoint string, config RateLimitConfig) (bool, int, error) {
	key := r.keyFor(ip, endpoint)

	// 1. Sayacı artır
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, 0, fmt.Errorf("redis incr hatası: %w", err)
	}

	// 2. İlk istekse TTL ayarla
	if count == 1 {
		r.client.Expire(ctx, key, config.Window)
	}

	// 3. Limit kontrolü
	remaining := config.Limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	allowed := count <= int64(config.Limit)
	return allowed, remaining, nil
}

// Middleware Gin middleware'i döndürür (sabit limit)
func (r *RateLimiter) Middleware(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		endpoint := c.FullPath() // "/api/urls" veya "/:code"

		allowed, remaining, err := r.Allow(c.Request.Context(), ip, endpoint, config)
		if err != nil {
			// Redis hatası - isteği geçir (fail open)
			c.Next()
			return
		}

		// Rate limit header'ları ekle
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Window", config.Window.String())

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": config.Window.String(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TieredMiddleware tier bazlı rate limiting (Free vs Pro)
func (r *RateLimiter) TieredMiddleware(config TieredRateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		endpoint := c.FullPath()

		// Kullanıcı tier'ını al (context'ten)
		tier := domain.TierFree // Default: free
		if t, exists := c.Get("tier"); exists {
			tier = t.(domain.UserTier)
		}

		// Tier'a göre config seç
		var limitConfig RateLimitConfig
		if tier == domain.TierPro {
			limitConfig = config.Pro
		} else {
			limitConfig = config.Free
		}

		allowed, remaining, err := r.Allow(c.Request.Context(), ip, endpoint, limitConfig)
		if err != nil {
			c.Next()
			return
		}

		// Header'lar
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limitConfig.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Window", limitConfig.Window.String())
		c.Header("X-RateLimit-Tier", string(tier))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": limitConfig.Window.String(),
				"tier":        tier,
				"upgrade_tip": "Upgrade to Pro for higher limits!",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
