package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTConfig struct {
	Secret            string
	ExpirationHours   int
	RefreshExpiration int
}

type JWTService struct {
	config *JWTConfig
}

func NewJWTService(config *JWTConfig) *JWTService {
	return &JWTService{config: config}
}

func (s *JWTService) GenerateToken(userID uuid.UUID) (string, string, error) {
	// Generate access token
	accessToken, err := s.generateAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshToken, err := s.generateRefreshToken(userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (s *JWTService) generateAccessToken(userID uuid.UUID) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(s.config.ExpirationHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Secret))
}

func (s *JWTService) generateRefreshToken(userID uuid.UUID) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(s.config.RefreshExpiration))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Secret))
}
