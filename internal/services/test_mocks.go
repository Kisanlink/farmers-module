package services

import (
	"context"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/mock"
)

// MockAAAServiceShared is a shared mock implementation of the AAA service
type MockAAAServiceShared struct {
	mock.Mock
}

func (m *MockAAAServiceShared) CheckPermission(ctx context.Context, req interface{}) (bool, error) {
	args := m.Called(ctx, req)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAServiceShared) GetUser(ctx context.Context, userID string) (interface{}, error) {
	args := m.Called(ctx, userID)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceShared) GetOrganization(ctx context.Context, orgID string) (interface{}, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceShared) CheckUserRole(ctx context.Context, userID, roleName string) (bool, error) {
	args := m.Called(ctx, userID, roleName)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAServiceShared) CreateUser(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceShared) GetUserByMobile(ctx context.Context, mobile string) (interface{}, error) {
	args := m.Called(ctx, mobile)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceShared) GetUserByEmail(ctx context.Context, email string) (interface{}, error) {
	args := m.Called(ctx, email)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceShared) CreateOrganization(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceShared) CreateUserGroup(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceShared) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockAAAServiceShared) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockAAAServiceShared) AssignRole(ctx context.Context, userID, orgID, roleName string) error {
	args := m.Called(ctx, userID, orgID, roleName)
	return args.Error(0)
}

func (m *MockAAAServiceShared) AssignPermissionToGroup(ctx context.Context, groupID, resource, action string) error {
	args := m.Called(ctx, groupID, resource, action)
	return args.Error(0)
}

func (m *MockAAAServiceShared) ValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAAAServiceShared) SeedRolesAndPermissions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAAAServiceShared) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockFarmerLinkageRepoShared is a shared mock implementation of the farmer linkage repository
type MockFarmerLinkageRepoShared struct {
	mock.Mock
}

func (m *MockFarmerLinkageRepoShared) Create(ctx context.Context, entity *entities.FarmerLink) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockFarmerLinkageRepoShared) Update(ctx context.Context, entity *entities.FarmerLink) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockFarmerLinkageRepoShared) Find(ctx context.Context, filter *base.Filter) ([]*entities.FarmerLink, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*entities.FarmerLink), args.Error(1)
}

func (m *MockFarmerLinkageRepoShared) FindOne(ctx context.Context, filter *base.Filter) (*entities.FarmerLink, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*entities.FarmerLink), args.Error(1)
}

func (m *MockFarmerLinkageRepoShared) GetByID(ctx context.Context, id string) (*entities.FarmerLink, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.FarmerLink), args.Error(1)
}

func (m *MockFarmerLinkageRepoShared) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFarmerLinkageRepoShared) SetDBManager(dbManager interface{}) {
	m.Called(dbManager)
}
