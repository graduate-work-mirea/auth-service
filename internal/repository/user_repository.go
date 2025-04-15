package repository

import (
	"log"

	"github.com/diplom/auth-service/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// UserRepository provides access to the user storage
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(email, passwordHash string) (uuid.UUID, error) {
	var userID uuid.UUID
	query := `
		INSERT INTO users (email, password_hash) 
		VALUES ($1, $2) 
		RETURNING id
	`

	err := r.db.QueryRow(query, email, passwordHash).Scan(&userID)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return uuid.Nil, err
	}

	return userID, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, role, created_at FROM users WHERE id = $1`

	err := r.db.Get(&user, query, id)
	if err != nil {
		log.Printf("Error getting user by ID: %v", err)
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1`

	err := r.db.Get(&user, query, email)
	if err != nil {
		log.Printf("Error getting user by email: %v", err)
		return nil, err
	}

	return &user, nil
}

// UserExists checks if a user with the given email already exists
func (r *UserRepository) UserExists(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		log.Printf("Error checking if user exists: %v", err)
		return false, err
	}

	return exists, nil
}

// InitDatabase initializes the database schema
func InitDatabase(db *sqlx.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'user',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		log.Printf("Error initializing database: %v", err)
		return err
	}

	log.Println("Database schema initialized successfully")
	return nil
}

// Connect establishes a connection to the database
func Connect(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		return nil, err
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Printf("Error pinging database: %v", err)
		return nil, err
	}

	log.Println("Connected to the database successfully")
	return db, nil
}
