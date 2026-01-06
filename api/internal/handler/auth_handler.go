package handler

import (
	"errors"
	"net/http"

	"github.com/biyonik/goshort/internal/domain"
	"github.com/biyonik/goshort/internal/repository"
	"github.com/biyonik/goshort/internal/service"
	"github.com/gin-gonic/gin"

	v "github.com/biyonik/go-fluent-validator"
)

// AuthHandler authentication endpoint'leri
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler yeni auth handler oluşturur
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	// 1. Parse
	var data map[string]any
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// 2. Validate
	schema := v.Make().Shape(map[string]v.Type{
		"email":    v.String().Email().Required().Label("Email"),
		"password": v.String().Min(6).Max(100).Required().Label("Password"),
	})
	result := schema.Validate(data)

	// 3. Hata kontrolü
	if result.HasErrors() {
		c.JSON(http.StatusBadRequest, gin.H{"errors": result.Errors()})
		return
	}

	// 4. Register
	validData := result.ValidData()
	req := domain.RegisterRequest{
		Email:    validData["email"].(string),
		Password: validData["password"].(string),
	}

	response, err := h.authService.Register(c.Request.Context(), req)
	if errors.Is(err, repository.ErrEmailExists) {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		return
	}

	// 5. Response
	c.JSON(http.StatusCreated, response)
}

// Login POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	// 1. Parse
	var data map[string]any
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// 2. Validate
	schema := v.Make().Shape(map[string]v.Type{
		"email":    v.String().Email().Required().Label("Email"),
		"password": v.String().Required().Label("Password"),
	})
	result := schema.Validate(data)

	// 3. Hata kontrolü
	if result.HasErrors() {
		c.JSON(http.StatusBadRequest, gin.H{"errors": result.Errors()})
		return
	}

	// 4. Login
	validData := result.ValidData()
	req := domain.LoginRequest{
		Email:    validData["email"].(string),
		Password: validData["password"].(string),
	}

	response, err := h.authService.Login(c.Request.Context(), req)
	if errors.Is(err, service.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	// 5. Response
	c.JSON(http.StatusOK, response)
}

// Me GET /api/auth/me - Mevcut kullanıcı bilgisi
func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")
	tier, _ := c.Get("tier")

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    userID,
			"email": email,
			"tier":  tier,
		},
	})
}
