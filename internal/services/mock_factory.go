package services

import (
	"fmt"
	"reflect"
)

// MockFactory provides centralized mock creation with validation and presets
type MockFactory struct {
	// DefaultPermissionMode controls the default permission behavior for AAA mocks
	DefaultPermissionMode PermissionMode
}

// PermissionMode defines the permission behavior for AAA service mocks
type PermissionMode int

const (
	// PermissionModeAllowAll allows all permissions (least secure, use only for basic tests)
	PermissionModeAllowAll PermissionMode = iota
	// PermissionModeDenyAll denies all permissions by default (most secure)
	PermissionModeDenyAll
	// PermissionModeCustom requires explicit permission matrix configuration
	PermissionModeCustom
)

// MockPreset defines common mock configurations for different test scenarios
type MockPreset string

const (
	// PresetAdmin configures mocks for admin user scenarios
	PresetAdmin MockPreset = "admin"
	// PresetFarmer configures mocks for farmer user scenarios
	PresetFarmer MockPreset = "farmer"
	// PresetKisanSathi configures mocks for KisanSathi user scenarios
	PresetKisanSathi MockPreset = "kisansathi"
	// PresetFPOManager configures mocks for FPO manager scenarios
	PresetFPOManager MockPreset = "fpo_manager"
	// PresetReadOnly configures mocks for read-only scenarios
	PresetReadOnly MockPreset = "readonly"
)

// NewMockFactory creates a new mock factory with the specified default permission mode
func NewMockFactory(defaultMode PermissionMode) *MockFactory {
	return &MockFactory{
		DefaultPermissionMode: defaultMode,
	}
}

// NewAAAServiceMock creates a new AAA service mock with the factory's default settings
func (f *MockFactory) NewAAAServiceMock() *MockAAAServiceShared {
	var defaultDeny bool
	switch f.DefaultPermissionMode {
	case PermissionModeDenyAll, PermissionModeCustom:
		defaultDeny = true
	case PermissionModeAllowAll:
		defaultDeny = false
	}

	return NewMockAAAServiceShared(defaultDeny)
}

// NewAAAServiceMockWithPreset creates an AAA service mock with a predefined permission preset
func (f *MockFactory) NewAAAServiceMockWithPreset(preset MockPreset, userID, orgID string) *MockAAAServiceShared {
	mock := f.NewAAAServiceMock()
	matrix := mock.GetPermissionMatrix()

	switch preset {
	case PresetAdmin:
		// Admin can do everything in their org
		matrix.AddAllowRule(userID, "*", "*", "*", orgID)

	case PresetFarmer:
		// Farmer can read and update their own data
		matrix.AddAllowRule(userID, "farmer", "read", "*", orgID)
		matrix.AddAllowRule(userID, "farmer", "update", userID, orgID)
		matrix.AddAllowRule(userID, "farm", "read", "*", orgID)
		matrix.AddAllowRule(userID, "farm", "create", "*", orgID)
		matrix.AddAllowRule(userID, "farm", "update", "*", orgID)
		matrix.AddAllowRule(userID, "cycle", "read", "*", orgID)
		matrix.AddAllowRule(userID, "cycle", "start", "*", orgID)
		matrix.AddAllowRule(userID, "activity", "read", "*", orgID)
		matrix.AddAllowRule(userID, "activity", "create", "*", orgID)

	case PresetKisanSathi:
		// KisanSathi can manage farmers and their farms
		matrix.AddAllowRule(userID, "farmer", "*", "*", orgID)
		matrix.AddAllowRule(userID, "farm", "*", "*", orgID)
		matrix.AddAllowRule(userID, "cycle", "*", "*", orgID)
		matrix.AddAllowRule(userID, "activity", "*", "*", orgID)

	case PresetFPOManager:
		// FPO Manager can manage organization and all farmers
		matrix.AddAllowRule(userID, "fpo", "*", "*", orgID)
		matrix.AddAllowRule(userID, "farmer", "*", "*", orgID)
		matrix.AddAllowRule(userID, "farm", "*", "*", orgID)
		matrix.AddAllowRule(userID, "cycle", "*", "*", orgID)
		matrix.AddAllowRule(userID, "activity", "*", "*", orgID)
		matrix.AddAllowRule(userID, "user", "create", "*", orgID)
		matrix.AddAllowRule(userID, "user", "read", "*", orgID)

	case PresetReadOnly:
		// Read-only access to everything in org
		matrix.AddAllowRule(userID, "*", "read", "*", orgID)
		matrix.AddAllowRule(userID, "*", "list", "*", orgID)
	}

	return mock
}

// ValidateMockInterface validates that a mock implements the expected interface
func (f *MockFactory) ValidateMockInterface(mock interface{}, expectedInterface interface{}) error {
	mockType := reflect.TypeOf(mock)
	expectedType := reflect.TypeOf(expectedInterface).Elem()

	if !mockType.Implements(expectedType) {
		return fmt.Errorf("mock type %s does not implement interface %s",
			mockType.String(), expectedType.String())
	}

	return nil
}

// NewCacheMock creates a new cache mock
func (f *MockFactory) NewCacheMock() *MockCache {
	return &MockCache{}
}

// NewEventEmitterMock creates a new event emitter mock
func (f *MockFactory) NewEventEmitterMock() *MockEventEmitter {
	return &MockEventEmitter{}
}

// NewDatabaseMock creates a new database mock
func (f *MockFactory) NewDatabaseMock() *MockDatabase {
	return &MockDatabase{}
}

// NewFarmerLinkageRepoMock creates a new farmer linkage repository mock
func (f *MockFactory) NewFarmerLinkageRepoMock() *MockFarmerLinkageRepoShared {
	return &MockFarmerLinkageRepoShared{}
}

// NewDataQualityServiceMock creates a new data quality service mock
func (f *MockFactory) NewDataQualityServiceMock() *MockDataQualityService {
	return &MockDataQualityService{}
}

// DefaultMockFactory provides a default factory instance with secure defaults (deny-all)
var DefaultMockFactory = NewMockFactory(PermissionModeDenyAll)

// TestMockFactory provides a permissive factory for simple tests (allow-all)
// WARNING: Only use this for basic tests that don't involve permission logic
var TestMockFactory = NewMockFactory(PermissionModeAllowAll)
