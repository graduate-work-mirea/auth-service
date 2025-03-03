package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Role         string    `db:"role" json:"role"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// UserRegisterRequest is the request structure for user registration
type UserRegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// UserRegisterResponse is the response structure for user registration
type UserRegisterResponse struct {
	UserID string `json:"user_id"`
}

// UserLoginRequest is the request structure for user login
type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserLoginResponse is the response structure for user login
type UserLoginResponse struct {
	AccessToken string `json:"access_token"`
}

// ErrorResponse is a generic error response format
type ErrorResponse struct {
	Error string `json:"error"`
}
