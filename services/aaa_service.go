package services

import (
	"context"
	"fmt"
	"log"

	pb "github.com/Kisanlink/farmers-module/pb" // Import generated gRPC code
	"google.golang.org/grpc"
)

// AAAServiceInterface defines methods required for AAA service
type AAAServiceInterface interface {
	CreateUser(username string, password string, userRoleIds []string) (string, error)
}

// AAAService handles communication with the AAA gRPC service
type AAAService struct {
	client pb.UserServiceClient
}

// NewAAAService creates a new AAA service client
func NewAAAService(aaaServiceAddress string) *AAAService {
	conn, err := grpc.Dial(aaaServiceAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to AAA Service at %s: %v", aaaServiceAddress, err)
	}
	client := pb.NewUserServiceClient(conn)
	return &AAAService{client: client}
}

// CreateUser creates a new user via gRPC
func (s *AAAService) CreateUser(username string, password string, userRoleIds []string) (string, error) {
	req := &pb.CreateUserRequest{
		Username:    username,
		Password:    password,
		UserRoleIds: userRoleIds,
	}
	resp, err := s.client.CreateUser(context.Background(), req)
	if err != nil {
		return "", err
	}
	if resp == nil || resp.User == nil {
		return "", fmt.Errorf("received nil response from AAA service")
	}
	return resp.User.Id, nil
}
