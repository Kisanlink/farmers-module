package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kisanlink/protobuf/pb-aaa"

	"bou.ke/monkey"
	"github.com/Kisanlink/farmers-module/config"
	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/permission"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/Kisanlink/farmers-module/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFarmerService struct {
	mock.Mock
}

func (m *MockFarmerService) CreateFarmer(userId string, req models.FarmerSignupRequest) (*models.Farmer, *pb.GetUserByIdResponse, error) {
	args := m.Called(userId, req)
	return args.Get(0).(*models.Farmer), args.Get(1).(*pb.GetUserByIdResponse), args.Error(2)
}

func (m *MockFarmerService) FetchFarmers(userID, farmerID, kisansathiUserID string) ([]models.Farmer, error) {
	args := m.Called(userID, farmerID, kisansathiUserID)
	return args.Get(0).([]models.Farmer), args.Error(1)
}

func (m *MockFarmerService) FetchFarmersWithoutUserDetails(farmerID, kisansathiUserID string) ([]models.Farmer, error) {
	args := m.Called(farmerID, kisansathiUserID)
	return args.Get(0).([]models.Farmer), args.Error(1)
}

func TestFarmerSignupHandler_ExistingUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	utils.InitLogger()

	userId := "123e4567-e89b-12d3-a456-426614174000"
	userName := "Testfarmer"
	email := "farmer@example.com"

	reqBody := models.FarmerSignupRequest{
		UserId:       &userId,
		UserName:     &userName,
		Email:        &email,
		CountryCode:  "91",
		MobileNumber: 9876543210,
	}

	// Mock FarmerService
	mockService := new(MockFarmerService)
	mockFarmer := &models.Farmer{}
	mockUserDetails := &pb.GetUserByIdResponse{
		StatusCode: 200,
		Message:    "User fetched successfully",
		Success:    true,
		Data: &pb.User{
			Id: userId,
			// Add more fields as needed for the test
		},
	}

	mockService.
		On("CreateFarmer", userId, mock.Anything).
		Return(mockFarmer, mockUserDetails, nil)

	patch := monkey.Patch(services.AssignRoleToUserClient, func(ctx context.Context, userId, role string) (*pb.AssignRoleToUserResponse, error) {
		return &pb.AssignRoleToUserResponse{
			StatusCode: 200,
			Success:    true,
			Message:    "Role assigned",
		}, nil
	})

	defer patch.Unpatch()

	handler := handlers.NewFarmerHandler(mockService)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/signup", handler.FarmerSignupHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockService.AssertExpectations(t)
}

func TestFarmerSignupHandler_NewUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	userName := "TestNewFarmer"
	email := "newfarmer@example.com"
	aadhaar := "123412341234"

	reqBody := models.FarmerSignupRequest{
		UserId:        nil, // New user
		UserName:      &userName,
		Email:         &email,
		AadhaarNumber: &aadhaar,
		CountryCode:   "91",
		MobileNumber:  9876543210,
	}

	expectedUserID := "new-user-id-0001"

	// Patch CreateUserClient to simulate AAA service response
	patchCreateUser := monkey.Patch(services.CreateUserClient, func(req models.FarmerSignupRequest, _ string) (*pb.CreateUserResponse, error) {
		return &pb.CreateUserResponse{
			StatusCode: 200,
			Success:    true,
			Message:    "User created",
			Data: &pb.MinimalUser{
				Id: expectedUserID,
			},
		}, nil
	})
	defer patchCreateUser.Unpatch()

	// Patch AssignRoleToUserClient to simulate success
	patchAssignRole := monkey.Patch(services.AssignRoleToUserClient, func(ctx context.Context, userId, role string) (*pb.AssignRoleToUserResponse, error) {
		return &pb.AssignRoleToUserResponse{
			StatusCode: 200,
			Success:    true,
			Message:    "Role assigned",
		}, nil
	})
	defer patchAssignRole.Unpatch()

	// Mock FarmerService
	mockService := new(MockFarmerService)
	mockFarmer := &models.Farmer{}
	mockUserDetails := &pb.GetUserByIdResponse{
		StatusCode: 200,
		Message:    "User fetched successfully",
		Success:    true,
		Data: &pb.User{
			Id: expectedUserID,
		},
	}

	mockService.
		On("CreateFarmer", expectedUserID, mock.Anything).
		Return(mockFarmer, mockUserDetails, nil)

	handler := handlers.NewFarmerHandler(mockService)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/signup", handler.FarmerSignupHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockService.AssertExpectations(t)
}

