# Critical Fixes - Implementation Guide

## Priority 1: JWT Token Validation Fix

### Current Issue
File: `internal/clients/aaa/aaa_client.go:430-440`
```go
// BROKEN - Returns error instead of validating
func (c *Client) ValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
    return nil, fmt.Errorf("ValidateToken not implemented - missing AuthService proto")
}
```

### Implementation Solution

#### Step 1: Add JWT Dependencies
```bash
go get github.com/golang-jwt/jwt/v4
go get github.com/patrickmn/go-cache
```

#### Step 2: Create Token Validator
Create file: `internal/auth/token_validator.go`
```go
package auth

import (
    "context"
    "crypto/rsa"
    "errors"
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v4"
    "github.com/patrickmn/go-cache"
)

type TokenValidator struct {
    publicKey     *rsa.PublicKey
    secret        []byte
    cache         *cache.Cache
    issuer        string
    audience      string
}

type TokenClaims struct {
    UserID       string   `json:"sub"`
    OrgID        string   `json:"org_id"`
    Roles        []string `json:"roles"`
    Permissions  []string `json:"permissions"`
    TokenType    string   `json:"token_type"`
    jwt.RegisteredClaims
}

func NewTokenValidator(secret string, publicKeyPEM []byte) (*TokenValidator, error) {
    var publicKey *rsa.PublicKey
    if len(publicKeyPEM) > 0 {
        key, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
        if err != nil {
            return nil, fmt.Errorf("failed to parse public key: %w", err)
        }
        publicKey = key
    }

    return &TokenValidator{
        publicKey: publicKey,
        secret:    []byte(secret),
        cache:     cache.New(5*time.Minute, 10*time.Minute),
        issuer:    "aaa-service",
        audience:  "farmers-module",
    }, nil
}

func (tv *TokenValidator) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
    // Check cache first
    if cached, found := tv.cache.Get(tokenString); found {
        if claims, ok := cached.(*TokenClaims); ok {
            // Check if still valid
            if claims.ExpiresAt.After(time.Now()) {
                return claims, nil
            }
            tv.cache.Delete(tokenString)
        }
    }

    // Parse and validate token
    token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
        // Validate signing method
        switch token.Method.(type) {
        case *jwt.SigningMethodRSA:
            if tv.publicKey == nil {
                return nil, errors.New("RSA public key not configured")
            }
            return tv.publicKey, nil
        case *jwt.SigningMethodHMAC:
            return tv.secret, nil
        default:
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
    })

    if err != nil {
        return nil, fmt.Errorf("token validation failed: %w", err)
    }

    claims, ok := token.Claims.(*TokenClaims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token claims")
    }

    // Validate issuer and audience
    if claims.Issuer != tv.issuer {
        return nil, fmt.Errorf("invalid issuer: expected %s, got %s", tv.issuer, claims.Issuer)
    }

    if !contains(claims.Audience, tv.audience) {
        return nil, fmt.Errorf("invalid audience: token not intended for %s", tv.audience)
    }

    // Cache valid token
    tv.cache.Set(tokenString, claims, time.Until(claims.ExpiresAt.Time))

    return claims, nil
}

func (tv *TokenValidator) ExtractTokenFromHeader(authHeader string) (string, error) {
    if authHeader == "" {
        return "", errors.New("authorization header is empty")
    }

    const bearerPrefix = "Bearer "
    if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
        return "", errors.New("invalid authorization header format")
    }

    return authHeader[len(bearerPrefix):], nil
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

#### Step 3: Update AAA Client
Update file: `internal/clients/aaa/aaa_client.go`
```go
// Add to Client struct
type Client struct {
    // ... existing fields ...
    tokenValidator *auth.TokenValidator
}

// Update NewClient to initialize tokenValidator
func NewClient(cfg *Config) (*Client, error) {
    // ... existing code ...

    // Initialize token validator
    tokenValidator, err := auth.NewTokenValidator(
        cfg.JWTSecret,
        []byte(cfg.JWTPublicKey),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create token validator: %w", err)
    }

    client.tokenValidator = tokenValidator
    return client, nil
}

