package interfaces

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// Logger interface for structured logging
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	GetZapLogger() *zap.Logger
}

// EventEmitter interface for emitting audit and business events
type EventEmitter interface {
	EmitAuditEvent(event interface{}) error
	EmitBusinessEvent(eventType string, data interface{}) error
}

// UserInfo represents user information returned from token validation
type UserInfo struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Phone    string   `json:"phone"`
	Roles    []string `json:"roles"`
	OrgID    string   `json:"org_id,omitempty"`
	OrgName  string   `json:"org_name,omitempty"`
	OrgType  string   `json:"org_type,omitempty"`
}

// AAAService interface for authentication and authorization
type AAAService interface {
	// Token validation
	ValidateToken(ctx context.Context, token string) (*UserInfo, error)

	// Permission checking
	CheckPermission(ctx context.Context, subject, resource, action, object, orgID string) (bool, error)

	// User Management
	CreateUser(ctx context.Context, req interface{}) (interface{}, error)
	GetUser(ctx context.Context, userID string) (interface{}, error)
	GetUserByMobile(ctx context.Context, mobileNumber string) (interface{}, error)
	GetUserByEmail(ctx context.Context, email string) (interface{}, error)

	// Organization Management
	CreateOrganization(ctx context.Context, req interface{}) (interface{}, error)
	GetOrganization(ctx context.Context, orgID string) (interface{}, error)

	// User Group Management
	CreateUserGroup(ctx context.Context, req interface{}) (interface{}, error)
	AddUserToGroup(ctx context.Context, userID, groupID string) error
	RemoveUserFromGroup(ctx context.Context, userID, groupID string) error

	// Role and Permission Management
	AssignRole(ctx context.Context, userID, orgID, roleName string) error
	CheckUserRole(ctx context.Context, userID, roleName string) (bool, error)
	AssignPermissionToGroup(ctx context.Context, groupID, resource, action string) error

	// System Management
	SeedRolesAndPermissions(ctx context.Context) error
	HealthCheck(ctx context.Context) error
}

// Database interface for database operations
type Database interface {
	Connect() error
	Close() error
	Ping() error
	Migrate() error
}

// Repository base interface
type Repository interface {
	Create(ctx context.Context, entity interface{}) error
	GetByID(ctx context.Context, id string) (interface{}, error)
	Update(ctx context.Context, entity interface{}) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters interface{}) (interface{}, error)
}

// Service base interface
type Service interface {
	// Each service should implement its specific methods
}

// HealthChecker interface for health checking
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// Validator interface for input validation
type Validator interface {
	Validate(data interface{}) error
}

// Cache interface for caching operations
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
}
