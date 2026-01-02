package main

import (
	"log"

	"github.com/biyonik/goshort/internal/config"
	"github.com/biyonik/goshort/internal/handler"
	"github.com/biyonik/goshort/internal/repository"
	"github.com/biyonik/goshort/internal/shortener"
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

	// 3. Dependencies oluştur
	urlRepo := repository.NewURLRepository(db.Pool)
	encoder := shortener.NewBase62Encoder(cfg.URL.ShortCodeLength)

	// 4. Handler oluştur
	urlHandler := handler.NewURLHandler(urlRepo, encoder, cfg.Server.BaseURL)

	// 5. Router oluştur
	r := gin.Default()

	// 6. Routes
	r.GET("/health", healthCheck)

	// API routes
	api := r.Group("/api")
	{
		api.POST("/urls", urlHandler.Create)
	}

	// Redirect route (en sonda olmalı)
	r.GET("/:code", urlHandler.Redirect)

	// 7. Server başlat
	log.Printf("Server starting on port %s...", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "OK", "message": "Çalışıyor kankaaa"})
}