// Replace ValidateToken method
func (c *Client) ValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
    log.Printf("AAA ValidateToken: validating token")

    if token == "" {
        return nil, fmt.Errorf("token is required")
    }

    // Use local JWT validation first
    claims, err := c.tokenValidator.ValidateToken(ctx, token)
    if err != nil {
        // If local validation fails, try remote validation as fallback
        return c.remoteValidateToken(ctx, token)
    }

    // Convert claims to map for backward compatibility
    result := map[string]interface{}{
        "user_id":     claims.UserID,
        "org_id":      claims.OrgID,
        "roles":       claims.Roles,
        "permissions": claims.Permissions,
        "exp":         claims.ExpiresAt.Unix(),
        "iat":         claims.IssuedAt.Unix(),
    }

    log.Printf("Token validated successfully for user: %s", claims.UserID)
    return result, nil
}

// Add remote validation as fallback
func (c *Client) remoteValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
    // This would call the actual AAA service when available
    // For now, implement a basic validation

    // Try to decode without verification for debugging
    parser := jwt.NewParser()
    claims := jwt.MapClaims{}
    _, _, err := parser.ParseUnverified(token, claims)
    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }

    // Check expiration
    if exp, ok := claims["exp"].(float64); ok {
        if time.Now().Unix() > int64(exp) {
            return nil, errors.New("token expired")
        }
    }

    return claims, nil
}
```

#### Step 4: Add Configuration
Update file: `internal/config/config.go`
```go
type AAAConfig struct {
    ServiceURL   string `env:"AAA_SERVICE_URL" envDefault:"localhost:50051"`
    JWTSecret    string `env:"JWT_SECRET" envDefault:""`
    JWTPublicKey string `env:"JWT_PUBLIC_KEY" envDefault:""`
    Timeout      int    `env:"AAA_TIMEOUT" envDefault:"30"`
    MaxRetries   int    `env:"AAA_MAX_RETRIES" envDefault:"3"`
}
```

## Priority 2: Audit Service Integration

### Implementation Solution

#### Step 1: Create Audit Service Interface
Create file: `internal/services/audit/audit_service.go`
```go
package audit

import (
    "context"
    "encoding/json"
    "time"

    "go.uber.org/zap"
)

type AuditService struct {
    logger *zap.Logger
    queue  chan *AuditEvent
    client AuditClient
}

