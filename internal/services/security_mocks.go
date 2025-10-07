package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"sync"
	"time"

	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/golang-jwt/jwt/v4"
)

// SecurityEnhancedMockAAA extends MockAAAServiceShared with security features
type SecurityEnhancedMockAAA struct {
	*MockAAAServiceShared
	jwtValidator    *JWTValidator
	rateLimiter     *RateLimiter
	auditLog        *AuditLogger
	securityEnabled bool
}

// JWTValidator handles JWT token validation for mocks
type JWTValidator struct {
	privateKey     *rsa.PrivateKey
	publicKey      *rsa.PublicKey
	issuer         string
	validityPeriod time.Duration
	mu             sync.RWMutex
}

// NewJWTValidator creates a new JWT validator with RSA key pair
func NewJWTValidator(issuer string, validityPeriod time.Duration) (*JWTValidator, error) {
	// Generate RSA key pair for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	return &JWTValidator{
		privateKey:     privateKey,
		publicKey:      &privateKey.PublicKey,
		issuer:         issuer,
		validityPeriod: validityPeriod,
	}, nil
}

// GenerateToken generates a valid JWT token for testing
func (v *JWTValidator) GenerateToken(userID, orgID string, roles []string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   userID,
		"org":   orgID,
		"roles": roles,
		"iss":   v.issuer,
		"iat":   now.Unix(),
		"exp":   now.Add(v.validityPeriod).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(v.privateKey)
}

// ValidateToken validates a JWT token and returns the claims
func (v *JWTValidator) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}

	// Validate issuer
	if iss, ok := claims["iss"].(string); !ok || iss != v.issuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	return &claims, nil
}

// RateLimiter simulates rate limiting for testing
type RateLimiter struct {
	limits       map[string]*RateLimit
	mu           sync.RWMutex
	enabled      bool
	defaultLimit int
	window       time.Duration
}

// RateLimit tracks rate limit state for a key
type RateLimit struct {
	count       int
	windowStart time.Time
	limit       int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(defaultLimit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limits:       make(map[string]*RateLimit),
		enabled:      true,
		defaultLimit: defaultLimit,
		window:       window,
	}
}

// Allow checks if a request is allowed under rate limits
func (rl *RateLimiter) Allow(key string) (bool, error) {
	if !rl.enabled {
		return true, nil
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limit, exists := rl.limits[key]

	if !exists || now.Sub(limit.windowStart) > rl.window {
		// New window
		rl.limits[key] = &RateLimit{
			count:       1,
			windowStart: now,
			limit:       rl.defaultLimit,
		}
		return true, nil
	}

	// Check if limit exceeded
	if limit.count >= limit.limit {
		return false, fmt.Errorf("rate limit exceeded for key: %s", key)
	}

	// Increment count
	limit.count++
	return true, nil
}

// SetLimit sets a custom limit for a specific key
func (rl *RateLimiter) SetLimit(key string, limit int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if existing, ok := rl.limits[key]; ok {
		existing.limit = limit
	} else {
		rl.limits[key] = &RateLimit{
			count:       0,
			windowStart: time.Now(),
			limit:       limit,
		}
	}
}

// Reset resets rate limit counters for a key
func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.limits, key)
}

// Enable enables rate limiting
func (rl *RateLimiter) Enable() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.enabled = true
}

// Disable disables rate limiting
func (rl *RateLimiter) Disable() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.enabled = false
}

// AuditLogger tracks security-critical operations for testing
type AuditLogger struct {
	events  []AuditEvent
	mu      sync.RWMutex
	enabled bool
}

// AuditEvent represents a security audit event
type AuditEvent struct {
	Timestamp time.Time
	EventType string
	UserID    string
	OrgID     string
	Resource  string
	Action    string
	Result    string // "success", "denied", "error"
	Details   map[string]interface{}
	IPAddress string
	UserAgent string
	RequestID string
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		events:  make([]AuditEvent, 0),
		enabled: true,
	}
}

// LogEvent logs an audit event
func (al *AuditLogger) LogEvent(event AuditEvent) {
	if !al.enabled {
		return
	}

	al.mu.Lock()
	defer al.mu.Unlock()

	event.Timestamp = time.Now()
	al.events = append(al.events, event)
}

// GetEvents returns all logged events
func (al *AuditLogger) GetEvents() []AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()
	return append([]AuditEvent{}, al.events...)
}