func TestFarmerSignupHandler_InvalidJSON_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	// Malformed JSON (missing quotes and commas)
	invalidJSON := `{
		"user_id": "123e4567-e89b-12d3-a456-426614174000"
		"user_name": "invalid"
	}`

	// Mock FarmerService - shouldn't be called
	mockService := new(MockFarmerService)

	handler := handlers.NewFarmerHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/signup", handler.FarmerSignupHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockService.AssertNotCalled(t, "CreateFarmer")
}

func TestFarmerSignupHandler_MissingMobile_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	userName := "MissingMobileUser"
	email := "nomobile@example.com"
	aadhaar := "111122223333"

	// Note: MobileNumber is missing (zero value)
	reqBody := models.FarmerSignupRequest{
		UserId:        nil,
		UserName:      &userName,
		Email:         &email,
		AadhaarNumber: &aadhaar,
		CountryCode:   "91",
		// MobileNumber is 0 (missing)
	}

	mockService := new(MockFarmerService) // Should not be used

	handler := handlers.NewFarmerHandler(mockService)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/signup", handler.FarmerSignupHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockService.AssertNotCalled(t, "CreateFarmer")
}

func TestFarmerSignupHandler_KisansathiUserWithoutPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	userId := "existing-user-123"
	kisansathiId := "kisansathi-user-456"
	userName := "FarmerName"
	email := "farmer@example.com"

	reqBody := models.FarmerSignupRequest{
		UserId:           &userId,
		KisansathiUserId: &kisansathiId,
		UserName:         &userName,
		Email:            &email,
		CountryCode:      "91",
		MobileNumber:     9876543210,
	}

	// Patch CheckUserPermission to simulate lack of permission
	patchCheckPerm := monkey.Patch(permission.CheckUserPermission, func(ctx context.Context, userId string, perm string) (bool, int, string, string) {
		assert.Equal(t, kisansathiId, userId)
		assert.Equal(t, config.PERMISSION_KISANSATHI, perm)
		return false, http.StatusForbidden, "Permission denied", "User lacks Kisansathi permission"
	})
	defer patchCheckPerm.Unpatch()

	mockService := new(MockFarmerService) // Should not be called
	handler := handlers.NewFarmerHandler(mockService)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/signup", handler.FarmerSignupHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockService.AssertNotCalled(t, "CreateFarmer")
}

func TestFarmerSignupHandler_ExistingUser_CreateFarmerFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	userId := "existing-user-123"
	userName := "FailedFarmer"
	email := "fail@example.com"

	reqBody := models.FarmerSignupRequest{
		UserId:       &userId,
		UserName:     &userName,
		Email:        &email,
		CountryCode:  "91",
		MobileNumber: 9876543210,
	}

	// Patch permission check to allow
	patchCheckPerm := monkey.Patch(permission.CheckUserPermission, func(ctx context.Context, userId string, perm string) (bool, int, string, string) {
		return true, http.StatusOK, "Permission granted", ""
	})
	defer patchCheckPerm.Unpatch()

	// Mock service where CreateFarmer returns error
	mockService := new(MockFarmerService)
	mockService.
		On("CreateFarmer", userId, mock.Anything).
		Return(nil, nil, errors.New("simulated failure"))

	handler := handlers.NewFarmerHandler(mockService)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/signup", handler.FarmerSignupHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockService.AssertExpectations(t)
}

