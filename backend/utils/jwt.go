package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTUtils struct {
	secretKey            string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewJWTUtils(secretKey string) *JWTUtils {
	return &JWTUtils{
		secretKey:            secretKey,
		accessTokenDuration:  20 * time.Minute,
		refreshTokenDuration: 30 * 24 * time.Hour,
	}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Claims struct {
	UserID   uint   `json:"id"`
	Username string `json:"username"`
	Us       string `json:"us"`
	Image    string `json:"image"`
	Type     string `json:"typ,omitempty"`
	jwt.RegisteredClaims
}

func (j *JWTUtils) GenerateTokenPair(id uint, username, us, image string) (TokenPair, error) {
	if j.secretKey == "" {
		return TokenPair{}, fmt.Errorf("пустой секретный ключ")
	}

	now := time.Now()

	// Access токен
	accessClaims := Claims{
		UserID:   id,
		Username: username,
		Us:       us,
		Image:    image,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessString, err := accessToken.SignedString([]byte(j.secretKey))
	if err != nil {
		return TokenPair{}, fmt.Errorf("ошибка создания access токена: %w", err)
	}

	// Refresh токен
	refreshClaims := Claims{
		UserID:   id,
		Username: username,
		Us:       us,
		Image:    image,
		Type:     "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshString, err := refreshToken.SignedString([]byte(j.secretKey))
	if err != nil {
		return TokenPair{}, fmt.Errorf("ошибка создания refresh токена: %w", err)
	}

	return TokenPair{
		AccessToken:  accessString,
		RefreshToken: refreshString,
	}, nil
}

func (j *JWTUtils) ParseAccessToken(tokenString string) (*Claims, error) {
	return j.parseToken(tokenString, false)
}

func (j *JWTUtils) ParseRefreshToken(tokenString string) (uint, error) {
	claims, err := j.parseToken(tokenString, true)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

func (j *JWTUtils) parseToken(tokenString string, isRefresh bool) (*Claims, error) {
	if j.secretKey == "" {
		return nil, fmt.Errorf("секретный ключ не установлен")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неподдерживаемый метод подписи: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("невалидный токен: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("невалидные claims")
	}

	if isRefresh && claims.Type != "refresh" {
		return nil, fmt.Errorf("токен не является refresh токеном")
	}
	if !isRefresh && claims.Type == "refresh" {
		return nil, fmt.Errorf("нельзя использовать refresh токен как access токен")
	}

	return claims, nil
}

func (j *JWTUtils) ValidateToken(tokenString string) error {
	_, err := j.ParseAccessToken(tokenString)
	return err
}
