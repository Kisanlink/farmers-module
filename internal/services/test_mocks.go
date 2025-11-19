package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	cropCycleEntity "github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
	farmerentity "github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/mock"
)

// PermissionRule represents a single permission rule in the matrix
type PermissionRule struct {
	Subject  string // User ID or pattern (* for wildcard)
	Resource string // Resource type (e.g., "farmer", "farm", "cycle")
	Action   string // Action (e.g., "create", "read", "update", "delete")
	Object   string // Object ID or pattern (* for wildcard)
	OrgID    string // Organization ID or pattern (* for wildcard)
	Allow    bool   // Whether to allow or deny
}

// PermissionMatrix represents a set of permission rules for testing
type PermissionMatrix struct {
	mu          sync.RWMutex
	rules       []PermissionRule
	defaultDeny bool // If true, deny by default unless explicitly allowed
	logDenials  bool // If true, log permission denials
}

// NewPermissionMatrix creates a new permission matrix with deny-by-default
func NewPermissionMatrix(defaultDeny bool) *PermissionMatrix {
	return &PermissionMatrix{
		rules:       make([]PermissionRule, 0),
		defaultDeny: defaultDeny,
		logDenials:  true,
	}
}

// AddRule adds a permission rule to the matrix
func (pm *PermissionMatrix) AddRule(rule PermissionRule) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.rules = append(pm.rules, rule)
}

// AddAllowRule is a convenience method to add an allow rule
func (pm *PermissionMatrix) AddAllowRule(subject, resource, action, object, orgID string) {
	pm.AddRule(PermissionRule{
		Subject:  subject,
		Resource: resource,
		Action:   action,
		Object:   object,
		OrgID:    orgID,
		Allow:    true,
	})
}

// AddDenyRule is a convenience method to add a deny rule
func (pm *PermissionMatrix) AddDenyRule(subject, resource, action, object, orgID string) {
	pm.AddRule(PermissionRule{
		Subject:  subject,
		Resource: resource,
		Action:   action,
		Object:   object,
		OrgID:    orgID,
		Allow:    false,
	})
}

// CheckPermission checks if the given permission is allowed based on the matrix
func (pm *PermissionMatrix) CheckPermission(subject, resource, action, object, orgID string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Convert empty object to wildcard
	if object == "" {
		object = "*"
	}

	// Check rules in order (first match wins)
	for _, rule := range pm.rules {
		if pm.ruleMatches(rule, subject, resource, action, object, orgID) {
			return rule.Allow
		}
	}

	// If no rule matched, use default behavior
	return !pm.defaultDeny
}

// ruleMatches checks if a rule matches the given parameters
func (pm *PermissionMatrix) ruleMatches(rule PermissionRule, subject, resource, action, object, orgID string) bool {
	return pm.matches(rule.Subject, subject) &&
		pm.matches(rule.Resource, resource) &&
		pm.matches(rule.Action, action) &&
		pm.matches(rule.Object, object) &&
		pm.matches(rule.OrgID, orgID)
}

// matches checks if a pattern matches a value (supports * wildcard)
func (pm *PermissionMatrix) matches(pattern, value string) bool {
	if pattern == "*" {
		return true
	}
	return pattern == value
}

// Clear removes all rules from the matrix
func (pm *PermissionMatrix) Clear() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.rules = make([]PermissionRule, 0)
}

// MockAAAServiceShared is a shared mock implementation of the AAA service
// with configurable permission matrix for security testing
type MockAAAServiceShared struct {
	mock.Mock
	permissionMatrix *PermissionMatrix
	mu               sync.RWMutex
}

// NewMockAAAServiceShared creates a new mock AAA service with deny-by-default permissions
func NewMockAAAServiceShared(defaultDeny bool) *MockAAAServiceShared {
	return &MockAAAServiceShared{
		permissionMatrix: NewPermissionMatrix(defaultDeny),
	}
}

