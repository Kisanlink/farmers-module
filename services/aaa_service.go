package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	grpcclient "github.com/Kisanlink/farmers-module/grpc_client"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func CreateUserClient(
	ctx context.Context,
	req models.FarmerSignupRequest,
	token string, // keep if you really need auth; otherwise "" is fine
) (*pb.CreateUserResponse, error) {
	log.Println("CreateUserClient: Starting user creation process")

	// Initialize gRPC connection with retry mechanism
	conn, err := InitializeGrpcClient(token, 3)
	if err != nil {
		log.Printf("CreateUserClient: Failed to establish gRPC connection: %v", err)
		return nil, fmt.Errorf("failed to establish gRPC connection: %v", err)
	}
	defer conn.Close()

	// Create User Service Client
	userClient := pb.NewUserServiceClient(conn)
	log.Println("CreateUserClient: UserServiceClient initialized")

	// Prepare gRPC request
	userRequest := &pb.CreateUserRequest{
		Username:     *req.UserName,
		MobileNumber: req.MobileNumber,
		Password:     "Default@123",
		CountryCode:  "+91",
	}

	// Aadhaar is optional
	if req.AadhaarNumber != nil {
		userRequest.AadhaarNumber = *req.AadhaarNumber
	}

	log.Printf("CreateUserClient: Prepared gRPC request: %+v", userRequest)

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// Call gRPC service
	log.Println("CreateUserClient: Sending gRPC request to RegisterUser")
	response, err := userClient.RegisterUser(ctx, userRequest)
	if err != nil {
		log.Printf("CreateUserClient: Failed to create user via gRPC: %v", err)
		return nil, err
	}

	log.Printf("CreateUserClient: Successfully created user: %+v", response)
	return response, nil
}

func GetUserByIdClient(ctx context.Context, userId string) (*pb.GetUserByIdResponse, error) {
	log.Printf("GetUserByIdClient: Fetching user with Id: %s", userId)

	// Initialize gRPC connection with retry mechanism
	conn, err := InitializeGrpcClient("", 3) // Assuming no auth token is needed
	if err != nil {
		log.Printf("GetUserByIdClient: Failed to establish gRPC connection: %v", err)
		return nil, fmt.Errorf("failed to establish gRPC connection: %v", err)
	}
	defer conn.Close()

	// Create User Service Client
	userClient := pb.NewUserServiceClient(conn)
	log.Println("GetUserByIdClient: UserServiceClient initialized")

	// Prepare gRPC request
	userReq := &pb.GetUserByIdRequest{Id: userId}
	log.Printf("GetUserByIdClient: Prepared gRPC request: %+v", userReq)

	// Set timeout for request
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Call gRPC service
	log.Println("GetUserByIdClient: Sending gRPC request to GetUserById")
	resp, err := userClient.GetUserById(ctx, userReq)
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
func AssignRoleToUserClient(ctx context.Context, userId string, roles string) (*pb.AssignRoleToUserResponse, error) {
	log.Printf("AssignRoleToUserClient: Assigning role '%s' to user Id: %s", roles, userId)

	// Initialize gRPC connection with retry mechanism
	conn, err := InitializeGrpcClient("", 3) // Assuming no auth token is needed
	if err != nil {
		log.Printf("AssignRoleToUserClient: Failed to establish gRPC connection: %v", err)
		return nil, fmt.Errorf("failed to establish gRPC connection: %v", err)
	}
	defer conn.Close()

	// Create Permission Service Client
	userClient := pb.NewUserServiceClient(conn)
	log.Println("AssignRoleToUserClient: UserServiceClient initialized")

	// Prepare gRPC request
	roleReq := &pb.AssignRoleToUserRequest{
		UserId: userId,
		Role:   roles,
	}
	log.Printf("AssignRoleToUserClient: Prepared gRPC request: %+v", roleReq)

	// Set timeout for request
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Call gRPC service
	log.Println("AssignRoleToUserClient: Sending gRPC request to AssignRole")
	resp, err := userClient.AssignRole(ctx, roleReq)
	if err != nil {
		log.Printf("AssignRoleToUserClient: Failed to assign role to user %s: %v", userId, err)
		return nil, err
	}

	log.Printf("AssignRoleToUserClient: Successfully assigned roles to user %s: %+v", userId, resp)
	return resp, nil
}

// GetUserByMobileClient asks AAA for a user by mobile number.
// If the user does not exist, it returns (nil, nil).
// Any other gRPC failure bubbles up as an error.
func GetUserByMobileClient(
	ctx context.Context,
	mobile uint64,
) (*pb.GetUserByMobileNumberResponse, error) {

	// ─── 1. gRPC connection with retries ────────────────────────────────────
	conn, err := InitializeGrpcClient("", 3) // no auth token for internal calls
	if err != nil {
		return nil, fmt.Errorf("grpc init: %w", err)
	}
	defer conn.Close()

	cli := pb.NewUserServiceClient(conn)

	// ─── 2. Prepare request ─────────────────────────────────────────────────
	req := &pb.GetUserByMobileNumberRequest{
		MobileNumber: mobile,
	}

	// Put a tight timeout on the downstream call
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// ─── 3. Call AAA service ───────────────────────────────────────────────
	resp, err := cli.GetUserByMobileNumber(ctx, req)
	if err != nil {
		// Normalise "user not found" so caller can treat it as nil, nil
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound ||
				strings.Contains(st.Message(), "user not found") ||
				strings.Contains(st.Message(), "record not found") {
				return nil, nil
			}
		}
		// Any other error is a real dependency failure
		return nil, err
	}

	// ─── 4. No data? Treat as not-found ─────────────────────────────────────
	if resp.GetData() == nil || resp.GetData().GetId() == "" {
		return nil, nil
	}

	return resp, nil
}
