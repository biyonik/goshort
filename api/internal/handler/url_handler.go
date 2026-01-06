package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/biyonik/goshort/internal/domain"
	"github.com/biyonik/goshort/internal/repository"
	"github.com/biyonik/goshort/internal/shortener"
	"github.com/biyonik/goshort/internal/worker"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	v "github.com/biyonik/go-fluent-validator"
)

type URLHandler struct {
	repo           *repository.URLRepository
	cache          *repository.CacheRepository
	encoder        *shortener.Base62Encoder
	clickProcessor *worker.ClickProcessor
	baseURL        string
}

func NewURLHandler(
	repo *repository.URLRepository,
	cache *repository.CacheRepository,
	encoder *shortener.Base62Encoder,
	clickProcessor *worker.ClickProcessor,
	baseURL string,
) *URLHandler {
	return &URLHandler{
		repo:           repo,
		cache:          cache,
		encoder:        encoder,
		clickProcessor: clickProcessor,
		baseURL:        baseURL,
	}
}

func (h *URLHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Parse
	var data map[string]any
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// 2. Validate
	schema := v.Make().Shape(map[string]v.Type{
		"long_url":       v.String().URL().Required().Label("Long URL"),
		"url_expiration": v.Date().After(time.Now()).Label("Expiration"),
		"custom_code":    v.String().Min(3).Max(20).Label("Custom Code"),
	})
	result := schema.Validate(data)

	// 3. Hata kontrolü
	if result.HasErrors() {
		c.JSON(http.StatusBadRequest, gin.H{"errors": result.Errors()})
		return
	}

	validData := result.ValidData()

	// 4. User bilgilerini al
	userID := ""
	tier := domain.TierFree
	if id, exists := c.Get("user_id"); exists {
		userID = id.(string)
	}
	if t, exists := c.Get("tier"); exists {
		tier = t.(domain.UserTier)
	}

	// 5. Short code belirle
	var code string
	customCode, hasCustomCode := validData["custom_code"].(string)

	if hasCustomCode && customCode != "" {
		// Custom code isteniyor

		// 5a. Pro tier kontrolü
		if tier != domain.TierPro {
			c.JSON(http.StatusForbidden, gin.H{
				"error":       "Custom short codes are a Pro feature",
				"upgrade_tip": "Upgrade to Pro to use custom codes!",
			})
			return
		}

		// 5b. Format kontrolü (alfanumerik + tire)
		if !isValidCustomCode(customCode) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Custom code can only contain letters, numbers, and hyphens",
			})
			return
		}

		// 5c. Reserved words kontrolü
		if isReservedCode(customCode) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "This code is reserved and cannot be used",
			})
			return
		}

		// 5d. Uniqueness kontrolü
		exists, err := h.repo.ExistsByCode(ctx, customCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check code availability"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{
				"error": "This custom code is already taken",
			})
			return
		}

		code = customCode
	} else {
		// Auto-generate code
		counter, err := h.repo.GetNextCounter(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate short code"})
			return
		}
		shuffled := h.encoder.Shuffle(counter)
		code = h.encoder.Encode(shuffled)
	}

	// 6. Expiration
	var expiration *time.Time
	if exp, ok := validData["url_expiration"]; ok && exp != nil {
		t := exp.(time.Time)
		expiration = &t
	}

	// 7. URL entity oluştur
	url := &domain.URL{
		ID:            uuid.Must(uuid.NewV7()).String(),
		LongURL:       validData["long_url"].(string),
		Code:          code,
		UrlExpiration: expiration,
		ClickCount:    0,
		UserID:        userID,
		CreatedAt:     time.Now(),
	}

	// 8. Database'e kaydet
	if err := h.repo.Create(ctx, url); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL"})
		return
	}

	// 9. Cache'e yaz (Write-Through)
	if err := h.cache.SetURL(ctx, url); err != nil {
		// Cache yazma hatası kritik değil, devam et
	}

	// 10. Response dön
	c.JSON(http.StatusCreated, domain.CreateURLResponse{
		ID:            url.ID,
		Code:          url.Code,
		ShortURL:      h.baseURL + "/" + url.Code,
		LongURL:       url.LongURL,
		UrlExpiration: url.UrlExpiration,
		CreatedAt:     url.CreatedAt,
	})
}

// isValidCustomCode custom code formatını kontrol eder
// Sadece alfanumerik karakterler ve tire kabul edilir
func isValidCustomCode(code string) bool {
	for _, r := range code {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}
	return true
}

// isReservedCode reserved code'ları kontrol eder
func isReservedCode(code string) bool {
	reserved := map[string]bool{
		"api":       true,
		"admin":     true,
		"auth":      true,
		"health":    true,
		"login":     true,
		"register":  true,
		"logout":    true,
		"dashboard": true,
		"settings":  true,
		"profile":   true,
		"help":      true,
		"about":     true,
		"terms":     true,
		"privacy":   true,
	}
	return reserved[code]
}

func (h *URLHandler) Redirect(c *gin.Context) {
	code := c.Param("code")
	ctx := c.Request.Context()

	var url *domain.URL
	var err error

	// 1. Önce cache'e bak
	url, err = h.cache.GetURL(ctx, code)

	if errors.Is(err, repository.ErrCacheMiss) {
		// 2. Cache MISS - Database'den al
		url, err = h.repo.FindByShortCode(ctx, code)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}

		// 3. Cache'e yaz (sonraki istekler için)
		_ = h.cache.SetURL(ctx, url)

	} else if err != nil {
		// Redis hatası - DB'ye fallback
		url, err = h.repo.FindByShortCode(ctx, code)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
	}

	// 4. Expiration kontrolü
	if url.UrlExpiration != nil && time.Now().After(*url.UrlExpiration) {
		// Expired - cache'ten sil
		_ = h.cache.DeleteURL(ctx, code)
		c.JSON(http.StatusGone, gin.H{"error": "URL has expired"})
		return
	}

	// 5. Click event'i async gönder (redirect'i yavaşlatmaz)
	h.clickProcessor.Push(domain.ClickEvent{
		URLCode:   code,
		Timestamp: time.Now(),
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Referer:   c.Request.Referer(),
	})

	// 6. Redirect
	c.Redirect(http.StatusFound, url.LongURL)
}