func TestFarmerSignupHandler_ExistingUser_AssignRoleFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	userId := "existing-user-123"
	userName := "RoleFailFarmer"
	email := "rolefail@example.com"

	reqBody := models.FarmerSignupRequest{
		UserId:       &userId,
		UserName:     &userName,
		Email:        &email,
		CountryCode:  "91",
		MobileNumber: 9876543210,
	}

	// Patch permission check to succeed
	patchPerm := monkey.Patch(permission.CheckUserPermission, func(ctx context.Context, userId string, perm string) (bool, int, string, string) {
		return true, http.StatusOK, "Permission granted", ""
	})
	defer patchPerm.Unpatch()

	// Patch AssignRoleToUserClient to simulate failure
	patchAssignRole := monkey.Patch(services.AssignRoleToUserClient, func(ctx context.Context, userId, role string) (*pb.AssignRoleToUserResponse, error) {
		return nil, errors.New("role assignment failed")
	})
	defer patchAssignRole.Unpatch()

	// Mock FarmerService - return success for CreateFarmer
	mockService := new(MockFarmerService)
	mockFarmer := &models.Farmer{}
	mockUserDetails := &pb.GetUserByIdResponse{
		StatusCode: 200,
		Message:    "Fetched user",
		Success:    true,
		Data:       &pb.User{Id: userId},
	}

	mockService.
		On("CreateFarmer", userId, mock.Anything).
		Return(mockFarmer, mockUserDetails, nil)

	handler := handlers.NewFarmerHandler(mockService)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/signup", handler.FarmerSignupHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockService.AssertExpectations(t)
}

func TestFarmerSignupHandler_NewUser_MissingUserNameOrAadhaar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	// Simulate a new user with missing AadhaarNumber
	userName := "TestNewFarmer"
	email := "newfarmer@example.com"

	reqBody := models.FarmerSignupRequest{
		UserId:        nil, // New user
		UserName:      &userName,
		Email:         &email,
		AadhaarNumber: nil, // Aadhaar missing
		CountryCode:   "91",
		MobileNumber:  9876543210,
	}

	mockService := new(MockFarmerService) // Should not be called
	handler := handlers.NewFarmerHandler(mockService)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/signup", handler.FarmerSignupHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockService.AssertNotCalled(t, "CreateFarmer")
}

func TestFarmerSignupHandler_NewUser_KisansathiUserWithoutPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	kisansathiId := "kisansathi-user-456"
	userName := "NewFarmer"
	email := "newfarmer@example.com"
	aadhaar := "123412341234"

	reqBody := models.FarmerSignupRequest{
		UserId:           nil, // New user
		KisansathiUserId: &kisansathiId,
		UserName:         &userName,
		Email:            &email,
		AadhaarNumber:    &aadhaar,
		CountryCode:      "91",
		MobileNumber:     9876543210,
	}

	// Patch CheckUserPermission to simulate missing permission
	patchCheckPerm := monkey.Patch(permission.CheckUserPermission, func(ctx context.Context, userId string, perm string) (bool, int, string, string) {
		assert.Equal(t, kisansathiId, userId)
		assert.Equal(t, config.PERMISSION_KISANSATHI, perm)
		return false, http.StatusForbidden, "Permission denied", "User lacks Kisansathi permission"
	})
	defer patchCheckPerm.Unpatch()

	mockService := new(MockFarmerService) // Should not be called
	handler := handlers.NewFarmerHandler(mockService)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/signup", handler.FarmerSignupHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockService.AssertNotCalled(t, "CreateFarmer")
}

func TestFarmerSignupHandler_NewUser_CreateUserClientFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	userName := "NewFarmer"
	email := "newfarmer@example.com"
	aadhaar := "123412341234"

	reqBody := models.FarmerSignupRequest{
		UserName:      &userName,
		Email:         &email,
		AadhaarNumber: &aadhaar,
		CountryCode:   "91",
		MobileNumber:  9876543210,
	}

	patchCreateUser := monkey.Patch(services.CreateUserClient, func(req models.FarmerSignupRequest, _ string) (*pb.CreateUserResponse, error) {
		return nil, fmt.Errorf("AAA service unavailable")
	})
	defer patchCreateUser.Unpatch()

	mockService := new(MockFarmerService)
	handler := handlers.NewFarmerHandler(mockService)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/signup", handler.FarmerSignupHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockService.AssertNotCalled(t, "CreateFarmer")
}

func TestFarmerSignupHandler_NewUser_CreateUserClientReturnsNilData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	userName := "NewFarmer"
	email := "newfarmer@example.com"
	aadhaar := "123412341234"

	reqBody := models.FarmerSignupRequest{
		UserName:      &userName,
		Email:         &email,
		AadhaarNumber: &aadhaar,
		CountryCode:   "91",
		MobileNumber:  9876543210,
	}

	patchCreateUser := monkey.Patch(services.CreateUserClient, func(req models.FarmerSignupRequest, _ string) (*pb.CreateUserResponse, error) {
		return &pb.CreateUserResponse{
			StatusCode: 200,
			Success:    true,
			Message:    "Success but no ID",
			Data:       nil, // Simulating missing user ID
		}, nil
	})
	defer patchCreateUser.Unpatch()

	mockService := new(MockFarmerService)
	handler := handlers.NewFarmerHandler(mockService)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/signup", handler.FarmerSignupHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockService.AssertNotCalled(t, "CreateFarmer")
}

