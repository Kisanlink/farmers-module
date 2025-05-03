package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kisanlink/protobuf/pb-aaa"

	"bou.ke/monkey"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFarmerRepository struct {
	mock.Mock
}

func (m *MockFarmerRepository) CreateFarmerEntry(farmer *models.Farmer) (*models.Farmer, error) {
	args := m.Called(farmer)
	return args.Get(0).(*models.Farmer), args.Error(1)
}

func (m *MockFarmerRepository) FetchFarmers(user_id, farmer_id, kisansathi_user_id string) ([]models.Farmer, error) {
	args := m.Called(user_id, farmer_id, kisansathi_user_id)
	return args.Get(0).([]models.Farmer), args.Error(1)
}

func TestCreateFarmer_Success(t *testing.T) {
	mockRepo := new(MockFarmerRepository)
	service := services.NewFarmerService(mockRepo)

	userId := "test-user-123"
	kisansathiId := "ks-001"
	req := models.FarmerSignupRequest{
		KisansathiUserId: &kisansathiId,
	}

	mockFarmer := &models.Farmer{
		UserId:           userId,
		KisansathiUserId: req.KisansathiUserId,
		IsActive:         true,
	}

	mockUserResp := &pb.GetUserByIdResponse{
		Data: &pb.User{
			Id:       userId,
			Username: "Test User",
		},
	}

	monkey.Patch(services.GetUserByIdClient, func(ctx context.Context, userID string) (*pb.GetUserByIdResponse, error) {
		assert.Equal(t, userId, userID)
		return mockUserResp, nil
	})
	defer monkey.Unpatch(services.GetUserByIdClient)

	mockRepo.On("CreateFarmerEntry", mock.Anything).Return(mockFarmer, nil)

	result, userDetails, err := service.CreateFarmer(userId, req)

	assert.NoError(t, err)
	assert.Equal(t, mockFarmer, result)
	assert.Equal(t, mockUserResp, userDetails)
	mockRepo.AssertExpectations(t)
}

func TestCreateFarmer_GetUserByIdFails(t *testing.T) {
	mockRepo := new(MockFarmerRepository)
	service := services.NewFarmerService(mockRepo)

	userId := "test-user-123"
	req := models.FarmerSignupRequest{}

	monkey.Patch(services.GetUserByIdClient, func(ctx context.Context, userID string) (*pb.GetUserByIdResponse, error) {
		assert.Equal(t, userId, userID)
		return nil, fmt.Errorf("user service unavailable")
	})
	defer monkey.Unpatch(services.GetUserByIdClient)

	result, userDetails, err := service.CreateFarmer(userId, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, userDetails)
	assert.Contains(t, err.Error(), "failed to fetch user details")
}

func TestCreateFarmer_CreateFarmerEntryFails(t *testing.T) {
	mockRepo := new(MockFarmerRepository)
	service := services.NewFarmerService(mockRepo)

	userId := "test-user-123"
	req := models.FarmerSignupRequest{}

	mockUserResp := &pb.GetUserByIdResponse{
		Data: &pb.User{Id: userId},
	}

	monkey.Patch(services.GetUserByIdClient, func(ctx context.Context, userID string) (*pb.GetUserByIdResponse, error) {
		return mockUserResp, nil
	})
	defer monkey.Unpatch(services.GetUserByIdClient)

	mockRepo.On("CreateFarmerEntry", mock.Anything).Return((*models.Farmer)(nil), fmt.Errorf("insert failed"))

	result, userDetails, err := service.CreateFarmer(userId, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, userDetails)
	assert.Contains(t, err.Error(), "failed to create farmer")
	mockRepo.AssertExpectations(t)
}
