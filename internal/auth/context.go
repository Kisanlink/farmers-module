package auth

import (
	"context"
	"errors"
)

// Context keys for storing authentication information
type contextKey string

const (
	// UserContextKey is the key for storing user information in context
	UserContextKey contextKey = "user"
	// OrgContextKey is the key for storing organization information in context
	OrgContextKey contextKey = "org"
	// RequestIDKey is the key for storing request ID in context
	RequestIDKey contextKey = "request_id"
	// TokenKey is the key for storing the raw token in context
	TokenKey contextKey = "token"
)

// UserContext represents the authenticated user information
type UserContext struct {
	AAAUserID string   `json:"aaa_user_id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Roles     []string `json:"roles"`
}

// OrgContext represents the organization context
type OrgContext struct {
	AAAOrgID string `json:"aaa_org_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
}

// GetUserFromContext extracts user context from the request context
func GetUserFromContext(ctx context.Context) (*UserContext, error) {
	user, ok := ctx.Value(UserContextKey).(*UserContext)
	if !ok || user == nil {
		return nil, errors.New("user context not found")
	}
	return user, nil
}

// GetOrgFromContext extracts organization context from the request context
func GetOrgFromContext(ctx context.Context) (*OrgContext, error) {
	org, ok := ctx.Value(OrgContextKey).(*OrgContext)
	if !ok || org == nil {
		return nil, errors.New("organization context not found")
	}
	return org, nil
}

// GetRequestIDFromContext extracts request ID from the request context
func GetRequestIDFromContext(ctx context.Context) string {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	if !ok {
		return "unknown"
	}
	return requestID
}

// GetTokenFromContext extracts the raw token from the request context
func GetTokenFromContext(ctx context.Context) string {
	token, ok := ctx.Value(TokenKey).(string)
	if !ok {
		return ""
	}
	return token
}

// SetUserInContext sets user context in the request context
func SetUserInContext(ctx context.Context, user *UserContext) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

// SetOrgInContext sets organization context in the request context
func SetOrgInContext(ctx context.Context, org *OrgContext) context.Context {
	return context.WithValue(ctx, OrgContextKey, org)
}

// SetRequestIDInContext sets request ID in the request context
func SetRequestIDInContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// SetTokenInContext sets the raw token in the request context
func SetTokenInContext(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, TokenKey, token)
}
