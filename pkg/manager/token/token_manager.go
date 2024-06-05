package token

import (
	"fmt"
	"math/rand"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type TokenManager interface {
	CreateTokensPair(userId string, ttl time.Duration) (*Tokens, error)
	Parse(accessToken string) (string, error)
}

type Manager struct {
	secret string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Time
}

func NewManager(secret string) *Manager {
	return &Manager{secret: secret}
}

func (m *Manager) CreateTokensPair(userId string, ttl time.Duration) (*Tokens, error) {
	const op = "lib.token-manager.token_manager.createTokensPair"

	accessToken, err := m.newJWT(userId, ttl)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := m.newRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    time.Now().Add(ttl),
	}, nil
}

func (m *Manager) Parse(accessToken string) (string, error) {
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

func (m *Manager) newJWT(userId string, ttl time.Duration) (string, error) {
	const op = "lib.token-manager.token_manager.Parse"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		Subject:   userId,
	})

	completeToken, err := token.SignedString([]byte(m.secret))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return completeToken, nil
}

func (m *Manager) newRefreshToken() (string, error) {
	const op = "lib.token-manager.token_manager.newRefreshToken"

	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return fmt.Sprintf("%x", b), nil
}
