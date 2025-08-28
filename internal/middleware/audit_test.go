package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kisanlink/farmers-module/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventEmitter for testing
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

func TestAuditMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		path        string
		method      string
		userContext *auth.UserContext
		orgContext  *auth.OrgContext
		statusCode  int
		setupMocks  func(*MockLogger, *MockEventEmitter)
		expectAudit bool
	}{
		{
			name:   "Successful request with full context",
			path:   "/api/v1/farmers",
			method: "GET",
			userContext: &auth.UserContext{
				AAAUserID: "user123",
				Username:  "testuser",
			},
			orgContext: &auth.OrgContext{
				AAAOrgID: "org123",
			},
			statusCode: http.StatusOK,
			setupMocks: func(logger *MockLogger, emitter *MockEventEmitter) {
				logger.On("Info", "Audit Event", mock.MatchedBy(func(fields []interface{}) bool {
					// Verify that audit fields are present
					return len(fields) > 0
				})).Return()
				emitter.On("EmitAuditEvent", mock.AnythingOfType("AuditEvent")).Return(nil)
			},
			expectAudit: true,
		},
		{
			name:       "Request without user context",
			path:       "/api/v1/health",
			method:     "GET",
			statusCode: http.StatusOK,
			setupMocks: func(logger *MockLogger, emitter *MockEventEmitter) {
				logger.On("Info", "Audit Event", mock.MatchedBy(func(fields []interface{}) bool {
					return len(fields) > 0
				})).Return()
				emitter.On("EmitAuditEvent", mock.AnythingOfType("AuditEvent")).Return(nil)
			},
			expectAudit: true,
		},
		{
			name:   "Failed request",
			path:   "/api/v1/farmers",
			method: "POST",
			userContext: &auth.UserContext{
				AAAUserID: "user123",
				Username:  "testuser",
			},
			statusCode: http.StatusBadRequest,
			setupMocks: func(logger *MockLogger, emitter *MockEventEmitter) {
				logger.On("Info", "Audit Event", mock.MatchedBy(func(fields []interface{}) bool {
					return len(fields) > 0
				})).Return()
				emitter.On("EmitAuditEvent", mock.AnythingOfType("AuditEvent")).Return(nil)
			},
			expectAudit: true,
		},
		{
			name:   "Request with query parameters",
			path:   "/api/v1/farmers?limit=10&offset=0",
			method: "GET",
			userContext: &auth.UserContext{
				AAAUserID: "user123",
				Username:  "testuser",
			},
			statusCode: http.StatusOK,
			setupMocks: func(logger *MockLogger, emitter *MockEventEmitter) {
				logger.On("Info", "Audit Event", mock.MatchedBy(func(fields []interface{}) bool {
					return len(fields) > 0
				})).Return()
				emitter.On("EmitAuditEvent", mock.AnythingOfType("AuditEvent")).Return(nil)
			},
			expectAudit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockLogger := &MockLogger{}
			mockEmitter := &MockEventEmitter{}
			tt.setupMocks(mockLogger, mockEmitter)

			// Setup Gin router
			router := gin.New()
			router.Use(RequestID())

			// Add middleware that sets context
			router.Use(func(c *gin.Context) {
				if tt.userContext != nil {
					c.Set("user_context", tt.userContext)
				}
				if tt.orgContext != nil {
					c.Set("org_context", tt.orgContext)
				}
				if mockEmitter != nil {
					c.Set("event_emitter", mockEmitter)
				}
				c.Next()
			})

			router.Use(AuditMiddleware(mockLogger))

			// Add test routes
			router.GET("/api/v1/farmers", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})
			router.POST("/api/v1/farmers", func(c *gin.Context) {
				c.JSON(tt.statusCode, gin.H{"message": "response"})
			})
			router.GET("/api/v1/health", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "healthy"})
			})

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("User-Agent", "test-agent")

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.statusCode, w.Code)

			// Verify mocks
			mockLogger.AssertExpectations(t)
			if tt.expectAudit && mockEmitter != nil {
				mockEmitter.AssertExpectations(t)
			}
		})
	}
}

func TestAuditEventStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	mockEmitter := &MockEventEmitter{}

	// Capture the audit event
	var capturedEvent AuditEvent
	mockLogger.On("Info", "Audit Event", mock.MatchedBy(func(fields []interface{}) bool {
		// Extract audit event details from the log fields
		return true
	})).Return()

	mockEmitter.On("EmitAuditEvent", mock.MatchedBy(func(event interface{}) bool {
		if auditEvent, ok := event.(AuditEvent); ok {
			capturedEvent = auditEvent
			return true
		}
		return false
	})).Return(nil)

	// Setup Gin router
	router := gin.New()
	router.Use(RequestID())

	// Add context middleware
	router.Use(func(c *gin.Context) {
		userContext := &auth.UserContext{
			AAAUserID: "user123",
			Username:  "testuser",
		}
		orgContext := &auth.OrgContext{
			AAAOrgID: "org123",
		}
		c.Set("user_context", userContext)
		c.Set("org_context", orgContext)
		c.Set("event_emitter", mockEmitter)
		c.Next()
	})

	router.Use(AuditMiddleware(mockLogger))

	// Add test route
	router.GET("/api/v1/farmers/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/farmers/123", nil)
	req.Header.Set("User-Agent", "test-agent")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify audit event structure
	assert.Equal(t, http.StatusOK, w.Code)
	mockLogger.AssertExpectations(t)
	mockEmitter.AssertExpectations(t)

	// Verify captured event has expected fields
	assert.Equal(t, "user123", capturedEvent.Subject)
	assert.Equal(t, "testuser", capturedEvent.Username)
	assert.Equal(t, "org123", capturedEvent.Organization)
	assert.Equal(t, "farmer", capturedEvent.Resource)
	assert.Equal(t, "read", capturedEvent.Action)
	assert.Equal(t, "123", capturedEvent.Object)
	assert.Equal(t, "GET", capturedEvent.Method)
	assert.Equal(t, "/api/v1/farmers/123", capturedEvent.Path)
	assert.Equal(t, http.StatusOK, capturedEvent.StatusCode)
	assert.True(t, capturedEvent.Success)
	assert.Equal(t, "test-agent", capturedEvent.UserAgent)
	assert.NotEmpty(t, capturedEvent.RequestID)
	assert.NotZero(t, capturedEvent.Duration)
}

func TestAuditMiddleware_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}

	// Setup expectation for error logging
	mockLogger.On("Info", "Audit Event", mock.MatchedBy(func(fields []interface{}) bool {
		return true
	})).Return()

	// Setup Gin router
	router := gin.New()
	router.Use(RequestID())
	router.Use(AuditMiddleware(mockLogger))

	// Add route that returns an error
	router.GET("/api/v1/error", func(c *gin.Context) {
		c.Error(assert.AnError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "test error"})
	})

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/error", nil)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Verify mocks
	mockLogger.AssertExpectations(t)
}

func TestAuditMiddleware_Performance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}

	// Setup expectation for audit logging
	mockLogger.On("Info", "Audit Event", mock.MatchedBy(func(fields []interface{}) bool {
		return true
	})).Return()

	// Setup Gin router
	router := gin.New()
	router.Use(RequestID())
	router.Use(AuditMiddleware(mockLogger))

	// Add route with artificial delay
	router.GET("/api/v1/slow", func(c *gin.Context) {
		time.Sleep(100 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"message": "slow response"})
	})

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/slow", nil)

	// Record response
	start := time.Now()
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	duration := time.Since(start)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, duration >= 100*time.Millisecond, "Duration should be at least 100ms")

	// Verify mocks
	mockLogger.AssertExpectations(t)
}
