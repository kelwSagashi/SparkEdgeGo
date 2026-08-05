package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kelwSagashi/sparkedge-go/internal/domain"
)

var ErrInvalidToken = errors.New("invalid token")

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

type TokenClaims struct {
	UserID string          `json:"id"`
	Email  string          `json:"email"`
	Role   domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	if secret == "" {
		secret = "dev-secret"
	}
	return &TokenManager{secret: []byte(secret), ttl: ttl}
}

func (m *TokenManager) Sign(user domain.User) (string, error) {
	now := time.Now().UTC()
	claims := TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *TokenManager) Verify(rawToken string) (TokenClaims, error) {
	token, err := jwt.ParseWithClaims(rawToken, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil {
		return TokenClaims{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid || claims.UserID == "" {
		return TokenClaims{}, ErrInvalidToken
	}

	return *claims, nil
}
