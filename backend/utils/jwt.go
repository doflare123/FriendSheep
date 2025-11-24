package utils

import (
	"fmt"
	"friendship/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type UserFinder func(email string) (*models.User, error)

func GenerateTokenPair(id int, email, name, us, image string) (TokenPair, error) {
	secretKey := os.Getenv("SECRET_KEY_JWT")
	if len(secretKey) == 0 {
		return TokenPair{}, fmt.Errorf("пустой секретный ключ")
	}

	now := time.Now()

	// Access токен (20 минут жизни)
	accessClaims := jwt.MapClaims{
		"Id":       id,
		"Email":    email,
		"Us":       us,
		"Username": name,
		"Image":    image,
		"exp":      now.Add(20 * time.Minute).Unix(),
		"iat":      now.Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessString, err := accessToken.SignedString([]byte(secretKey))
	if err != nil {
		return TokenPair{}, err
	}

	// Refresh токен (30 дней жизни)
	refreshClaims := jwt.MapClaims{
		"Id":       id,
		"Email":    email,
		"Us":       us,
		"Username": name,
		"Image":    image, // Добавляем Image и в refresh токен
		"exp":      now.Add(30 * 24 * time.Hour).Unix(),
		"iat":      now.Unix(),
		"typ":      "refresh",
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

func RefreshTokens(refreshTokenString string, findUser UserFinder) (TokenPair, error) {
	secretKey := os.Getenv("SECRET_KEY_JWT")
	if len(secretKey) == 0 {
		return TokenPair{}, fmt.Errorf("пустой секретный ключ")
	}

	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка типа подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неподдерживаемый метод подписи: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return TokenPair{}, fmt.Errorf("невалидный refresh токен: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return TokenPair{}, fmt.Errorf("невалидный refresh токен")
	}

	// Правильное извлечение типа токена
	if typ, ok := claims["typ"].(string); !ok || typ != "refresh" {
		return TokenPair{}, fmt.Errorf("токен не является refresh токеном")
	}

	// Правильное извлечение значений из claims
	email, okEmail := claims["Email"].(string)
	if !okEmail {
		return TokenPair{}, fmt.Errorf("неверный формат Email в claims")
	}

	id, okEmail := claims["Id"].(int)
	if !okEmail {
		return TokenPair{}, fmt.Errorf("неверный формат Email в claims")
	}

	user, err := findUser(email)
	if err != nil {
		return TokenPair{}, err
	}

	return GenerateTokenPair(id, email, user.Name, user.Us, user.Image)
}

func ParseJWT(tokenString string) (string, error) {
	secretKey := os.Getenv("SECRET_KEY_JWT")
	if secretKey == "" {
		return "", fmt.Errorf("секретный ключ JWT не найден в переменных окружения")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка типа подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неподдерживаемый метод подписи: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("некорректный токен: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("некорректные claims")
	}

	// Правильное извлечение email
	email, ok := claims["Email"].(string)
	if !ok {
		return "", fmt.Errorf("email не найден в токене")
	}

	return email, nil
}
