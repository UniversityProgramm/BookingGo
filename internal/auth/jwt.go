package auth

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/enum"
	"BookingGo/pkg/logger"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

type Claims struct {
	UserID int       `json:"user_id"`
	Email  string    `json:"email"`
	Role   enum.Role `json:"role"`
	jwt.RegisteredClaims
}

func InitAuth() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		logger.Fatal("[jwt] JWT_SECRET.env is not set")
	}

	jwtSecret = []byte(secret)
}

func GenerateToken(userID int, email string, role enum.Role) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrSigningMethodNotSupported
		}
		return jwtSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrExpiredToken
		}
		return nil, domain.ErrInvalidToken
	}

	if !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}
