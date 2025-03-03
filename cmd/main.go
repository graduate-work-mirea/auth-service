package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diplom/auth-service/internal/grpc"
	"github.com/diplom/auth-service/internal/handlers"
	"github.com/diplom/auth-service/internal/repository"
	"github.com/gin-gonic/gin"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Println("Starting Auth Service...")

	// Get environment variables
	postgresDSN := os.Getenv("POSTGRES_DSN")
	if postgresDSN == "" {
		log.Println("POSTGRES_DSN environment variable not set, using default")
		postgresDSN = "postgres://user:password@postgres:5432/auth_db?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Println("JWT_SECRET environment variable not set, using default (for development only)")
		jwtSecret = "your_secret_key_here"
		os.Setenv("JWT_SECRET", jwtSecret)
	}

	// Connect to the database
	db, err := repository.Connect(postgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize database schema
	err = repository.InitDatabase(db)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create repository and handler
	userRepo := repository.NewUserRepository(db)
	authHandler := handlers.NewAuthHandler(userRepo)

	// Setup Gin router
	router := gin.Default()
	handlers.SetupRoutes(router, authHandler)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start gRPC server in a goroutine
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen on port 50051: %v", err)
		}

		s := ggrpc.NewServer()
		grpc.RegisterGRPCServer(s)
		reflection.Register(s) // Enable reflection for debugging

		log.Println("Starting gRPC server on :50051")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server shutdown error: %v", err)
	}

	log.Println("Servers shutdown gracefully")
}
