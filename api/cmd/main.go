package main

import (
	"context"
	"log"
	"time"

	"github.com/biyonik/goshort/internal/config"
	"github.com/biyonik/goshort/internal/domain"
	"github.com/biyonik/goshort/internal/handler"
	"github.com/biyonik/goshort/internal/middleware"
	"github.com/biyonik/goshort/internal/repository"
	"github.com/biyonik/goshort/internal/service"
	"github.com/biyonik/goshort/internal/shortener"
	"github.com/biyonik/goshort/internal/worker"
	"github.com/biyonik/goshort/pkg/cache"
	"github.com/biyonik/goshort/pkg/database"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Config yükle
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 2. Database bağlantısı
	db, err := database.NewPostgresDB(cfg.Postgres.DSN())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 3. Redis bağlantısı
	redisClient, err := cache.NewRedisClient(cache.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: "",
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	// 4. Repositories oluştur
	urlRepo := repository.NewURLRepository(db.Pool)
	userRepo := repository.NewUserRepository(db.Pool)
	cacheRepo := repository.NewCacheRepository(
		redisClient.Client,
		24*time.Hour, // Default cache TTL: 1 gün
	)
	clickRepo := repository.NewClickRepository(db.Pool)

	// 5. Services oluştur
	authService := service.NewAuthService(
		userRepo,
		cfg.JWT.Secret,
		time.Duration(cfg.JWT.ExpiresIn)*time.Hour,
	)

	// 6. Click processor oluştur ve başlat
	clickProcessor := worker.NewClickProcessor(
		10000,         // Buffer: 10.000 event
		100,           // Batch: 100 event'te bir yaz
		5*time.Second, // Flush: 5 saniyede bir yaz
		func(ctx context.Context, events []domain.ClickEvent) error {
			return clickRepo.SaveBatch(ctx, events)
		},
	)
	clickProcessor.Start(context.Background())
	defer clickProcessor.Stop()

	// 7. Encoder oluştur
	encoder := shortener.NewBase62Encoder(cfg.URL.ShortCodeLength)

	// 8. Middleware'ler oluştur
	rateLimiter := middleware.NewRateLimiter(redisClient.Client)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Rate limit config'leri
	createLimitTiered := middleware.TieredRateLimitConfig{
		Free: middleware.RateLimitConfig{
			Limit:  10, // Free: 10 istek/dakika
			Window: 1 * time.Minute,
		},
		Pro: middleware.RateLimitConfig{
			Limit:  100, // Pro: 100 istek/dakika
			Window: 1 * time.Minute,
		},
	}
	redirectLimit := middleware.RateLimitConfig{
		Limit:  1000,            // 1000 istek
		Window: 1 * time.Minute, // 1 dakikada
	}
	authLimit := middleware.RateLimitConfig{
		Limit:  5,               // 5 istek (brute-force koruması)
		Window: 1 * time.Minute, // 1 dakikada
	}

	// 9. Handlers oluştur
	urlHandler := handler.NewURLHandler(urlRepo, cacheRepo, encoder, clickProcessor, cfg.Server.BaseURL)
	authHandler := handler.NewAuthHandler(authService)

	// 10. Router oluştur
	r := gin.Default()

	// 11. Routes
	r.GET("/health", healthCheck)

	// Auth routes (rate limited - brute-force koruması)
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", rateLimiter.Middleware(authLimit), authHandler.Register)
		auth.POST("/login", rateLimiter.Middleware(authLimit), authHandler.Login)
		auth.GET("/me", authMiddleware.Required(), authHandler.Me)
	}

	// URL routes
	// Önemli: Önce auth (tier'ı context'e ekler), sonra rate limit (tier'a göre limit uygular)
	api := r.Group("/api")
	{
		// Sıra önemli: auth → tier context'e eklenir → tiered rate limit
		api.POST("/urls", authMiddleware.Optional(), rateLimiter.TieredMiddleware(createLimitTiered), urlHandler.Create)
	}

	// Redirect route (rate limited - herkese aynı)
	r.GET("/:code", rateLimiter.Middleware(redirectLimit), urlHandler.Redirect)

	// 12. Server başlat
	log.Printf("Server starting on port %s...", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "OK", "message": "Çalışıyor kankaaa"})
}
