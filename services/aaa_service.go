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
		Username:      *req.Name,
		MobileNumber:  req.MobileNumber,
		AadhaarNumber: *req.AadhaarNumber,
		Password:      "Default@123", // ✅ Add a default password if not provided
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


func GetUserByIdClient(ctx context.Context, userID string) (*pb.GetUserByIdResponse, error) {
	// Initialize gRPC connection with retry mechanism
	conn, err := InitializeGrpcClient("", 3) // Assuming no auth token is needed
	if err != nil {
		return nil, fmt.Errorf("failed to establish gRPC connection: %v", err)
	}
	defer conn.Close()

	// Create User Service Client
	userClient := pb.NewUserServiceClient(conn)

	// Prepare gRPC request
	userReq := &pb.GetUserByIdRequest{Id: userID}

	// Set timeout for request
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Call gRPC service
	resp, err := userClient.GetUserById(ctx, userReq)
	if err != nil {
		log.Printf("Failed to fetch user from AAA service: %v", err)
		return nil, err
	}

	// Check if the response contains user data
	if resp.User == nil {
		log.Println("User not found in AAA service response")
		return nil, fmt.Errorf("user not found")
	}

	log.Printf("Successfully fetched user: %v", resp.User)
	return resp, nil
}

func CheckPermissionClient(ctx context.Context, userID string, actions []string, token string) (*pb.CheckPermissionResponse, error) {
	// Initialize gRPC connection with retry mechanism
	conn, err := InitializeGrpcClient(token, 3) 
	if err != nil {
		return nil, fmt.Errorf("failed to establish gRPC connection: %v", err)
	}
	defer conn.Close()

	// Create Permission Service Client
	userClient := pb.NewUserServiceClient(conn)

	// Prepare gRPC request
	permReq := &pb.CheckPermissionRequest{
		Principal: userID,
		Source:    "farmer_module",
		Actions:   actions,
	}

	// Set timeout for request
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Call gRPC service
	resp, err := userClient.CheckUserPermission(ctx, permReq)
	if err != nil {
		log.Printf("Failed to check permission for user %s: %v", userID, err)
		return nil, err
	}

	// Validate response
	if resp == nil || len(resp.Actions) == 0 {
		log.Printf("User %s does not have permission: %v", userID, actions)
		return nil, fmt.Errorf("user does not have required permission")
	}

	log.Printf("User %s has permission: %v", userID, actions)
	return resp, nil
}

// AssignRoleToUserClient assigns a role to a user via AAA service
func AssignRoleToUserClient(ctx context.Context, userID string, roles []string ,) (*pb.AssignRoleToUserResponse, error) {
	// Initialize gRPC connection with retry mechanism
	conn, err := InitializeGrpcClient("", 3) // Assuming no auth token is needed
	if err != nil {
		return nil, fmt.Errorf("failed to establish gRPC connection: %v", err)
	}
	defer conn.Close()

		// Create Permission Service Client
	userClient := pb.NewUserServiceClient(conn)

	// Prepare gRPC request
	roleReq := &pb.AssignRoleToUserRequest{
		UserId: userID,
		Roles:  roles,
	}

	// Set timeout for request
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Call gRPC service
	resp, err := userClient.AssignRoleToUser(ctx, roleReq)
	if err != nil {
		log.Printf("Failed to assign role to user %s: %v", userID, err)
		return nil, err
	}

	

	log.Printf("Successfully assigned roles to user %s: %v", userID, roles)
	return resp, nil
}

func ValidateActionClient(ctx context.Context, userID string, action string) (bool, error) {
	// Initialize gRPC connection
	conn, err := InitializeGrpcClient("", 3) // Reuse your existing connection logic
	if err != nil {
		return false, fmt.Errorf("gRPC connection failed: %v", err)
	}
	defer conn.Close()

	// Create Permission Service Client
	
	return true,nil
	}

	
	


