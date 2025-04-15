package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/diplom/auth-service/internal/models"
	"github.com/diplom/auth-service/internal/repository"
	"github.com/diplom/auth-service/internal/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication related operations
type AuthHandler struct {
	userRepo *repository.UserRepository
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(userRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

// RegisterHandler handles user registration
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var req models.UserRegisterRequest

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Check if user already exists
	exists, err := h.userRepo.UserExists(req.Email)
	if err != nil {
		log.Printf("Error checking if user exists: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Internal server error"})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "User with this email already exists"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to process password"})
		return
	}

	// Create the user
	userID, err := h.userRepo.CreateUser(req.Email, string(hashedPassword))
	if err != nil {
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create user"})
		return
	}

	// Get full user details
	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		log.Printf("Error fetching created user: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "User created but failed to fetch details"})
		return
	}

	// Generate tokens
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Role)
	if err != nil {
		log.Printf("Error generating tokens: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate authentication tokens"})
		return
	}

	// Return success response with enhanced data
	c.JSON(http.StatusCreated, models.UserRegisterResponse{
		UserID:       user.ID.String(),
		Email:        user.Email,
		Role:         user.Role,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		CreatedAt:    user.CreatedAt,
	})
}

// LoginHandler handles user login
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req models.UserLoginRequest

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Get user by email
	user, err := h.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("User not found: %v", err)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid email or password"})
		return
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		log.Printf("Invalid password: %v", err)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid email or password"})
		return
	}

	// Generate tokens
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Role)
	if err != nil {
		log.Printf("Error generating tokens: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate authentication tokens"})
		return
	}

	// Record current time as last login
	lastLoginAt := time.Now()

	// Return success response with enhanced data
	c.JSON(http.StatusOK, models.UserLoginResponse{
		UserID:       user.ID.String(),
		Email:        user.Email,
		Role:         user.Role,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		LastLoginAt:  lastLoginAt,
	})
}

// SetupRoutes sets up the authentication routes
func SetupRoutes(router *gin.Engine, authHandler *AuthHandler) {
	// Group auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.RegisterHandler)
		auth.POST("/login", authHandler.LoginHandler)
	}
}
