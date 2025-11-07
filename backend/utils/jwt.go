package utils

import (
	"fmt"
	"friendship/models"
	"friendship/models/dto"
	"friendship/repository"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type UserFinder func(email string) (*models.User, error)

func GenerateTokenPair(id uint, name, us, image string) (dto.AuthResponse, error) {
	secretKey := os.Getenv("SECRET_KEY_JWT")
	if len(secretKey) == 0 {
		return dto.AuthResponse{}, fmt.Errorf("пустой секретный ключ")
	}

	now := time.Now()

	// Access токен (20 минут жизни)
	accessClaims := jwt.MapClaims{
		"id":       id,
		"Us":       us,
		"Username": name,
		"Image":    image,
		"exp":      now.Add(20 * time.Minute).Unix(),
		"iat":      now.Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessString, err := accessToken.SignedString([]byte(secretKey))
	if err != nil {
		return dto.AuthResponse{}, err
	}

	// Refresh токен (30 дней жизни)
	refreshClaims := jwt.MapClaims{
		"id":       id,
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
		return dto.AuthResponse{}, err
	}

	return dto.AuthResponse{
		AccessToken:  accessString,
		RefreshToken: refreshString,
	}, nil
}

func RefreshTokens(refreshTokenString string, findUser dto.UserDto, rep repository.PostgresRepository, secretKey string) (dto.AuthResponse, error) {
	if len(secretKey) == 0 {
		return dto.AuthResponse{}, fmt.Errorf("пустой секретный ключ")
	}

	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка типа подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неподдерживаемый метод подписи: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return dto.AuthResponse{}, fmt.Errorf("невалидный refresh токен: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return dto.AuthResponse{}, fmt.Errorf("невалидный refresh токен")
	}

	// Правильное извлечение типа токена
	if typ, ok := claims["typ"].(string); !ok || typ != "refresh" {
		return dto.AuthResponse{}, fmt.Errorf("токен не является refresh токеном")
	}

	// Правильное извлечение значений из claims
	id, okEmail := claims["id"].(uint)
	if !okEmail {
		return dto.AuthResponse{}, fmt.Errorf("неверный формат Email в claims")
	}

	user, err := new(models.User).FindUserByID(id, rep)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	return GenerateTokenPair(id, user.Username, user.Us, user.Image)
}

func ParseJWT(tokenString string) (*uint, error) {
	secretKey := os.Getenv("SECRET_KEY_JWT")
	if secretKey == "" {
		return nil, fmt.Errorf("секретный ключ JWT не найден в переменных окружения")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка типа подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неподдерживаемый метод подписи: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("некорректный токен: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("некорректные claims")
	}

	// Правильное извлечение email
	id, ok := claims["id"].(uint)
	if !ok {
		return nil, fmt.Errorf("email не найден в токене")
	}

	return &id, nil
}
