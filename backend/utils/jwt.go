package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// Создаёт access token и refresh token
func GenerateTokenPair(email, us string) (TokenPair, error) {
	secretKey := os.Getenv("SECRET_KEY_JWT")
	if len(secretKey) == 0 {
		return TokenPair{}, fmt.Errorf("пустой секретный ключ")
	}

	now := time.Now()

	// Access токен (6 часов жизни)
	accessClaims := jwt.MapClaims{
		"Email": email,
		"Us":    us,
		"exp":   now.Add(6 * time.Hour).Unix(),
		"iat":   now.Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessString, err := accessToken.SignedString([]byte(secretKey))
	if err != nil {
		return TokenPair{}, err
	}

	refreshClaims := jwt.MapClaims{
		"Email": email,
		"Us":    us,
		"exp":   now.Add(7 * 24 * time.Hour).Unix(),
		"iat":   now.Unix(),
		"typ":   "refresh",
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshString, err := refreshToken.SignedString([]byte(secretKey))
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessString,
		RefreshToken: refreshString,
	}, nil
}

func RefreshTokens(refreshTokenString string) (TokenPair, error) {
	secretKey := os.Getenv("SECRET_KEY_JWT")
	if len(secretKey) == 0 {
		return TokenPair{}, fmt.Errorf("пустой секретный ключ")
	}

	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return TokenPair{}, fmt.Errorf("невалидный refresh токен: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return TokenPair{}, fmt.Errorf("невалидный refresh токен")
	}

	if typ, ok := claims["typ"].(string); !ok || typ != "refresh" {
		return TokenPair{}, fmt.Errorf("токен не является refresh токеном")
	}

	email, okEmail := claims["Email"].(string)
	us, okUs := claims["Us"].(string)
	if !okEmail || !okUs {
		return TokenPair{}, fmt.Errorf("недостаточно данных в refresh токене")
	}

	return GenerateTokenPair(email, us)
}