func TestFarmerSignupHandler_NewUser_CreateFarmerFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	userName := "NewFarmer"
	email := "newfarmer@example.com"
	aadhaar := "123412341234"
	expectedUserID := "user-123"

	reqBody := models.FarmerSignupRequest{
		UserName:      &userName,
		Email:         &email,
		AadhaarNumber: &aadhaar,
		CountryCode:   "91",
		MobileNumber:  9876543210,
	}

	patchCreateUser := monkey.Patch(services.CreateUserClient, func(req models.FarmerSignupRequest, _ string) (*pb.CreateUserResponse, error) {
		return &pb.CreateUserResponse{
			StatusCode: 200,
			Success:    true,
			Message:    "User created",
			Data: &pb.MinimalUser{
				Id: expectedUserID,
			},
		}, nil
	})
	defer patchCreateUser.Unpatch()

	mockService := new(MockFarmerService)
	mockService.
		On("CreateFarmer", expectedUserID, mock.Anything).
		Return(nil, nil, fmt.Errorf("DB error"))

	handler := handlers.NewFarmerHandler(mockService)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/signup", handler.FarmerSignupHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockService.AssertExpectations(t)
}

func TestFarmerSignupHandler_NewUser_AssignRoleFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	userName := "NewFarmer"
	email := "newfarmer@example.com"
	aadhaar := "123412341234"
	expectedUserID := "user-123"

	reqBody := models.FarmerSignupRequest{
		UserName:      &userName,
		Email:         &email,
		AadhaarNumber: &aadhaar,
		CountryCode:   "91",
		MobileNumber:  9876543210,
	}

	patchCreateUser := monkey.Patch(services.CreateUserClient, func(req models.FarmerSignupRequest, _ string) (*pb.CreateUserResponse, error) {
		return &pb.CreateUserResponse{
			StatusCode: 200,
			Success:    true,
			Message:    "User created",
			Data: &pb.MinimalUser{
				Id: expectedUserID,
			},
		}, nil
	})
	defer patchCreateUser.Unpatch()

	patchAssignRole := monkey.Patch(services.AssignRoleToUserClient, func(ctx context.Context, userId, role string) (*pb.AssignRoleToUserResponse, error) {
		return nil, fmt.Errorf("Role assignment error")
	})
	defer patchAssignRole.Unpatch()

	mockService := new(MockFarmerService)
	mockFarmer := &models.Farmer{}
	mockUserDetails := &pb.GetUserByIdResponse{
		StatusCode: 200,
		Message:    "Success",
		Success:    true,
		Data:       &pb.User{Id: expectedUserID},
	}

	mockService.
		On("CreateFarmer", expectedUserID, mock.Anything).
		Return(mockFarmer, mockUserDetails, nil)

	handler := handlers.NewFarmerHandler(mockService)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/signup", handler.FarmerSignupHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockService.AssertExpectations(t)
}

func TestFarmerSignupHandler_KisansathiUserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	userName := "NewFarmer"
	email := "farmer@example.com"
	kisansathiId := "non-existent-user-123"

	reqBody := models.FarmerSignupRequest{
		UserName:         &userName,
		Email:            &email,
		CountryCode:      "91",
		MobileNumber:     9876543210,
		KisansathiUserId: &kisansathiId,
	}

	// Patch permission to simulate user not found
	patch := monkey.Patch(permission.CheckUserPermission, func(ctx context.Context, userId string, perm string) (bool, int, string, string) {
		return false, http.StatusUnauthorized, "Unauthorized", "User not found"
	})
	defer patch.Unpatch()

	mockService := new(MockFarmerService) // Should not be called
	handler := handlers.NewFarmerHandler(mockService)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := gin.Default()
	r.POST("/signup", handler.FarmerSignupHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockService.AssertNotCalled(t, "CreateFarmer")
}