// GetEventsByUser returns events for a specific user
func (al *AuditLogger) GetEventsByUser(userID string) []AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var filtered []AuditEvent
	for _, event := range al.events {
		if event.UserID == userID {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// GetEventsByType returns events of a specific type
func (al *AuditLogger) GetEventsByType(eventType string) []AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var filtered []AuditEvent
	for _, event := range al.events {
		if event.EventType == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// Clear clears all audit events
func (al *AuditLogger) Clear() {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.events = make([]AuditEvent, 0)
}

// NewSecurityEnhancedMockAAA creates a new security-enhanced mock AAA service
func NewSecurityEnhancedMockAAA(enableSecurity bool) (*SecurityEnhancedMockAAA, error) {
	validator, err := NewJWTValidator("test-issuer", 1*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT validator: %w", err)
	}

	return &SecurityEnhancedMockAAA{
		MockAAAServiceShared: NewMockAAAServiceShared(true), // Deny by default
		jwtValidator:         validator,
		rateLimiter:          NewRateLimiter(100, 1*time.Minute),
		auditLog:             NewAuditLogger(),
		securityEnabled:      enableSecurity,
	}, nil
}

// CheckPermission overrides base method with security enhancements
func (s *SecurityEnhancedMockAAA) CheckPermission(ctx context.Context, subject, resource, action, object, orgID string) (bool, error) {
	// Rate limiting
	if s.securityEnabled {
		rateLimitKey := fmt.Sprintf("check_permission:%s", subject)
		allowed, err := s.rateLimiter.Allow(rateLimitKey)
		if err != nil || !allowed {
			// Log rate limit violation
			s.auditLog.LogEvent(AuditEvent{
				EventType: "rate_limit_exceeded",
				UserID:    subject,
				OrgID:     orgID,
				Resource:  resource,
				Action:    action,
				Result:    "denied",
				Details: map[string]interface{}{
					"reason": "rate limit exceeded",
				},
			})
			return false, fmt.Errorf("rate limit exceeded")
		}
	}

	// Call base permission check
	allowed, err := s.MockAAAServiceShared.CheckPermission(ctx, subject, resource, action, object, orgID)

	// Audit logging
	if s.securityEnabled {
		result := "denied"
		if err == nil && allowed {
			result = "success"
		} else if err != nil {
			result = "error"
		}

		s.auditLog.LogEvent(AuditEvent{
			EventType: "permission_check",
			UserID:    subject,
			OrgID:     orgID,
			Resource:  resource,
			Action:    action,
			Result:    result,
			Details: map[string]interface{}{
				"object":  object,
				"allowed": allowed,
				"error":   err,
			},
		})
	}

	return allowed, err
}

// ValidateToken validates a JWT token with security checks
func (s *SecurityEnhancedMockAAA) ValidateToken(ctx context.Context, token string) (*interfaces.UserInfo, error) {
	if !s.securityEnabled {
		// Fall back to base mock behavior
		return s.MockAAAServiceShared.ValidateToken(ctx, token)
	}

	// Validate JWT signature and claims
	claims, err := s.jwtValidator.ValidateToken(token)
	if err != nil {
		s.auditLog.LogEvent(AuditEvent{
			EventType: "token_validation_failed",
			Result:    "error",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Extract user info from claims
	userID, _ := (*claims)["sub"].(string)
	orgID, _ := (*claims)["org"].(string)

	// Log successful validation
	s.auditLog.LogEvent(AuditEvent{
		EventType: "token_validation_success",
		UserID:    userID,
		OrgID:     orgID,
		Result:    "success",
	})

	return &interfaces.UserInfo{
		UserID: userID,
		OrgID:  orgID,
	}, nil
}

// GenerateTestToken generates a valid test token
func (s *SecurityEnhancedMockAAA) GenerateTestToken(userID, orgID string, roles []string) (string, error) {
	return s.jwtValidator.GenerateToken(userID, orgID, roles)
}

// GetAuditEvents returns all audit events for testing verification
func (s *SecurityEnhancedMockAAA) GetAuditEvents() []AuditEvent {
	return s.auditLog.GetEvents()
}

// GetRateLimiter returns the rate limiter for test configuration
func (s *SecurityEnhancedMockAAA) GetRateLimiter() *RateLimiter {
	return s.rateLimiter
}

// GetJWTValidator returns the JWT validator for test configuration
func (s *SecurityEnhancedMockAAA) GetJWTValidator() *JWTValidator {
	return s.jwtValidator
}

// EnableSecurity enables security features
func (s *SecurityEnhancedMockAAA) EnableSecurity() {
	s.securityEnabled = true
}

// DisableSecurity disables security features for basic testing
func (s *SecurityEnhancedMockAAA) DisableSecurity() {
	s.securityEnabled = false
}
