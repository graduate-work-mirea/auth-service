package utils

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenClaims represents the JWT claims structure
type TokenClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair holds both access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(userID uuid.UUID, role string) (TokenPair, error) {
	// Generate access token
	accessToken, accessExpiresAt, err := generateAccessToken(userID, role)
	if err != nil {
		return TokenPair{}, err
	}

	// Generate refresh token
	refreshToken, err := generateRefreshToken(userID, role)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiresAt.Unix(),
	}, nil
}

// GenerateToken is kept for backward compatibility
func GenerateToken(userID uuid.UUID, role string) (string, error) {
	token, _, err := generateAccessToken(userID, role)
	return token, err
}

// generateAccessToken generates a new JWT access token for a user
func generateAccessToken(userID uuid.UUID, role string) (string, time.Time, error) {
	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", time.Time{}, errors.New("JWT_SECRET environment variable not set")
	}

	// Set expiration time (1 hour)
	expirationTime := time.Now().Add(1 * time.Hour)

	// Create the JWT claims
	claims := &TokenClaims{
		UserID: userID.String(),
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("Error signing token: %v", err)
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}

// generateRefreshToken generates a new JWT refresh token for a user
func generateRefreshToken(userID uuid.UUID, role string) (string, error) {
	// Get refresh token secret from environment
	refreshSecret := os.Getenv("REFRESH_TOKEN_SECRET")
	if refreshSecret == "" {
		// Fallback to JWT_SECRET if REFRESH_TOKEN_SECRET is not set
		refreshSecret = os.Getenv("JWT_SECRET")
		if refreshSecret == "" {
			return "", errors.New("neither REFRESH_TOKEN_SECRET nor JWT_SECRET environment variables are set")
		}
	}

	// Set expiration time (30 days)
	expirationTime := time.Now().Add(30 * 24 * time.Hour)

	// Create the JWT claims
	claims := &TokenClaims{
		UserID: userID.String(),
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(refreshSecret))
	if err != nil {
		log.Printf("Error signing refresh token: %v", err)
		return "", err
	}

	return tokenString, nil
}

// ParseToken parses and validates a JWT token
func ParseToken(tokenString string) (uuid.UUID, string, error) {
	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return uuid.Nil, "", errors.New("JWT_SECRET environment variable not set")
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return uuid.Nil, "", err
	}

	// Check if the token is valid
	if !token.Valid {
		return uuid.Nil, "", errors.New("invalid token")
	}

	// Extract the claims
	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return uuid.Nil, "", errors.New("invalid token claims")
	}

	// Parse user ID
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, "", errors.New("invalid user ID in token")
	}

	return userID, claims.Role, nil
}
