package services

import (
	"context"
	"fmt"
	"log"
	"time"

	grpcclient "github.com/Kisanlink/farmers-module/grpc_client"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc"
)

func InitializeGrpcClient(token string, retries int) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	var err error

	for i := 0; i < retries; i++ {
		conn, err = grpcclient.GrpcClient(token)
		if err == nil {
			return conn, nil
		}
		log.Printf("Failed to initialize gRPC client (attempt %d): %v", i+1, err)
		time.Sleep(10 * time.Second)
	}

	return nil, fmt.Errorf("failed to initialize gRPC client after %d retries: %v", retries, err)
}

func CreateUserClient(req models.FarmerSignupRequest, token string) (*pb.CreateUserResponse, error) {
	// Initialize gRPC connection with retry mechanism
	conn, err := InitializeGrpcClient(token, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to establish gRPC connection: %v", err)
	}
	defer conn.Close()

	// Create User Service Client
	userClient := pb.NewUserServiceClient(conn)

	// Prepare gRPC request
	userRequest := &pb.CreateUserRequest{
		Username:      req.Name,
		MobileNumber:  req.MobileNumber,
		AadhaarNumber: req.AadhaarNumber,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call gRPC service
	response, err := userClient.CreateUser(ctx, userRequest)
	if err != nil {
		log.Printf("Failed to create user via gRPC: %v", err)
		return nil, err
	}

	log.Printf("Successfully created user: %v", response)
	return response, nil
}