// SetPermissionMatrix sets the permission matrix for this mock
func (m *MockAAAServiceShared) SetPermissionMatrix(matrix *PermissionMatrix) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.permissionMatrix = matrix
}

// GetPermissionMatrix returns the permission matrix for configuration
func (m *MockAAAServiceShared) GetPermissionMatrix() *PermissionMatrix {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.permissionMatrix
}

func (m *MockAAAServiceShared) CheckPermission(ctx context.Context, subject, resource, action, object, orgID string) (bool, error) {
	// Get the permission matrix
	m.mu.RLock()
	matrix := m.permissionMatrix
	m.mu.RUnlock()

	// Always use permission matrix if it exists
	// The matrix handles both allow-all and deny-all modes internally
	if matrix != nil {
		// Use the permission matrix for this check
		allowed := matrix.CheckPermission(subject, resource, action, object, orgID)
		if !allowed && matrix.logDenials {
			// Log denied permissions for debugging
			fmt.Printf("Permission denied: subject=%s, resource=%s, action=%s, object=%s, orgID=%s\n",
				subject, resource, action, object, orgID)
		}
		return allowed, nil
	}

	// Fall back to testify mock behavior only if no matrix is configured
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

func (m *MockAAAServiceShared) SeedRolesAndPermissions(ctx context.Context, force bool) error {
	args := m.Called(ctx, force)
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

func (m *MockFarmerLinkageRepoShared) Create(ctx context.Context, entity *farmerentity.FarmerLink) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockFarmerLinkageRepoShared) Update(ctx context.Context, entity *farmerentity.FarmerLink) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockFarmerLinkageRepoShared) Find(ctx context.Context, filter *base.Filter) ([]*farmerentity.FarmerLink, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*farmerentity.FarmerLink), args.Error(1)
}

func (m *MockFarmerLinkageRepoShared) FindOne(ctx context.Context, filter *base.Filter) (*farmerentity.FarmerLink, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*farmerentity.FarmerLink), args.Error(1)
}

func (m *MockFarmerLinkageRepoShared) GetByID(ctx context.Context, id string) (*farmerentity.FarmerLink, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*farmerentity.FarmerLink), args.Error(1)
}

func (m *MockFarmerLinkageRepoShared) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFarmerLinkageRepoShared) SetDBManager(dbManager any) {
	m.Called(dbManager)
}

// MockAAAService is an alias for MockAAAServiceShared for backward compatibility
// This ensures all tests can use either MockAAAService or MockAAAServiceShared
// and get the same enhanced functionality including permission matrix support
type MockAAAService = MockAAAServiceShared

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

// MockNotificationService is a mock implementation of NotificationService for testing
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendOrphanedLinkAlert(ctx context.Context, fpoOrgID string, orphanedLinks []*farmerentity.FarmerLink) error {
	args := m.Called(ctx, fpoOrgID, orphanedLinks)
	return args.Error(0)
}

func (m *MockNotificationService) SendDataQualityAlert(ctx context.Context, alert DataQualityAlert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockNotificationService) SendNotification(ctx context.Context, req *NotificationRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockNotificationService) QueueNotification(ctx context.Context, req *NotificationRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

// MockCache is a mock implementation of Cache interface for testing
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCache) Clear(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockEventEmitter is a mock implementation of EventEmitter interface for testing
type MockEventEmitter struct {
	mock.Mock
}

func (m *MockEventEmitter) EmitAuditEvent(event interface{}) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockEventEmitter) EmitBusinessEvent(eventType string, data interface{}) error {
	args := m.Called(eventType, data)
	return args.Error(0)
}

// MockDatabase is a mock implementation of Database interface for testing
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatabase) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatabase) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatabase) Migrate() error {
	args := m.Called()
	return args.Error(0)
}

// MockFarmerRepository is a mock implementation of FarmerRepository for testing
type MockFarmerRepository struct {
	mock.Mock
}

