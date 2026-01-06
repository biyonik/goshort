package service

import (
	"context"
	"errors"
	"time"

	"github.com/biyonik/goshort/internal/domain"
	"github.com/biyonik/goshort/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Auth hataları
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid token")
)

// JWTClaims JWT payload içeriği
type JWTClaims struct {
	UserID string          `json:"user_id"`
	Email  string          `json:"email"`
	Tier   domain.UserTier `json:"tier"`
	jwt.RegisteredClaims
}

// AuthService authentication işlemleri
type AuthService struct {
	userRepo     *repository.UserRepository
	jwtSecret    []byte
	jwtExpiresIn time.Duration
}

// NewAuthService yeni auth service oluşturur
func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, jwtExpiresIn time.Duration) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		jwtSecret:    []byte(jwtSecret),
		jwtExpiresIn: jwtExpiresIn,
	}
}

// Register yeni kullanıcı kaydı
func (s *AuthService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error) {
	// 1. Password hash'le
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 2. User oluştur
	user := &domain.User{
		ID:           uuid.Must(uuid.NewV7()).String(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Tier:         domain.TierFree, // Yeni kullanıcılar free tier
		CreatedAt:    time.Now(),
	}

	// 3. Kaydet
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 4. Token üret ve dön
	return s.generateAuthResponse(user)
}

// Login kullanıcı girişi
func (s *AuthService) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	// 1. Kullanıcıyı bul
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	// 2. Password doğrula
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 3. Token üret ve dön
	return s.generateAuthResponse(user)
}

// ValidateToken token'ı doğrular ve claims döner
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// generateAuthResponse token üretir ve response oluşturur
func (s *AuthService) generateAuthResponse(user *domain.User) (*domain.AuthResponse, error) {
	expiresAt := time.Now().Add(s.jwtExpiresIn)

	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Tier:   user.Tier,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt.Unix(),
		User: domain.UserInfo{
			ID:    user.ID,
			Email: user.Email,
			Tier:  user.Tier,
		},
	}, nil
}
