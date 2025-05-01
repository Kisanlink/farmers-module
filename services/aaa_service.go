package services

import (
	"context"
	"fmt"
	"time"

	grpcclient "github.com/Kisanlink/farmers-module/grpc_client"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/utils"

	"github.com/kisanlink/protobuf/pb-aaa"
)

func CreateUserClient(req models.FarmerSignupRequest, token string) (*pb.CreateUserResponse, error) {
	utils.Log.Info("CreateUserClient: Starting user creation process")

	// Initialize the client only if not already initialized
	userClient, err := grpcclient.InitGrpcClient(token)
	if err != nil {
		utils.Log.Errorf("CreateUserClient: Failed to initialize UserServiceClient: %v", err)
		return nil, fmt.Errorf("failed to initialize UserServiceClient: %v", err)
	}
	utils.Log.Info("CreateUserClient: UserServiceClient initialized")

	userRequest := &pb.CreateUserRequest{
		Username:      *req.UserName,
		MobileNumber:  req.MobileNumber,
		AadhaarNumber: *req.AadhaarNumber,
		Password:      "Default@123",
		CountryCode:   "+91",
	}
	utils.Log.Infof("CreateUserClient: Prepared gRPC request: %+v", userRequest)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	utils.Log.Info("CreateUserClient: Sending gRPC request to RegisterUser")
	response, err := userClient.RegisterUser(ctx, userRequest)
	if err != nil {
		utils.Log.Errorf("CreateUserClient: Failed to create user via gRPC: %v", err)
		return nil, err
	}

	utils.Log.Infof("CreateUserClient: Successfully created user: %+v", response)
	return response, nil
}

func GetUserByIdClient(ctx context.Context, user_id string) (*pb.GetUserByIdResponse, error) {
	utils.Log.Info("GetUserByIdClient: Starting process to fetch user by ID")

	// Initialize the client only if not already initialized
	userClient, err := grpcclient.InitGrpcClient("")
	if err != nil {
		utils.Log.Errorf("GetUserByIdClient: Failed to initialize UserServiceClient: %v", err)
		return nil, fmt.Errorf("failed to initialize UserServiceClient: %v", err)
	}
	utils.Log.Info("GetUserByIdClient: UserServiceClient initialized")

	userRequest := &pb.GetUserByIdRequest{
		Id: user_id,
	}
	utils.Log.Infof("GetUserByIdClient: Prepared gRPC request: %+v", userRequest)

	// Rename ctx to newCtx to avoid conflict
	newCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	utils.Log.Info("GetUserByIdClient: Sending gRPC request to GetUserById")
	response, err := userClient.GetUserById(newCtx, userRequest)
	if err != nil {
		utils.Log.Errorf("GetUserByIdClient: Failed to fetch user by ID via gRPC: %v", err)
		return nil, err
	}

	utils.Log.Infof("GetUserByIdClient: Successfully fetched user by ID: %+v", response)
	return response, nil
}

// AssignRoleToUserClient assigns a role to a user via AAA service
func AssignRoleToUserClient(ctx context.Context, user_id string, roles string) (*pb.AssignRoleToUserResponse, error) {
	utils.Log.Info("AssignRoleToUserClient: Starting process to assign role to user")

	// Initialize the client only if not already initialized
	userClient, err := grpcclient.InitGrpcClient("")
	if err != nil {
		utils.Log.Errorf("AssignRoleToUserClient: Failed to initialize UserServiceClient: %v", err)
		return nil, fmt.Errorf("failed to initialize UserServiceClient: %v", err)
	}
	utils.Log.Info("AssignRoleToUserClient: UserServiceClient initialized")

	roleRequest := &pb.AssignRoleToUserRequest{
		UserId: user_id,
		Role:   roles,
	}
	utils.Log.Infof("AssignRoleToUserClient: Prepared gRPC request: %+v", roleRequest)

	// Use a different name for the new context to avoid overwriting the function parameter
	newCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	utils.Log.Info("AssignRoleToUserClient: Sending gRPC request to AssignRoleToUser")
	response, err := userClient.AssignRole(newCtx, roleRequest)
	if err != nil {
		utils.Log.Errorf("AssignRoleToUserClient: Failed to assign role to user via gRPC: %v", err)
		return nil, err
	}

	utils.Log.Infof("AssignRoleToUserClient: Successfully assigned role to user: %+v", response)
	return response, nil
}