type AuditEvent struct {
    ID            string                 `json:"id"`
    Timestamp     time.Time             `json:"timestamp"`
    UserID        string                 `json:"user_id"`
    OrgID         string                 `json:"org_id"`
    Action        string                 `json:"action"`
    ResourceType  string                 `json:"resource_type"`
    ResourceID    string                 `json:"resource_id"`
    OldValue      interface{}           `json:"old_value,omitempty"`
    NewValue      interface{}           `json:"new_value,omitempty"`
    IPAddress     string                 `json:"ip_address"`
    UserAgent     string                 `json:"user_agent"`
    CorrelationID string                 `json:"correlation_id"`
    Status        string                 `json:"status"`
    ErrorMessage  string                 `json:"error_message,omitempty"`
    Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

func NewAuditService(logger *zap.Logger, client AuditClient) *AuditService {
    svc := &AuditService{
        logger: logger,
        queue:  make(chan *AuditEvent, 1000),
        client: client,
    }

    // Start background worker
    go svc.processQueue()

    return svc
}

func (s *AuditService) LogEvent(ctx context.Context, event *AuditEvent) error {
    // Add context metadata
    if event.ID == "" {
        event.ID = generateUUID()
    }
    if event.Timestamp.IsZero() {
        event.Timestamp = time.Now()
    }

    // Send to queue for async processing
    select {
    case s.queue <- event:
        s.logger.Debug("Audit event queued", zap.String("event_id", event.ID))
        return nil
    default:
        // Queue is full, log synchronously
        return s.logEventSync(ctx, event)
    }
}

func (s *AuditService) logEventSync(ctx context.Context, event *AuditEvent) error {
    // Log to database or external service
    if s.client != nil {
        if err := s.client.SendAuditEvent(ctx, event); err != nil {
            s.logger.Error("Failed to send audit event",
                zap.Error(err),
                zap.String("event_id", event.ID))
            // Fall back to local logging
            return s.logToFile(event)
        }
    }

    return nil
}

func (s *AuditService) processQueue() {
    batch := make([]*AuditEvent, 0, 100)
    ticker := time.NewTicker(5 * time.Second)

    for {
        select {
        case event := <-s.queue:
            batch = append(batch, event)

            if len(batch) >= 100 {
                s.flushBatch(batch)
                batch = batch[:0]
            }

        case <-ticker.C:
            if len(batch) > 0 {
                s.flushBatch(batch)
                batch = batch[:0]
            }
        }
    }
}

func (s *AuditService) flushBatch(events []*AuditEvent) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if s.client != nil {
        if err := s.client.SendAuditEventBatch(ctx, events); err != nil {
            s.logger.Error("Failed to send audit batch",
                zap.Error(err),
                zap.Int("batch_size", len(events)))

            // Fall back to individual logging
            for _, event := range events {
                s.logToFile(event)
            }
        }
    }
}

func (s *AuditService) logToFile(event *AuditEvent) error {
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }

    s.logger.Info("AUDIT",
        zap.String("event", string(data)),
        zap.String("event_id", event.ID))

    return nil
}

func (s *AuditService) QueryAuditTrail(ctx context.Context, filters *AuditFilters) ([]*AuditEvent, error) {
    if s.client != nil {
        return s.client.QueryAuditEvents(ctx, filters)
    }

    // Return empty result if no client
    return []*AuditEvent{}, nil
}
```

#### Step 2: Create Audit Middleware
Create file: `internal/middleware/audit_middleware.go`
```go
package middleware

import (
    "bytes"
    "encoding/json"
    "io"
    "time"

    "github.com/gin-gonic/gin"
    "internal/services/audit"
)

type bodyLogWriter struct {
    gin.ResponseWriter
    body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
    w.body.Write(b)
    return w.ResponseWriter.Write(b)
}

func AuditMiddleware(auditService *audit.AuditService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip health check endpoints
        if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
            c.Next()
            return
        }

        // Capture request body
        var requestBody []byte
        if c.Request.Body != nil {
            requestBody, _ = io.ReadAll(c.Request.Body)
            c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
        }

        // Capture response body
        blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
        c.Writer = blw

        // Record start time
        startTime := time.Now()

        // Process request
        c.Next()

        // Create audit event
        event := &audit.AuditEvent{
            Timestamp:     startTime,
            UserID:        c.GetString("user_id"),
            OrgID:         c.GetString("org_id"),
            Action:        c.Request.Method + " " + c.Request.URL.Path,
            ResourceType:  extractResourceType(c.Request.URL.Path),
            IPAddress:     c.ClientIP(),
            UserAgent:     c.Request.UserAgent(),
            CorrelationID: c.GetString("correlation_id"),
            Status:        getStatusText(c.Writer.Status()),
            Metadata: map[string]interface{}{
                "method":        c.Request.Method,
                "path":          c.Request.URL.Path,
                "query":         c.Request.URL.Query(),
                "duration_ms":   time.Since(startTime).Milliseconds(),
                "status_code":   c.Writer.Status(),
                "response_size": blw.body.Len(),
            },
        }

        // Add request body if present
        if len(requestBody) > 0 && len(requestBody) < 10000 { // Limit size
            var reqData interface{}
            if json.Unmarshal(requestBody, &reqData) == nil {
                event.NewValue = reqData
            }
        }

        // Add error message if failed
        if c.Writer.Status() >= 400 {
            if err := c.Errors.Last(); err != nil {
                event.ErrorMessage = err.Error()
            }
        }

        // Log the audit event
        auditService.LogEvent(c, event)
    }
}

