package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestErrorHandlerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupRoute     func(*gin.Engine, *MockLogger)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Panic recovery",
			setupRoute: func(router *gin.Engine, logger *MockLogger) {
				logger.On("Error", "Panic recovered", mock.Anything).Return()
				router.GET("/panic", func(c *gin.Context) {
					panic("test panic")
				})
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "internal_server_error",
		},
		{
			name: "GORM record not found error",
			setupRoute: func(router *gin.Engine, logger *MockLogger) {
				logger.On("Error", "Request error", mock.Anything).Return()
				router.GET("/not-found", func(c *gin.Context) {
					c.Error(gorm.ErrRecordNotFound)
				})
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "not_found",
		},
		{
			name: "Validation error",
			setupRoute: func(router *gin.Engine, logger *MockLogger) {
				logger.On("Error", "Request error", mock.Anything).Return()
				router.GET("/validation", func(c *gin.Context) {
					c.Error(errors.New("validation failed: field is required"))
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation_error",
		},
		{
			name: "Duplicate key error",
			setupRoute: func(router *gin.Engine, logger *MockLogger) {
				logger.On("Error", "Request error", mock.Anything).Return()
				router.GET("/duplicate", func(c *gin.Context) {
					c.Error(errors.New("duplicate key value violates unique constraint"))
				})
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "conflict",
		},
		{
			name: "Foreign key error",
			setupRoute: func(router *gin.Engine, logger *MockLogger) {
				logger.On("Error", "Request error", mock.Anything).Return()
				router.GET("/foreign-key", func(c *gin.Context) {
					c.Error(errors.New("violates foreign key constraint fk_users"))
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad_request",
		},
		{
			name: "Constraint error",
			setupRoute: func(router *gin.Engine, logger *MockLogger) {
				logger.On("Error", "Request error", mock.Anything).Return()
				router.GET("/constraint", func(c *gin.Context) {
					c.Error(errors.New("violates check constraint"))
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad_request",
		},
		{
			name: "Authentication error",
			setupRoute: func(router *gin.Engine, logger *MockLogger) {
				logger.On("Error", "Request error", mock.Anything).Return()
				router.GET("/auth", func(c *gin.Context) {
					c.Error(errors.New("authentication failed: invalid token"))
				})
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name: "Authorization error",
			setupRoute: func(router *gin.Engine, logger *MockLogger) {
				logger.On("Error", "Request error", mock.Anything).Return()
				router.GET("/authz", func(c *gin.Context) {
					c.Error(errors.New("permission denied for resource"))
				})
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "forbidden",
		},
		{
			name: "Service unavailable error",
			setupRoute: func(router *gin.Engine, logger *MockLogger) {
				logger.On("Error", "Request error", mock.Anything).Return()
				router.GET("/unavailable", func(c *gin.Context) {
					c.Error(errors.New("AAA service unavailable"))
				})
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  "service_unavailable",
		},
		{
			name: "Generic error",
			setupRoute: func(router *gin.Engine, logger *MockLogger) {
				logger.On("Error", "Request error", mock.Anything).Return()
				router.GET("/generic", func(c *gin.Context) {
					c.Error(errors.New("some unexpected error"))
				})
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "internal_server_error",
		},
		{
			name: "No error - successful request",
			setupRoute: func(router *gin.Engine, logger *MockLogger) {
				router.GET("/success", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "success"})
				})
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock logger
			mockLogger := &MockLogger{}

			// Setup Gin router
			router := gin.New()
			router.Use(RequestID())
			router.Use(ErrorHandlerMiddleware(mockLogger))

			// Setup route
			tt.setupRoute(router, mockLogger)

			// Create request
			var reqPath string
			if tt.name == "Panic recovery" {
				reqPath = "/panic"
			} else if tt.name == "GORM record not found error" {
				reqPath = "/not-found"
			} else if tt.name == "Validation error" {
				reqPath = "/validation"
			} else if tt.name == "Duplicate key error" {
				reqPath = "/duplicate"
			} else if tt.name == "Foreign key error" {
				reqPath = "/foreign-key"
			} else if tt.name == "Constraint error" {
				reqPath = "/constraint"
			} else if tt.name == "Authentication error" {
				reqPath = "/auth"
			} else if tt.name == "Authorization error" {
				reqPath = "/authz"
			} else if tt.name == "Service unavailable error" {
				reqPath = "/unavailable"
			} else if tt.name == "Generic error" {
				reqPath = "/generic"
			} else if tt.name == "No error - successful request" {
				reqPath = "/success"
			} else {
				reqPath = "/test"
			}

			req := httptest.NewRequest("GET", reqPath, nil)

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				// Parse response body to check error structure
				assert.Contains(t, w.Body.String(), tt.expectedError)
				assert.Contains(t, w.Body.String(), "correlation_id")
			}

			// Verify mocks
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestErrorTypeChecking(t *testing.T) {
	tests := []struct {
		name     string
		error    error
		checker  func(error) bool
		expected bool
	}{
		{
			name:     "Validation error - positive",
			error:    errors.New("validation failed: field is required"),
			checker:  isValidationError,
			expected: true,
		},
		{
			name:     "Validation error - negative",
			error:    errors.New("some other error"),
			checker:  isValidationError,
			expected: false,
		},
		{
			name:     "Duplicate key error - positive",
			error:    errors.New("duplicate key value violates unique constraint"),
			checker:  isDuplicateKeyError,
			expected: true,
		},
		{
			name:     "Duplicate key error - already exists",
			error:    errors.New("resource already exists"),
			checker:  isDuplicateKeyError,
			expected: true,
		},
		{
			name:     "Foreign key error - positive",
			error:    errors.New("violates foreign key constraint fk_users"),
			checker:  isForeignKeyError,
			expected: true,
		},
		{
			name:     "Constraint error - positive",
			error:    errors.New("violates check constraint"),
			checker:  isConstraintError,
			expected: true,
		},
		{
			name:     "Authentication error - positive",
			error:    errors.New("invalid token provided"),
			checker:  isAuthenticationError,
			expected: true,
		},
		{
			name:     "Authorization error - positive",
			error:    errors.New("permission denied for this resource"),
			checker:  isAuthorizationError,
			expected: true,
		},
		{
			name:     "Service unavailable error - positive",
			error:    errors.New("AAA service connection refused"),
			checker:  isServiceUnavailableError,
			expected: true,
		},
		{
			name:     "Service unavailable error - timeout",
			error:    errors.New("request timeout occurred"),
			checker:  isServiceUnavailableError,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checker(tt.error)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorHandlerMiddleware_MultipleErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	mockLogger.On("Error", "Request error", mock.Anything).Return()

	// Setup Gin router
	router := gin.New()
	router.Use(RequestID())
	router.Use(ErrorHandlerMiddleware(mockLogger))

	// Add route that adds multiple errors
	router.GET("/multiple-errors", func(c *gin.Context) {
		c.Error(errors.New("first error"))
		c.Error(errors.New("validation failed: second error"))
		c.JSON(http.StatusBadRequest, gin.H{"message": "response"})
	})

	// Create request
	req := httptest.NewRequest("GET", "/multiple-errors", nil)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response - should handle the last error
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "validation_error")

	// Verify mocks
	mockLogger.AssertExpectations(t)
}

func TestErrorHandlerMiddleware_NoRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	mockLogger.On("Error", "Request error", mock.Anything).Return()

	// Setup Gin router without RequestID middleware
	router := gin.New()
	router.Use(ErrorHandlerMiddleware(mockLogger))

	// Add route that causes an error
	router.GET("/error", func(c *gin.Context) {
		c.Error(errors.New("test error"))
	})

	// Create request
	req := httptest.NewRequest("GET", "/error", nil)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "unknown") // Should use "unknown" as correlation ID

	// Verify mocks
	mockLogger.AssertExpectations(t)
}
