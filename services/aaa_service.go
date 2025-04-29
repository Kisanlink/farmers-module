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
		log.Printf("InitializeGrpcClient: Attempt %d to establish gRPC connection", i+1)
		conn, err = grpcclient.GrpcClient(token)
		if err == nil {
			log.Println("InitializeGrpcClient: Successfully established gRPC connection")
			return conn, nil
		}
		log.Printf("InitializeGrpcClient: Failed to establish gRPC connection (attempt %d): %v", i+1, err)
		time.Sleep(10 * time.Second)
	}

	log.Printf("InitializeGrpcClient: Exhausted retries, failed to establish gRPC connection: %v", err)
	return nil, fmt.Errorf("failed to initialize gRPC client after %d retries: %v", retries, err)
}

func CreateUserClient(req models.FarmerSignupRequest, token string) (*pb.CreateUserResponse, error) {
	log.Println("CreateUserClient: Starting user creation process")

	// Initialize gRPC connection with retry mechanism
	conn, err := InitializeGrpcClient(token, 3)
	if err != nil {
		log.Printf("CreateUserClient: Failed to establish gRPC connection: %v", err)
		return nil, fmt.Errorf("failed to establish gRPC connection: %v", err)
	}
	defer conn.Close()

	// Create User Service Client
	user_client := pb.NewUserServiceClient(conn)
	log.Println("CreateUserClient: UserServiceClient initialized")

	// Prepare gRPC request
	user_request := &pb.CreateUserRequest{
		Username:      *req.UserName,
		MobileNumber:  req.MobileNumber,
		AadhaarNumber: *req.AadhaarNumber,
		Password:      "Default@123",
		CountryCode:   "+91",
	}
	log.Printf("CreateUserClient: Prepared gRPC request: %+v", user_request)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Call gRPC service
	log.Println("CreateUserClient: Sending gRPC request to RegisterUser")
	response, err := user_client.RegisterUser(ctx, user_request)
	if err != nil {
		log.Printf("CreateUserClient: Failed to create user via gRPC: %v", err)
		return nil, err
	}

	log.Printf("CreateUserClient: Successfully created user: %+v", response)
	return response, nil
}

func GetUserByIdClient(ctx context.Context, user_id string) (*pb.GetUserByIdResponse, error) {
	log.Printf("GetUserByIdClient: Fetching user with Id: %s", user_id)

	// Initialize gRPC connection with retry mechanism
	conn, err := InitializeGrpcClient("", 3) // Assuming no auth token is needed
	if err != nil {
		log.Printf("GetUserByIdClient: Failed to establish gRPC connection: %v", err)
		return nil, fmt.Errorf("failed to establish gRPC connection: %v", err)
	}
	defer conn.Close()

	// Create User Service Client
	user_client := pb.NewUserServiceClient(conn)
	log.Println("GetUserByIdClient: UserServiceClient initialized")

	// Prepare gRPC request
	user_req := &pb.GetUserByIdRequest{Id: user_id}
	log.Printf("GetUserByIdClient: Prepared gRPC request: %+v", user_req)

	// Set timeout for request
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Call gRPC service
	log.Println("GetUserByIdClient: Sending gRPC request to GetUserById")
	resp, err := user_client.GetUserById(ctx, user_req)
	if err != nil {
		log.Printf("GetUserByIdClient: Failed to fetch user from AAA service: %v", err)
		return nil, err
	}

	// Check if the response contains user data
	if resp.Data == nil {
		log.Println("GetUserByIdClient: User not found in AAA service response")
		return nil, fmt.Errorf("user not found")
	}

	// log.Printf("GetUserByIdClient: Successfully fetched user: %+v", resp.Data)
	return resp, nil
}

// AssignRoleToUserClient assigns a role to a user via AAA service
func AssignRoleToUserClient(ctx context.Context, user_id string, roles string) (*pb.AssignRoleToUserResponse, error) {
	log.Printf("AssignRoleToUserClient: Assigning role '%s' to user Id: %s", roles, user_id)

	// Initialize gRPC connection with retry mechanism
	conn, err := InitializeGrpcClient("", 3) // Assuming no auth token is needed
	if err != nil {
		log.Printf("AssignRoleToUserClient: Failed to establish gRPC connection: %v", err)
		return nil, fmt.Errorf("failed to establish gRPC connection: %v", err)
	}
	defer conn.Close()

	// Create Permission Service Client
	user_client := pb.NewUserServiceClient(conn)
	log.Println("AssignRoleToUserClient: UserServiceClient initialized")

	// Prepare gRPC request
	role_req := &pb.AssignRoleToUserRequest{
		UserId: user_id,
		Role:   roles,
	}
	log.Printf("AssignRoleToUserClient: Prepared gRPC request: %+v", role_req)

	// Set timeout for request
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Call gRPC service
	log.Println("AssignRoleToUserClient: Sending gRPC request to AssignRole")
	resp, err := user_client.AssignRole(ctx, role_req)
	if err != nil {
		log.Printf("AssignRoleToUserClient: Failed to assign role to user %s: %v", user_id, err)
		return nil, err
	}

	log.Printf("AssignRoleToUserClient: Successfully assigned roles to user %s: %+v", user_id, resp)
	return resp, nil
}
