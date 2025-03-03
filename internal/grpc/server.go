package grpc

import (
	"context"
	"log"

	pb "github.com/diplom/auth-service/proto"
	"github.com/diplom/auth-service/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server is the gRPC server implementation
type Server struct {
	pb.UnimplementedAuthServiceServer
}

// NewServer creates a new gRPC server instance
func NewServer() *Server {
	return &Server{}
}

// ValidateToken validates a JWT token
func (s *Server) ValidateToken(ctx context.Context, req *pb.TokenRequest) (*pb.TokenResponse, error) {
	// Check if token is empty
	if req.Token == "" {
		log.Println("Empty token received")
		return nil, status.Error(codes.InvalidArgument, "token cannot be empty")
	}

	// Parse and validate the token
	userID, role, err := utils.ParseToken(req.Token)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		return &pb.TokenResponse{
			Valid:  false,
			UserId: "",
			Role:   "",
		}, nil
	}

	// Return successful response
	return &pb.TokenResponse{
		Valid:  true,
		UserId: userID.String(),
		Role:   role,
	}, nil
}

// RegisterGRPCServer registers the gRPC server with a grpc.Server instance
func RegisterGRPCServer(grpcServer *grpc.Server) {
	pb.RegisterAuthServiceServer(grpcServer, NewServer())
}
