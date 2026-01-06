package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis bağlantısını tutar
type RedisClient struct {
	Client *redis.Client
}

// RedisConfig bağlantı ayarları
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// NewRedisClient yeni bir Redis bağlantısı oluşturur
func NewRedisClient(cfg RedisConfig) (*RedisClient, error) {
	// 1. Client oluştur
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port, // "localhost:6379"
		Password: cfg.Password,              // Şifre (yoksa boş string)
		DB:       cfg.DB,                    // Database numarası (varsayılan 0)
	})

	// 2. Bağlantıyı test et (5 saniye timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis bağlantı hatası: %w", err)
	}

	return &RedisClient{Client: client}, nil
}

// Close bağlantıyı kapatır
func (r *RedisClient) Close() error {
	return r.Client.Close()
}

// Health bağlantının sağlığını kontrol eder
func (r *RedisClient) Health(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}