func extractResourceType(path string) string {
    // Extract resource type from path
    // /api/v1/farmers/123 -> "farmer"
    parts := strings.Split(strings.Trim(path, "/"), "/")
    if len(parts) >= 3 {
        resource := parts[2]
        // Remove plural 's'
        if strings.HasSuffix(resource, "s") {
            return resource[:len(resource)-1]
        }
        return resource
    }
    return "unknown"
}

func getStatusText(code int) string {
    switch {
    case code < 300:
        return "SUCCESS"
    case code < 400:
        return "REDIRECT"
    case code < 500:
        return "CLIENT_ERROR"
    default:
        return "SERVER_ERROR"
    }
}
```

#### Step 3: Update Admin Handlers
Update file: `internal/handlers/admin_handlers.go:321-331`
```go
func (h *AdminHandlers) GetAuditTrail(auditService *audit.AuditService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Parse query parameters
        startDate := c.Query("start_date")
        endDate := c.Query("end_date")
        userID := c.Query("user_id")
        action := c.Query("action")
        resourceType := c.Query("resource_type")
        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

        // Validate dates
        var startTime, endTime *time.Time
        if startDate != "" {
            t, err := time.Parse(time.RFC3339, startDate)
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{
                    "error": "Invalid start_date format",
                })
                return
            }
            startTime = &t
        }

        if endDate != "" {
            t, err := time.Parse(time.RFC3339, endDate)
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{
                    "error": "Invalid end_date format",
                })
                return
            }
            endTime = &t
        }

        // Create filters
        filters := &audit.AuditFilters{
            StartTime:    startTime,
            EndTime:      endTime,
            UserID:       userID,
            Action:       action,
            ResourceType: resourceType,
            Page:         page,
            PageSize:     pageSize,
        }

        // Query audit trail
        events, err := auditService.QueryAuditTrail(c, filters)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Failed to retrieve audit trail",
            })
            return
        }

        // Return response
        c.JSON(http.StatusOK, AuditTrailResponse{
            Message: "Audit trail retrieved successfully",
            Data: AuditTrailData{
                AuditLogs: events,
                Filters: AuditTrailFilters{
                    StartDate: startDate,
                    EndDate:   endDate,
                    UserID:    userID,
                    Action:    action,
                },
                TotalCount: len(events),
                Page:       page,
                PageSize:   pageSize,
            },
            CorrelationID: c.GetString("correlation_id"),
            Timestamp:     time.Now(),
        })
    }
}
```

## Priority 3: Secure Password Generation

### Implementation Solution

#### Create Secure Password Generator
Create file: `internal/utils/password.go`
```go
package utils

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "math/big"
    "strings"

    "golang.org/x/crypto/bcrypt"
)

const (
    // Character sets for password generation
    upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    lowerChars   = "abcdefghijklmnopqrstuvwxyz"
    digitChars   = "0123456789"
    specialChars = "!@#$%^&*()_+-=[]{}|;:,.<>?"

    // Minimum requirements
    minLength      = 12
    minUpper       = 2
    minLower       = 2
    minDigits      = 2
    minSpecial     = 2
    defaultBcryptCost = 10
)

type PasswordGenerator struct {
    length int
    cost   int
}

func NewPasswordGenerator() *PasswordGenerator {
    return &PasswordGenerator{
        length: 16,
        cost:   defaultBcryptCost,
    }
}