func (m *MockFarmerRepository) Create(ctx context.Context, entity *farmerentity.Farmer) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockFarmerRepository) Update(ctx context.Context, entity *farmerentity.Farmer) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockFarmerRepository) Find(ctx context.Context, filter *base.Filter) ([]*farmerentity.Farmer, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*farmerentity.Farmer), args.Error(1)
}

func (m *MockFarmerRepository) FindOne(ctx context.Context, filter *base.Filter) (*farmerentity.Farmer, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*farmerentity.Farmer), args.Error(1)
}

func (m *MockFarmerRepository) GetByID(ctx context.Context, id string) (*farmerentity.Farmer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*farmerentity.Farmer), args.Error(1)
}

func (m *MockFarmerRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFarmerRepository) SoftDelete(ctx context.Context, id string, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockFarmerRepository) SetDBManager(dbManager any) {
	m.Called(dbManager)
}

func (m *MockFarmerRepository) Count(ctx context.Context, filter *base.Filter, model *farmerentity.Farmer) (int64, error) {
	args := m.Called(ctx, filter, model)
	return args.Get(0).(int64), args.Error(1)
}

// MockCropCycleRepository is a mock implementation of CropCycleRepository for testing
type MockCropCycleRepository struct {
	mock.Mock
}

func (m *MockCropCycleRepository) Create(ctx context.Context, entity *cropCycleEntity.CropCycle) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockCropCycleRepository) Update(ctx context.Context, entity *cropCycleEntity.CropCycle) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockCropCycleRepository) Find(ctx context.Context, filter *base.Filter) ([]*cropCycleEntity.CropCycle, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*cropCycleEntity.CropCycle), args.Error(1)
}

func (m *MockCropCycleRepository) FindOne(ctx context.Context, filter *base.Filter) (*cropCycleEntity.CropCycle, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cropCycleEntity.CropCycle), args.Error(1)
}

func (m *MockCropCycleRepository) GetByID(ctx context.Context, id string, entity *cropCycleEntity.CropCycle) (*cropCycleEntity.CropCycle, error) {
	args := m.Called(ctx, id, entity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cropCycleEntity.CropCycle), args.Error(1)
}

func (m *MockCropCycleRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCropCycleRepository) SetDBManager(dbManager any) {
	m.Called(dbManager)
}

func (m *MockCropCycleRepository) Count(ctx context.Context, filter *base.Filter, model *cropCycleEntity.CropCycle) (int64, error) {
	args := m.Called(ctx, filter, model)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCropCycleRepository) ValidateAreaAllocation(ctx context.Context, farmID string, cycleID string, requestedArea float64) error {
	args := m.Called(ctx, farmID, cycleID, requestedArea)
	return args.Error(0)
}

func (m *MockCropCycleRepository) GetTotalAllocatedArea(ctx context.Context, farmID string, excludeCycleID string) (float64, error) {
	args := m.Called(ctx, farmID, excludeCycleID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockCropCycleRepository) GetAreaAllocationSummary(ctx context.Context, farmID string) (interface{}, error) {
	args := m.Called(ctx, farmID)
	return args.Get(0), args.Error(1)
}

// MockFarmService is a mock implementation of FarmService for testing
type MockFarmService struct {
	mock.Mock
}

func (m *MockFarmService) CreateFarm(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockFarmService) GetFarm(ctx context.Context, farmID string) (interface{}, error) {
	args := m.Called(ctx, farmID)
	return args.Get(0), args.Error(1)
}

func (m *MockFarmService) UpdateFarm(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockFarmService) DeleteFarm(ctx context.Context, farmID string) error {
	args := m.Called(ctx, farmID)
	return args.Error(0)
}

func (m *MockFarmService) ListFarms(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockFarmService) GetFarmsByFarmer(ctx context.Context, farmerID string) (interface{}, error) {
	args := m.Called(ctx, farmerID)
	return args.Get(0), args.Error(1)
}

func (m *MockFarmService) AddSoilTypes(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockFarmService) AddIrrigationSources(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}
