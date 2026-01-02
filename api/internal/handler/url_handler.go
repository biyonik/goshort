package handler

import (
	"net/http"
	"time"

	"github.com/biyonik/goshort/internal/domain"
	"github.com/biyonik/goshort/internal/repository"
	"github.com/biyonik/goshort/internal/shortener"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	v "github.com/biyonik/go-fluent-validator"
)

type URLHandler struct {
	repo    *repository.URLRepository
	encoder *shortener.Base62Encoder
	baseURL string
}

func NewURLHandler(repo *repository.URLRepository, encoder *shortener.Base62Encoder, baseURL string) *URLHandler {
	return &URLHandler{
		repo:    repo,
		encoder: encoder,
		baseURL: baseURL,
	}
}

func (h *URLHandler) Create(c *gin.Context) {
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
	})
	result := schema.Validate(data)

	// 3. Hata kontrolü
	if result.HasErrors() {
		c.JSON(http.StatusBadRequest, gin.H{"errors": result.Errors()})
		return
	}

	// 4. Short code oluştur
	counter, err := h.repo.GetNextCounter(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate short code"})
		return
	}

	shuffled := h.encoder.Shuffle(counter)
	code := h.encoder.Encode(shuffled)

	// 5. URL entity oluştur
	validData := result.ValidData()

	var expiration *time.Time
	if exp, ok := validData["url_expiration"]; ok && exp != nil {
		t := exp.(time.Time)
		expiration = &t
	}

	url := &domain.URL{
		ID:            uuid.Must(uuid.NewV7()).String(),
		LongURL:       validData["long_url"].(string),
		Code:          code,
		UrlExpiration: expiration,
		ClickCount:    0,
		UserID:        "", // TODO: Auth'dan alınacak
		CreatedAt:     time.Now(),
	}

	// 6. Database'e kaydet
	if err := h.repo.Create(c.Request.Context(), url); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL"})
		return
	}

	// 7. Response dön
	c.JSON(http.StatusCreated, domain.CreateURLResponse{
		ID:            url.ID,
		Code:          url.Code,
		ShortURL:      h.baseURL + "/" + url.Code,
		LongURL:       url.LongURL,
		UrlExpiration: url.UrlExpiration,
		CreatedAt:     url.CreatedAt,
	})
}

func (h *URLHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	// Database'den URL'i bul
	url, err := h.repo.FindByShortCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	// Expiration kontrolü
	if url.UrlExpiration != nil && time.Now().After(*url.UrlExpiration) {
		c.JSON(http.StatusGone, gin.H{"error": "URL has expired"})
		return
	}

	// TODO: Click count artır (async)

	// Redirect
	c.Redirect(http.StatusFound, url.LongURL)
}
