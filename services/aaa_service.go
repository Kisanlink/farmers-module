package services

import (
	"context"
	"fmt"
	"time"

	grpcclient "github.com/Kisanlink/farmers-module/grpc_client"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/utils"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc"
)

func InitializeGrpcClient(token string, retries int) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	var err error

	for i := 0; i < retries; i++ {
		utils.Log.Infof("InitializeGrpcClient: Attempt %d to establish gRPC connection", i+1)
		conn, err = grpcclient.GrpcClient(token)
		if err == nil {
			utils.Log.Info("InitializeGrpcClient: Successfully established gRPC connection")
			return conn, nil
		}
		utils.Log.Errorf("InitializeGrpcClient: Failed to establish gRPC connection (attempt %d): %v", i+1, err)
		time.Sleep(10 * time.Second)
	}

	utils.Log.Errorf("InitializeGrpcClient: Exhausted retries, failed to establish gRPC connection: %v", err)
	return nil, fmt.Errorf("failed to initialize gRPC client after %d retries: %v", retries, err)
}

func CreateUserClient(req models.FarmerSignupRequest, token string) (*pb.CreateUserResponse, error) {
	utils.Log.Info("CreateUserClient: Starting user creation process")

	// Initialize gRPC connection with retry mechanism
	conn, err := InitializeGrpcClient(token, 3)
	if err != nil {
		utils.Log.Errorf("CreateUserClient: Failed to establish gRPC connection: %v", err)
		return nil, fmt.Errorf("failed to establish gRPC connection: %v", err)
	}
	defer conn.Close()

	// Create User Service Client
	user_client := pb.NewUserServiceClient(conn)
	utils.Log.Info("CreateUserClient: UserServiceClient initialized")

	// Prepare gRPC request
	user_request := &pb.CreateUserRequest{
		Username:      *req.UserName,
		MobileNumber:  req.MobileNumber,
		AadhaarNumber: *req.AadhaarNumber,
		Password:      "Default@123",
		CountryCode:   "+91",
	}
	utils.Log.Infof("CreateUserClient: Prepared gRPC request: %+v", user_request)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Call gRPC service
	utils.Log.Info("CreateUserClient: Sending gRPC request to RegisterUser")
	response, err := user_client.RegisterUser(ctx, user_request)
	if err != nil {
		utils.Log.Errorf("CreateUserClient: Failed to create user via gRPC: %v", err)
		return nil, err
	}

	utils.Log.Infof("CreateUserClient: Successfully created user: %+v", response)
	return response, nil
}

func GetUserByIdClient(ctx context.Context, user_id string) (*pb.GetUserByIdResponse, error) {
	utils.Log.Infof("GetUserByIdClient: Fetching user with Id: %s", user_id)

	// Initialize gRPC connection with retry mechanism
	conn, err := InitializeGrpcClient("", 3) // Assuming no auth token is needed
	if err != nil {
		utils.Log.Errorf("GetUserByIdClient: Failed to establish gRPC connection: %v", err)
		return nil, fmt.Errorf("failed to establish gRPC connection: %v", err)
	}
	defer conn.Close()

	// Create User Service Client
	user_client := pb.NewUserServiceClient(conn)
	utils.Log.Info("GetUserByIdClient: UserServiceClient initialized")

	// Prepare gRPC request
	user_req := &pb.GetUserByIdRequest{Id: user_id}
	utils.Log.Infof("GetUserByIdClient: Prepared gRPC request: %+v", user_req)

	// Set timeout for request
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Call gRPC service
	utils.Log.Info("GetUserByIdClient: Sending gRPC request to GetUserById")
	resp, err := user_client.GetUserById(ctx, user_req)
	if err != nil {
		utils.Log.Errorf("GetUserByIdClient: Failed to fetch user from AAA service: %v", err)
		return nil, err
	}

	// Check if the response contains user data
	if resp.Data == nil {
		utils.Log.Warn("GetUserByIdClient: User not found in AAA service response")
		return nil, fmt.Errorf("user not found")
	}

	// utils.Log.Infof("GetUserByIdClient: Successfully fetched user: %+v", resp.Data)
	return resp, nil
}

// AssignRoleToUserClient assigns a role to a user via AAA service
func AssignRoleToUserClient(ctx context.Context, user_id string, roles string) (*pb.AssignRoleToUserResponse, error) {
	utils.Log.Infof("AssignRoleToUserClient: Assigning role '%s' to user Id: %s", roles, user_id)

	// Initialize gRPC connection with retry mechanism
	conn, err := InitializeGrpcClient("", 3) // Assuming no auth token is needed
	if err != nil {
		utils.Log.Errorf("AssignRoleToUserClient: Failed to establish gRPC connection: %v", err)
		return nil, fmt.Errorf("failed to establish gRPC connection: %v", err)
	}
	defer conn.Close()

	// Create Permission Service Client
	user_client := pb.NewUserServiceClient(conn)
	utils.Log.Info("AssignRoleToUserClient: UserServiceClient initialized")

	// Prepare gRPC request
	role_req := &pb.AssignRoleToUserRequest{
		UserId: user_id,
		Role:   roles,
	}
	utils.Log.Infof("AssignRoleToUserClient: Prepared gRPC request: %+v", role_req)

	// Set timeout for request
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Call gRPC service
	utils.Log.Info("AssignRoleToUserClient: Sending gRPC request to AssignRole")
	resp, err := user_client.AssignRole(ctx, role_req)
	if err != nil {
		utils.Log.Errorf("AssignRoleToUserClient: Failed to assign role to user %s: %v", user_id, err)
		return nil, err
	}

	utils.Log.Infof("AssignRoleToUserClient: Successfully assigned roles to user %s: %+v", user_id, resp)
	return resp, nil
}
