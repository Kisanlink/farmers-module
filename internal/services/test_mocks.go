package services

import (
	"context"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/mock"
)

// MockAAAServiceShared is a shared mock implementation of the AAA service
type MockAAAServiceShared struct {
	mock.Mock
}

func (m *MockAAAServiceShared) CheckPermission(ctx context.Context, subject, resource, action, object, orgID string) (bool, error) {
	args := m.Called(ctx, subject, resource, action, object, orgID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAServiceShared) GetUser(ctx context.Context, userID string) (any, error) {
	args := m.Called(ctx, userID)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceShared) GetOrganization(ctx context.Context, orgID string) (any, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceShared) CheckUserRole(ctx context.Context, userID, roleName string) (bool, error) {
	args := m.Called(ctx, userID, roleName)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAServiceShared) CreateUser(ctx context.Context, req any) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceShared) GetUserByMobile(ctx context.Context, mobile string) (any, error) {
	args := m.Called(ctx, mobile)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceShared) GetUserByEmail(ctx context.Context, email string) (any, error) {
	args := m.Called(ctx, email)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceShared) CreateOrganization(ctx context.Context, req any) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceShared) CreateUserGroup(ctx context.Context, req any) (any, error) {
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

func (m *MockAAAServiceShared) ValidateToken(ctx context.Context, token string) (*interfaces.UserInfo, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.UserInfo), args.Error(1)
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

func (m *MockFarmerLinkageRepoShared) SetDBManager(dbManager any) {
	m.Called(dbManager)
}

// MockAAAService is a mock implementation of AAAService for testing
type MockAAAService struct {
	mock.Mock
}

func (m *MockAAAService) CheckPermission(ctx context.Context, subject, resource, action, object, orgID string) (bool, error) {
	args := m.Called(ctx, subject, resource, action, object, orgID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAService) GetUser(ctx context.Context, userID string) (any, error) {
	args := m.Called(ctx, userID)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) GetOrganization(ctx context.Context, orgID string) (any, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) CheckUserRole(ctx context.Context, userID, roleName string) (bool, error) {
	args := m.Called(ctx, userID, roleName)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAService) CreateUser(ctx context.Context, req any) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) GetUserByMobile(ctx context.Context, mobile string) (any, error) {
	args := m.Called(ctx, mobile)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) GetUserByEmail(ctx context.Context, email string) (any, error) {
	args := m.Called(ctx, email)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) CreateOrganization(ctx context.Context, req any) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) CreateUserGroup(ctx context.Context, req any) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockAAAService) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockAAAService) AssignRole(ctx context.Context, userID, orgID, roleName string) error {
	args := m.Called(ctx, userID, orgID, roleName)
	return args.Error(0)
}

func (m *MockAAAService) AssignPermissionToGroup(ctx context.Context, groupID, resource, action string) error {
	args := m.Called(ctx, groupID, resource, action)
	return args.Error(0)
}

func (m *MockAAAService) ValidateToken(ctx context.Context, token string) (*interfaces.UserInfo, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.UserInfo), args.Error(1)
}

func (m *MockAAAService) SeedRolesAndPermissions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAAAService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockBaseFilterableRepository is a generic mock for BaseFilterableRepository
type MockBaseFilterableRepository[T any] struct {
	mock.Mock
}

func (m *MockBaseFilterableRepository[T]) Create(ctx context.Context, entity T) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockBaseFilterableRepository[T]) Update(ctx context.Context, entity T) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockBaseFilterableRepository[T]) GetByID(ctx context.Context, id string, entity T) (T, error) {
	args := m.Called(ctx, id, entity)
	return args.Get(0).(T), args.Error(1)
}

func (m *MockBaseFilterableRepository[T]) Find(ctx context.Context, filter *base.Filter) ([]T, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]T), args.Error(1)
}

func (m *MockBaseFilterableRepository[T]) FindOne(ctx context.Context, filter *base.Filter) (T, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(T), args.Error(1)
}

func (m *MockBaseFilterableRepository[T]) Count(ctx context.Context, filter *base.Filter, entity T) (int64, error) {
	args := m.Called(ctx, filter, entity)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBaseFilterableRepository[T]) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBaseFilterableRepository[T]) SetDBManager(dbManager any) {
	m.Called(dbManager)
}

// MockDataQualityService is a mock implementation of DataQualityService for testing
type MockDataQualityService struct {
	mock.Mock
}

func (m *MockDataQualityService) ValidateGeometry(ctx context.Context, req any) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockDataQualityService) ReconcileAAALinks(ctx context.Context, req any) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockDataQualityService) RebuildSpatialIndexes(ctx context.Context, req any) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockDataQualityService) DetectFarmOverlaps(ctx context.Context, req any) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}
