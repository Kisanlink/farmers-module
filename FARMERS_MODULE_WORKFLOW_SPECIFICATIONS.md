# Farmers Module - Critical Workflows Technical Specifications

## Table of Contents

1. [Database Foundation](#database-foundation)
2. [Core Models & Repositories](#core-models--repositories)
3. [AAA Integration](#aaa-integration)
4. [Middleware](#middleware)
5. [Identity & Farm Workflows](#identity--farm-workflows)
6. [Admin & Health Workflows](#admin--health-workflows)

---

## Database Foundation

### GORM Connection Setup

**File**: `internal/db/db.go`

```go
// Database configuration matching kisanlink-db
type DatabaseConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    Name     string
    SSLMode  string
    MaxConns int
}

// PostGIS extension bootstrap
func SetupDatabase(postgresManager *db.PostgresManager) error {
    // Enable PostGIS extension
    gormDB.Exec(`CREATE EXTENSION IF NOT EXISTS postgis;`)

    // Create custom ENUMs
    createEnums(gormDB)

    // AutoMigrate models
    models := []interface{}{
        &fpo.FPORef{},
        &farmer.FarmerLink{},
        &farmer.Farmer{},
        &farm.Farm{},
        &crop_cycle.CropCycle{},
        &farm_activity.FarmActivity{},
    }

    // Post-migration setup
    setupPostMigration(gormDB)
}
```

### Database Schema Requirements

#### Custom ENUMs

```sql
-- Season enum
CREATE TYPE season AS ENUM ('RABI','KHARIF','ZAID','OTHER');

-- Cycle status enum
CREATE TYPE cycle_status AS ENUM ('PLANNED','ACTIVE','COMPLETED','CANCELLED');

-- Activity status enum
CREATE TYPE activity_status AS ENUM ('PLANNED','COMPLETED','CANCELLED');

-- Link status enum
CREATE TYPE link_status AS ENUM ('ACTIVE','INACTIVE');

-- Farmer status enum
CREATE TYPE farmer_status AS ENUM ('ACTIVE','INACTIVE','SUSPENDED');
```

#### PostGIS Setup

```sql
-- SRID check constraint
ALTER TABLE farms ADD CONSTRAINT farms_geometry_srid_check
    CHECK (ST_SRID(geometry) = 4326);

-- Geometry validity constraint
ALTER TABLE farms ADD CONSTRAINT farms_geometry_valid_check
    CHECK (ST_IsValid(geometry));

-- Computed area column
ALTER TABLE farms ADD COLUMN area_ha_computed NUMERIC(12,4)
    GENERATED ALWAYS AS (ST_Area(geometry::geometry)/10000.0) STORED;

-- GIST spatial index
CREATE INDEX farms_geometry_gist ON farms USING GIST (geometry::geometry);
```

---

## Core Models & Repositories

### FPORef Model

**File**: `internal/entities/fpo/fpo.go`

```go
type FPORef struct {
    base.BaseModel
    AAAOrgID    string            `json:"aaa_org_id" gorm:"type:varchar(255);not null;uniqueIndex"`
    Name        string            `json:"name" gorm:"type:varchar(255);not null"`
    Description string            `json:"description" gorm:"type:text"`
    Type        string            `json:"type" gorm:"type:varchar(100);not null"`
    Status      string            `json:"status" gorm:"type:varchar(50);not null;default:'ACTIVE'"`
    Metadata    map[string]string `json:"metadata" gorm:"type:jsonb;default:'{}'"`
}

func (f *FPORef) TableName() string { return "fpo_refs" }
```

### FarmerLink Model

**File**: `internal/entities/farmer/farmer.go`

```go
type FarmerLink struct {
    base.BaseModel
    AAAUserID        string  `json:"aaa_user_id" gorm:"type:varchar(255);not null"`
    AAAOrgID         string  `json:"aaa_org_id" gorm:"type:varchar(255);not null"`
    KisanSathiUserID *string `json:"kisan_sathi_user_id" gorm:"type:varchar(255)"`
    Status           string  `json:"status" gorm:"type:link_status;not null;default:'ACTIVE'"`
}

func (fl *FarmerLink) TableName() string { return "farmer_links" }
```

### Farm Model

**File**: `internal/entities/farm/farm.go`

```go
type Farm struct {
    base.BaseModel
    AAAFarmerUserID string            `json:"aaa_farmer_user_id" gorm:"type:varchar(255);not null"`
    AAAOrgID        string            `json:"aaa_org_id" gorm:"type:varchar(255);not null"`
    Geometry        interface{}       `json:"geometry" gorm:"type:geometry(POLYGON,4326)"`
    AreaHa          float64           `json:"area_ha" gorm:"type:numeric(12,4);not null"`
    Metadata        map[string]string `json:"metadata" gorm:"type:jsonb;default:'{}'"`
}

func (f *Farm) TableName() string { return "farms" }
```

### Repository Interfaces

**File**: `internal/repo/interfaces.go`

```go
type FarmerLinkRepository interface {
    Create(ctx context.Context, link *farmer.FarmerLink) error
    GetByUserAndOrg(ctx context.Context, userID, orgID string) (*farmer.FarmerLink, error)
    Update(ctx context.Context, link *farmer.FarmerLink) error
    Delete(ctx context.Context, userID, orgID string) error
    List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*farmer.FarmerLink, int64, error)
}

type FarmRepository interface {
    Create(ctx context.Context, farm *farm.Farm) error
    GetByID(ctx context.Context, id string) (*farm.Farm, error)
    Update(ctx context.Context, farm *farm.Farm) error
    Delete(ctx context.Context, id string) error
    ListByFarmer(ctx context.Context, farmerUserID, orgID string, page, pageSize int) ([]*farm.Farm, int64, error)
    ValidateGeometry(wkt string) error
}
```

---

## AAA Integration

### gRPC Client Setup

**File**: `internal/clients/aaa/aaa_client.go`

```go
type Client struct {
    conn        *grpc.ClientConn
    config      *config.Config
    userClient  proto.UserServiceV2Client
    authzClient proto.AuthorizationServiceClient
}

// Core AAA operations
func (c *Client) CheckPermission(ctx context.Context, subject, resource, action, object, orgID string) (bool, error)
func (c *Client) ResolveUser(ctx context.Context, userID string) (map[string]interface{}, error)
func (c *Client) EnsureRolesAndPermissions(ctx context.Context) error
```

### Permission Mapping

```go
// Resource-Action mapping for workflows
var WorkflowPermissions = map[string]struct{
    Resource string
    Action   string
}{
    "LinkFarmerToFPO":    {Resource: "farmer_link", Action: "create"},
    "CreateFarm":         {Resource: "farm", Action: "create"},
    "ListFarms":          {Resource: "farm", Action: "read"},
    "SeedPermissions":    {Resource: "admin", Action: "manage"},
}
```

### Startup Seeding

**File**: `internal/services/aaa_service.go`

```go
func (s *AAAServiceImpl) SeedRolesAndPermissions(ctx context.Context) error {
    resources := []string{"farmer_link", "farm", "crop_cycle", "farm_activity", "admin"}
    actions := []string{"create", "read", "update", "delete", "manage"}
    roles := []string{"farmer", "kisan_sathi", "fpo_admin", "system_admin"}

    // Idempotent seeding logic
    for _, resource := range resources {
        for _, action := range actions {
            // Create resource-action permissions
            // Bind to appropriate roles
        }
    }
}
```

---

## Middleware

### Authentication Middleware

**File**: `internal/middleware/auth.go`

```go
func AuthenticationMiddleware(aaaClient *aaa.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
            c.Abort()
            return
        }

        // Extract Bearer token
        if strings.HasPrefix(token, "Bearer ") {
            token = token[7:]
        }

        // Validate with AAA service
        userInfo, err := aaaClient.ValidateToken(c.Request.Context(), token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }

        // Set context values
        c.Set("aaa_subject", userInfo["user_id"])
        c.Set("aaa_org", userInfo["org_id"])
        c.Next()
    }
}
```

### Authorization Middleware

**File**: `internal/middleware/authz.go`

```go
func AuthorizationMiddleware(aaaClient *aaa.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        subject := c.GetString("aaa_subject")
        org := c.GetString("aaa_org")

        // Map route to resource-action
        resource, action := mapRouteToPermission(c.Request.Method, c.FullPath())

        // Check permission with AAA
        allowed, err := aaaClient.CheckPermission(
            c.Request.Context(), subject, resource, action, "", org)

        if err != nil || !allowed {
            c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
            c.Abort()
            return
        }

        c.Next()
    }
}

func mapRouteToPermission(method, path string) (resource, action string) {
    switch {
    case strings.Contains(path, "/farmer-links") && method == "POST":
        return "farmer_link", "create"
    case strings.Contains(path, "/farms") && method == "POST":
        return "farm", "create"
    case strings.Contains(path, "/farms") && method == "GET":
        return "farm", "read"
    default:
        return "unknown", "unknown"
    }
}
```

### Audit Middleware

**File**: `internal/middleware/audit.go`

```go
func AuditMiddleware(logger interfaces.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        // Process request
        c.Next()

        // Log audit trail
        logger.Info("API Request",
            zap.String("subject", c.GetString("aaa_subject")),
            zap.String("org", c.GetString("aaa_org")),
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("duration", time.Since(start)),
            zap.String("request_id", c.GetString("request_id")),
        )
    }
}
```

---

## Identity & Farm Workflows

### W1: Link Farmer to FPO

#### API Specification

- **Endpoint**: `POST /api/v1/identity/farmer-links`
- **Method**: POST
- **Authentication**: Required (Bearer token)
- **Authorization**: `farmer_link:create` permission

#### Request Schema

```json
{
  "request_id": "req_123456789",
  "aaa_user_id": "user_abc123",
  "aaa_org_id": "org_xyz789",
  "kisan_sathi_user_id": "user_def456"
}
```

#### Response Schema

```json
{
  "request_id": "req_123456789",
  "status": "success",
  "message": "Farmer linked to FPO successfully",
  "data": {
    "id": "fl_987654321",
    "aaa_user_id": "user_abc123",
    "aaa_org_id": "org_xyz789",
    "kisan_sathi_user_id": "user_def456",
    "status": "ACTIVE",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### Service Implementation

**File**: `internal/services/farmer_linkage_service.go`

```go
func (s *FarmerLinkageServiceImpl) LinkFarmerToFPO(ctx context.Context, req interface{}) error {
    linkReq := req.(*requests.LinkFarmerRequest)

    // Validate AAA user exists
    _, err := s.aaaClient.GetUser(ctx, linkReq.AAAUserID)
    if err != nil {
        return fmt.Errorf("user not found in AAA: %w", err)
    }

    // Validate AAA org exists
    _, err = s.aaaClient.VerifyOrganization(ctx, linkReq.AAAOrgID)
    if err != nil {
        return fmt.Errorf("organization not found in AAA: %w", err)
    }

    // Check if link already exists
    existing, _ := s.repo.GetByUserAndOrg(ctx, linkReq.AAAUserID, linkReq.AAAOrgID)
    if existing != nil {
        return fmt.Errorf("farmer already linked to FPO")
    }

    // Create farmer link
    link := &farmer.FarmerLink{
        AAAUserID:        linkReq.AAAUserID,
        AAAOrgID:         linkReq.AAAOrgID,
        KisanSathiUserID: linkReq.KisanSathiUserID,
        Status:           "ACTIVE",
    }

    return s.repo.Create(ctx, link)
}
```

#### Error Codes

- `400`: Invalid request format
- `401`: Unauthorized (missing/invalid token)
- `403`: Forbidden (insufficient permissions)
- `404`: User or organization not found
- `409`: Farmer already linked to FPO
- `500`: Internal server error

#### Example cURL

```bash
curl -X POST http://localhost:8000/api/v1/identity/farmer-links \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "req_123456789",
    "aaa_user_id": "user_abc123",
    "aaa_org_id": "org_xyz789",
    "kisan_sathi_user_id": "user_def456"
  }'
```

### W6: Create Farm

#### API Specification

- **Endpoint**: `POST /api/v1/farms`
- **Method**: POST
- **Authentication**: Required (Bearer token)
- **Authorization**: `farm:create` permission

#### Request Schema

```json
{
  "request_id": "req_farm_001",
  "aaa_farmer_user_id": "user_farmer123",
  "aaa_org_id": "org_fpo456",
  "geometry_wkt": "POLYGON((77.1025 28.7041, 77.1035 28.7041, 77.1035 28.7051, 77.1025 28.7051, 77.1025 28.7041))",
  "metadata": {
    "soil_type": "loamy",
    "irrigation": "drip"
  }
}
```

#### Response Schema

```json
{
  "request_id": "req_farm_001",
  "status": "success",
  "message": "Farm created successfully",
  "data": {
    "id": "farm_789012345",
    "aaa_farmer_user_id": "user_farmer123",
    "aaa_org_id": "org_fpo456",
    "area_ha": 2.5,
    "area_ha_computed": 2.5,
    "geometry": "POLYGON((77.1025 28.7041, 77.1035 28.7041, 77.1035 28.7051, 77.1025 28.7051, 77.1025 28.7041))",
    "metadata": {
      "soil_type": "loamy",
      "irrigation": "drip"
    },
    "created_at": "2024-01-15T11:00:00Z"
  }
}
```

#### Service Implementation

**File**: `internal/services/farm_service.go`

```go
func (s *FarmServiceImpl) CreateFarm(ctx context.Context, req interface{}) (interface{}, error) {
    farmReq := req.(*requests.CreateFarmRequest)

    // Validate WKT geometry
    if err := s.validateWKT(farmReq.GeometryWKT); err != nil {
        return nil, fmt.Errorf("invalid geometry: %w", err)
    }

    // Validate farmer is linked to FPO
    link, err := s.farmerLinkRepo.GetByUserAndOrg(ctx, farmReq.AAAFarmerUserID, farmReq.AAAOrgID)
    if err != nil || link == nil {
        return nil, fmt.Errorf("farmer not linked to FPO")
    }

    // Create farm with PostGIS geometry
    farm := &farm.Farm{
        AAAFarmerUserID: farmReq.AAAFarmerUserID,
        AAAOrgID:        farmReq.AAAOrgID,
        Geometry:        farmReq.GeometryWKT,
        Metadata:        farmReq.Metadata,
    }

    if err := s.repo.Create(ctx, farm); err != nil {
        return nil, fmt.Errorf("failed to create farm: %w", err)
    }

    return responses.NewFarmResponse(farm, "Farm created successfully"), nil
}

func (s *FarmServiceImpl) validateWKT(wkt string) error {
    // Validate WKT format and SRID
    var isValid bool
    err := s.db.Raw("SELECT ST_IsValid(ST_GeomFromText(?, 4326))", wkt).Scan(&isValid).Error
    if err != nil || !isValid {
        return fmt.Errorf("invalid WKT geometry")
    }
    return nil
}
```

#### Database Requirements

```sql
-- Farms table with PostGIS
CREATE TABLE farms (
    id VARCHAR(255) PRIMARY KEY,
    aaa_farmer_user_id VARCHAR(255) NOT NULL,
    aaa_org_id VARCHAR(255) NOT NULL,
    geometry GEOMETRY(POLYGON, 4326) NOT NULL,
    area_ha NUMERIC(12,4) NOT NULL,
    area_ha_computed NUMERIC(12,4) GENERATED ALWAYS AS (ST_Area(geometry::geometry)/10000.0) STORED,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,

    CONSTRAINT farms_geometry_srid_check CHECK (ST_SRID(geometry) = 4326),
    CONSTRAINT farms_geometry_valid_check CHECK (ST_IsValid(geometry))
);

-- Indexes
CREATE INDEX farms_geometry_gist ON farms USING GIST (geometry);
CREATE INDEX farms_farmer_id_idx ON farms (aaa_farmer_user_id);
CREATE INDEX farms_fpo_id_idx ON farms (aaa_org_id);
```

#### Validation Rules

- WKT must be valid POLYGON geometry
- SRID must be 4326 (WGS84)
- Farmer must be linked to the specified FPO
- Area computed automatically by PostGIS
- Geometry must pass ST_IsValid check

#### Example cURL

```bash
curl -X POST http://localhost:8000/api/v1/farms \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "req_farm_001",
    "aaa_farmer_user_id": "user_farmer123",
    "aaa_org_id": "org_fpo456",
    "geometry_wkt": "POLYGON((77.1025 28.7041, 77.1035 28.7041, 77.1035 28.7051, 77.1025 28.7051, 77.1025 28.7041))",
    "metadata": {
      "soil_type": "loamy",
      "irrigation": "drip"
    }
  }'
```

### W9: List Farms

#### API Specification

- **Endpoint**: `GET /api/v1/farms`
- **Method**: GET
- **Authentication**: Required (Bearer token)
- **Authorization**: `farm:read` permission

#### Query Parameters

- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 10, max: 100)
- `aaa_farmer_user_id`: Filter by farmer ID
- `aaa_org_id`: Filter by organization ID
- `min_area_ha`: Minimum area filter
- `max_area_ha`: Maximum area filter

#### Response Schema

```json
{
  "request_id": "req_list_farms_001",
  "status": "success",
  "message": "Farms retrieved successfully",
  "data": [
    {
      "id": "farm_789012345",
      "aaa_farmer_user_id": "user_farmer123",
      "aaa_org_id": "org_fpo456",
      "area_ha": 2.5,
      "area_ha_computed": 2.5,
      "metadata": {
        "soil_type": "loamy",
        "irrigation": "drip"
      },
      "created_at": "2024-01-15T11:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total_count": 25,
    "total_pages": 3,
    "has_next": true,
    "has_prev": false
  }
}
```

#### Service Implementation

```go
func (s *FarmServiceImpl) ListFarms(ctx context.Context, req interface{}) (interface{}, error) {
    listReq := req.(*requests.ListFarmsRequest)

    // Build filters
    filters := make(map[string]interface{})
    if listReq.AAAFarmerUserID != "" {
        filters["aaa_farmer_user_id"] = listReq.AAAFarmerUserID
    }
    if listReq.AAAOrgID != "" {
        filters["aaa_org_id"] = listReq.AAAOrgID
    }
    if listReq.MinAreaHa > 0 {
        filters["min_area_ha"] = listReq.MinAreaHa
    }
    if listReq.MaxAreaHa > 0 {
        filters["max_area_ha"] = listReq.MaxAreaHa
    }

    // Get farms with pagination
    farms, total, err := s.repo.List(ctx, filters, listReq.Page, listReq.PageSize)
    if err != nil {
        return nil, fmt.Errorf("failed to list farms: %w", err)
    }

    return responses.NewFarmListResponse(farms, listReq.Page, listReq.PageSize, total), nil
}
```

---

## Admin & Health Workflows

### W18: Seed Roles and Permissions

#### API Specification

- **Endpoint**: `POST /api/v1/admin/seed-permissions`
- **Method**: POST
- **Authentication**: Required (Bearer token)
- **Authorization**: `admin:manage` permission

#### Request Schema

```json
{
  "request_id": "req_seed_001",
  "force_reseed": false
}
```

#### Response Schema

```json
{
  "request_id": "req_seed_001",
  "status": "success",
  "message": "Roles and permissions seeded successfully",
  "data": {
    "resources_created": 5,
    "actions_created": 20,
    "roles_created": 4,
    "bindings_created": 15
  }
}
```

#### Service Implementation

```go
func (s *AAAServiceImpl) SeedRolesAndPermissions(ctx context.Context) error {
    // Define resources and actions
    resources := []string{"farmer_link", "farm", "crop_cycle", "farm_activity", "admin"}
    actions := []string{"create", "read", "update", "delete", "manage"}
    roles := []string{"farmer", "kisan_sathi", "fpo_admin", "system_admin"}

    // Create resources (idempotent)
    for _, resource := range resources {
        if err := s.client.CreateResource(ctx, resource); err != nil {
            log.Printf("Resource %s may already exist: %v", resource, err)
        }
    }

    // Create role bindings
    roleBindings := map[string][]string{
        "farmer":       {"farm:create", "farm:read", "farm:update"},
        "kisan_sathi":  {"farm:read", "farmer_link:read", "crop_cycle:read"},
        "fpo_admin":    {"*:*"}, // All permissions within org
        "system_admin": {"admin:manage"},
    }

    for role, permissions := range roleBindings {
        for _, permission := range permissions {
            if err := s.client.BindRolePermission(ctx, role, permission); err != nil {
                log.Printf("Failed to bind %s to %s: %v", permission, role, err)
            }
        }
    }

    return nil
}
```

### Health Check Endpoint

#### API Specification

- **Endpoint**: `GET /health`
- **Method**: GET
- **Authentication**: Not required

#### Response Schema

```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T12:00:00Z",
  "checks": {
    "database": {
      "status": "healthy",
      "response_time_ms": 5
    },
    "aaa_service": {
      "status": "healthy",
      "response_time_ms": 12
    },
    "postgis": {
      "status": "healthy",
      "version": "3.4.0"
    }
  }
}
```

#### Implementation

```go
func (h *HealthHandler) HealthCheck(c *gin.Context) {
    ctx := c.Request.Context()
    checks := make(map[string]interface{})

    // Database check
    start := time.Now()
    if err := h.db.Raw("SELECT 1").Error; err != nil {
        checks["database"] = map[string]interface{}{
            "status": "unhealthy",
            "error":  err.Error(),
        }
    } else {
        checks["database"] = map[string]interface{}{
            "status":           "healthy",
            "response_time_ms": time.Since(start).Milliseconds(),
        }
    }

    // AAA service check
    start = time.Now()
    if err := h.aaaClient.HealthCheck(ctx); err != nil {
        checks["aaa_service"] = map[string]interface{}{
            "status": "unhealthy",
            "error":  err.Error(),
        }
    } else {
        checks["aaa_service"] = map[string]interface{}{
            "status":           "healthy",
            "response_time_ms": time.Since(start).Milliseconds(),
        }
    }

    // PostGIS check
    var version string
    if err := h.db.Raw("SELECT PostGIS_Version()").Scan(&version).Error; err != nil {
        checks["postgis"] = map[string]interface{}{
            "status": "unhealthy",
            "error":  err.Error(),
        }
    } else {
        checks["postgis"] = map[string]interface{}{
            "status":  "healthy",
            "version": version,
        }
    }

    // Overall status
    status := "healthy"
    for _, check := range checks {
        if checkMap, ok := check.(map[string]interface{}); ok {
            if checkMap["status"] == "unhealthy" {
                status = "unhealthy"
                break
            }
        }
    }

    response := map[string]interface{}{
        "status":    status,
        "timestamp": time.Now().UTC().Format(time.RFC3339),
        "checks":    checks,
    }

    statusCode := http.StatusOK
    if status == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    }

    c.JSON(statusCode, response)
}
```

---

## Integration Testing

### Test Database Setup

```go
func setupTestDB() *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    // Auto migrate test models
    db.AutoMigrate(&farmer.FarmerLink{}, &farm.Farm{})

    return db
}
```

### Integration Test Example

```go
func TestLinkFarmerToFPO_Integration(t *testing.T) {
    // Setup
    db := setupTestDB()
    mockAAA := &MockAAAClient{}
    service := NewFarmerLinkageService(db, mockAAA)

    // Mock AAA responses
    mockAAA.On("GetUser", mock.Anything, "user123").Return(map[string]interface{}{
        "id": "user123", "status": "active",
    }, nil)

    // Test request
    req := &requests.LinkFarmerRequest{
        AAAUserID: "user123",
        AAAOrgID:  "org456",
    }

    // Execute
    err := service.LinkFarmerToFPO(context.Background(), req)

    // Assert
    assert.NoError(t, err)

    // Verify database state
    var link farmer.FarmerLink
    db.Where("aaa_user_id = ? AND aaa_org_id = ?", "user123", "org456").First(&link)
    assert.Equal(t, "ACTIVE", link.Status)
}
```

---

## Build and Test Commands

### Makefile Additions

```makefile
# Test commands
test-unit:
	go test ./internal/... -v

test-integration:
	go test ./tests/integration/... -v -tags=integration

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# Database commands
migrate-up:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go down

# Development workflow
dev-test: build test-unit docs
	@echo "✅ Development testing complete"

commit-ready: test-unit test-integration lint
	@echo "✅ Ready for commit"
```

This specification provides comprehensive technical details for implementing the critical workflows of the farmers-module service, following the established patterns and architecture while ensuring proper AAA integration, spatial data handling, and robust error handling.
