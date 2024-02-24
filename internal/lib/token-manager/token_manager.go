package token_manager

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type TokenManager struct {
	secret string
}

func New(secret string) *TokenManager {
	return &TokenManager{secret: secret}
}

func (m *TokenManager) NewJWT(userId string, ttl time.Duration) (string, error) {
	const op = "lib.token-manager.token_manager.Parse"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(ttl).Unix(),
		Subject:   userId,
	})

	completeToken, err := token.SignedString([]byte(m.secret))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return completeToken, nil
}

func (m *TokenManager) Parse(accessToken string) (string, error) {
	const op = "lib.token-manager.token_manager.Parse"

	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(m.secret), nil
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("%s: can't get user claims from token", op)
	}

	return claims["sub"].(string), nil
}

func (m *TokenManager) NewRefreshToken() (string, error) {
	const op = "lib.token-manager.token_manager.NewRefreshToken"

	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return fmt.Sprintf("%x", b), nil
}
