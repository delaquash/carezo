package utils

import (
	"fmt"
	"time"

	"github.com/delaquash/carezo/configs"
	"github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}


// jwt auth token is sent with every API request for identification

func GenerateAccessToken(userID, email, role string, cfg *configs.Config) (string, error) {
	// Set token expiration time
	expirationTime := time.Now().Add(time.Duration(cfg.JWTExpirationHours) * time.Hour)

	// data inside the token

	claims := &JWTClaims {
		UserID : userID,
		Email: email,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Issuer: cfg.AppName,
		},
	}

	// creating token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// signing the token with secret key\
	tokenString, err  := token.SignedString([]byte(cfg.JWTSecret))

	if err != nil {
		return "", fmt.Errorf("Failed to sign token: %w", err)
	}
	return tokenString, nil
}

func GenerateRefreshToken(userID string, cfg *configs.Config) (string, error) {
	// Refresh token lives longer (e.g., 30 days)
	expirationTime := time.Now().Add(cfg.RefreshTokenExpiration)

	claims := &JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.AppName,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}


// ValidateToken checks if a JWT token is valid and return the claims
func ValidateToken(tokenString string, cfg *configs.Config) (*JWTClaims, error) {
	// pars the token\
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %w", token.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to parse token: %w", err)
	}
	// Extract claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("Invalid token")
	}

	return claims, nil
}