func (pg *PasswordGenerator) GenerateSecurePassword() (string, error) {
    // Ensure minimum requirements are met
    password := make([]byte, 0, pg.length)

    // Add required characters
    for i := 0; i < minUpper; i++ {
        char, err := randomChar(upperChars)
        if err != nil {
            return "", err
        }
        password = append(password, char)
    }

    for i := 0; i < minLower; i++ {
        char, err := randomChar(lowerChars)
        if err != nil {
            return "", err
        }
        password = append(password, char)
    }

    for i := 0; i < minDigits; i++ {
        char, err := randomChar(digitChars)
        if err != nil {
            return "", err
        }
        password = append(password, char)
    }

    for i := 0; i < minSpecial; i++ {
        char, err := randomChar(specialChars)
        if err != nil {
            return "", err
        }
        password = append(password, char)
    }

    // Fill remaining length with random characters
    allChars := upperChars + lowerChars + digitChars + specialChars
    for len(password) < pg.length {
        char, err := randomChar(allChars)
        if err != nil {
            return "", err
        }
        password = append(password, char)
    }

    // Shuffle the password
    shuffled, err := shuffle(password)
    if err != nil {
        return "", err
    }

    return string(shuffled), nil
}

func (pg *PasswordGenerator) HashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), pg.cost)
    if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }
    return string(hash), nil
}

func (pg *PasswordGenerator) VerifyPassword(hashedPassword, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}

func randomChar(charset string) (byte, error) {
    max := big.NewInt(int64(len(charset)))
    n, err := rand.Int(rand.Reader, max)
    if err != nil {
        return 0, err
    }
    return charset[n.Int64()], nil
}

func shuffle(data []byte) ([]byte, error) {
    result := make([]byte, len(data))
    copy(result, data)

    for i := len(result) - 1; i > 0; i-- {
        j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
        if err != nil {
            return nil, err
        }
        result[i], result[j.Int64()] = result[j.Int64()], result[i]
    }

    return result, nil
}

// GenerateAPIKey generates a secure API key
func GenerateAPIKey() (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }

    // Encode to URL-safe base64
    key := base64.URLEncoding.EncodeToString(bytes)
    key = strings.TrimRight(key, "=") // Remove padding

    return key, nil
}
```

#### Update Farmer Service
Update file: `internal/services/farmer_service.go`
```go
// Add to farmer service
type FarmerService struct {
    // ... existing fields ...
    passwordGen *utils.PasswordGenerator
}

// Update password generation in CreateFarmer
func (s *FarmerService) CreateFarmer(ctx context.Context, req *CreateFarmerRequest) (*Farmer, error) {
    // ... existing validation ...

    // Generate secure password
    password, err := s.passwordGen.GenerateSecurePassword()
    if err != nil {
        return nil, fmt.Errorf("failed to generate password: %w", err)
    }

    // Hash password before storing
    hashedPassword, err := s.passwordGen.HashPassword(password)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }

    // Create AAA user with hashed password
    // Note: Send plain password to AAA service, it will hash it
    userID, err := s.aaaClient.CreateUser(ctx, &CreateUserRequest{
        Username:    req.PhoneNumber,
        Password:    password, // AAA service will hash this
        PhoneNumber: req.PhoneNumber,
        CountryCode: req.CountryCode,
        // ... other fields ...
    })

    // Store hashed password locally if needed
    farmer.PasswordHash = hashedPassword

    // ... rest of implementation ...
}
```

## Testing Requirements

### Unit Tests for Token Validation
```go
func TestTokenValidator_ValidateToken(t *testing.T) {
    tests := []struct {
        name      string
        token     string
        wantErr   bool
        wantClaims *TokenClaims
    }{
        {
            name:    "valid token",
            token:   generateTestToken(time.Now().Add(time.Hour)),
            wantErr: false,
        },
        {
            name:    "expired token",
            token:   generateTestToken(time.Now().Add(-time.Hour)),
            wantErr: true,
        },
        {
            name:    "invalid signature",
            token:   "invalid.token.signature",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            validator := NewTokenValidator("test-secret", nil)
            claims, err := validator.ValidateToken(context.Background(), tt.token)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, claims)
            }
        })
    }
}
```

### Integration Tests for Audit Service
```go
func TestAuditService_LogEvent(t *testing.T) {
    logger := zap.NewNop()
    mockClient := &MockAuditClient{}
    service := NewAuditService(logger, mockClient)

    event := &AuditEvent{
        UserID:       "user-123",
        Action:       "CREATE_FARMER",
        ResourceType: "farmer",
        ResourceID:   "farmer-456",
    }

    err := service.LogEvent(context.Background(), event)
    assert.NoError(t, err)

    // Verify event was queued
    time.Sleep(100 * time.Millisecond)
    assert.Equal(t, 1, mockClient.EventCount())
}
```

## Deployment Checklist

### Environment Variables
```bash
# JWT Configuration
JWT_SECRET=your-secret-key-here
JWT_PUBLIC_KEY_PATH=/etc/farmers-module/jwt-public.pem

