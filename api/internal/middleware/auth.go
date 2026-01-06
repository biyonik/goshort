package middleware

import (
	"net/http"
	"strings"

	"github.com/biyonik/goshort/internal/domain"
	"github.com/biyonik/goshort/internal/service"
	"github.com/gin-gonic/gin"
)

// Context key'leri
const (
	ContextUserID = "user_id"
	ContextEmail  = "email"
	ContextTier   = "tier"
)

// AuthMiddleware JWT doğrulama middleware'i
type AuthMiddleware struct {
	authService *service.AuthService
}

// NewAuthMiddleware yeni auth middleware oluşturur
func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// Required token zorunlu - yoksa 401
func (m *AuthMiddleware) Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := m.extractAndValidateToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Claims'i context'e ekle
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextEmail, claims.Email)
		c.Set(ContextTier, claims.Tier)

		c.Next()
	}
}

// Optional token opsiyonel - yoksa da devam et
func (m *AuthMiddleware) Optional() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := m.extractAndValidateToken(c)
		if err == nil {
			// Token varsa ve geçerliyse context'e ekle
			c.Set(ContextUserID, claims.UserID)
			c.Set(ContextEmail, claims.Email)
			c.Set(ContextTier, claims.Tier)
		}
		// Token yoksa veya geçersizse de devam et

		c.Next()
	}
}

// extractAndValidateToken header'dan token'ı çıkarır ve doğrular
func (m *AuthMiddleware) extractAndValidateToken(c *gin.Context) (*service.JWTClaims, error) {
	// Header: Authorization: Bearer <token>
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, service.ErrInvalidToken
	}

	// "Bearer " prefix'ini kaldır
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, service.ErrInvalidToken
	}

	tokenString := parts[1]
	return m.authService.ValidateToken(tokenString)
}

// GetUserID context'ten user ID alır
func GetUserID(c *gin.Context) string {
	if val, exists := c.Get(ContextUserID); exists {
		return val.(string)
	}
	return ""
}

// GetUserTier context'ten tier alır
func GetUserTier(c *gin.Context) string {
	if val, exists := c.Get(ContextTier); exists {
		return string(val.(domain.UserTier))
	}
	return ""
}