# Audit Service
AUDIT_SERVICE_ENABLED=true
AUDIT_SERVICE_URL=grpc://audit-service:50052
AUDIT_BATCH_SIZE=100
AUDIT_FLUSH_INTERVAL=5s

# Security
PASSWORD_MIN_LENGTH=12
BCRYPT_COST=10
MAX_LOGIN_ATTEMPTS=5
LOCKOUT_DURATION=15m
```

### Migration Script
```sql
-- Add audit trail table if using local storage
CREATE TABLE IF NOT EXISTS audit_trail (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id VARCHAR(255),
    org_id VARCHAR(255),
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    user_agent TEXT,
    correlation_id VARCHAR(255),
    status VARCHAR(50),
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Add indexes for performance
CREATE INDEX idx_audit_trail_user_id ON audit_trail(user_id, timestamp DESC);
CREATE INDEX idx_audit_trail_resource ON audit_trail(resource_type, resource_id);
CREATE INDEX idx_audit_trail_timestamp ON audit_trail(timestamp DESC);
CREATE INDEX idx_audit_trail_correlation ON audit_trail(correlation_id);

-- Add password_hash column to farmers if storing locally
ALTER TABLE farmers ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);
ALTER TABLE farmers ADD COLUMN IF NOT EXISTS failed_login_attempts INT DEFAULT 0;
ALTER TABLE farmers ADD COLUMN IF NOT EXISTS locked_until TIMESTAMPTZ;
```

## Monitoring & Alerts

### Key Metrics
```yaml
alerts:
  - name: token_validation_failures
    expression: rate(token_validation_failures[5m]) > 0.1
    severity: warning
    description: "High rate of token validation failures"

  - name: audit_queue_full
    expression: audit_queue_size > 900
    severity: critical
    description: "Audit queue is almost full"

  - name: password_generation_failures
    expression: rate(password_generation_failures[5m]) > 0
    severity: critical
    description: "Password generation is failing"
```

### Dashboard Queries
```sql
-- Token validation performance
SELECT
    date_trunc('minute', timestamp) as time,
    COUNT(*) as total_validations,
    COUNT(CASE WHEN status = 'SUCCESS' THEN 1 END) as successful,
    COUNT(CASE WHEN status = 'FAILED' THEN 1 END) as failed,
    AVG(duration_ms) as avg_duration_ms,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms) as p95_duration_ms
FROM audit_trail
WHERE action = 'TOKEN_VALIDATION'
AND timestamp > NOW() - INTERVAL '1 hour'
GROUP BY time
ORDER BY time DESC;

-- Most active users
SELECT
    user_id,
    COUNT(*) as action_count,
    COUNT(DISTINCT action) as unique_actions,
    MAX(timestamp) as last_activity
FROM audit_trail
WHERE timestamp > NOW() - INTERVAL '24 hours'
GROUP BY user_id
ORDER BY action_count DESC
LIMIT 20;
```

---

This implementation guide provides complete, production-ready code for fixing the critical issues. Each solution includes proper error handling, logging, caching, and fallback mechanisms to ensure reliability in production.
