# Subscription & Services API Strategy

## Executive Summary

This document outlines the strategy for implementing admin-gated subscription APIs that manage the product and services suite offered to farmers within the KisanLink ecosystem.

## Problem Statement

### Current Gaps
1. No centralized system to manage subscription plans and tiers
2. No catalog of products/services offered to farmers
3. No admin interfaces to configure offerings
4. No subscription lifecycle management (activation, renewal, cancellation)
5. No usage tracking and entitlement enforcement
6. No pricing and billing integration hooks

### Business Requirements
- **FPO Admin Control**: FPO administrators need to subscribe farmers to various service tiers
- **Product Catalog**: Centralized catalog of services (advisory, inputs, marketplace access, etc.)
- **Subscription Management**: Create, activate, suspend, renew subscriptions
- **Entitlement System**: Check farmer access to specific features/services
- **Usage Tracking**: Monitor service utilization
- **Multi-tenant**: FPO-level and platform-level subscriptions

## Domain Model

### Core Entities

#### 1. Service Catalog
```go
// Service represents a product/service offering
type Service struct {
    base.BaseModel

    // Core Fields
    Name               string          `json:"name" gorm:"type:varchar(255);not null;uniqueIndex"`
    DisplayName        string          `json:"display_name" gorm:"type:varchar(255);not null"`
    Description        string          `json:"description" gorm:"type:text"`
    Category           ServiceCategory `json:"category" gorm:"type:service_category;not null"`

    // Service Configuration
    IsActive           bool            `json:"is_active" gorm:"default:true"`
    RequiresApproval   bool            `json:"requires_approval" gorm:"default:false"`
    IsPremium          bool            `json:"is_premium" gorm:"default:false"`

    // Entitlements/Features
    Features           JSONB           `json:"features" gorm:"type:jsonb;default:'{}'"`
    Limits             JSONB           `json:"limits" gorm:"type:jsonb;default:'{}'"`

    // Metadata
    Icon               *string         `json:"icon"`
    Documentation      *string         `json:"documentation"`
    TermsOfService     *string         `json:"terms_of_service"`
    Metadata           JSONB           `json:"metadata" gorm:"type:jsonb;default:'{}'"`
}

// ServiceCategory represents types of services
type ServiceCategory string

const (
    CategoryAdvisory    ServiceCategory = "ADVISORY"      // Crop advisory, expert consultation
    CategoryInputs      ServiceCategory = "INPUTS"        // Seeds, fertilizers, pesticides
    CategoryMarketplace ServiceCategory = "MARKETPLACE"   // Buyer connections, trading
    CategoryFinance     ServiceCategory = "FINANCE"       // Credit, insurance, loans
    CategoryLogistics   ServiceCategory = "LOGISTICS"     // Transportation, storage
    CategoryAnalytics   ServiceCategory = "ANALYTICS"     // Farm insights, reports
    CategoryTraining    ServiceCategory = "TRAINING"      // Educational content
    CategoryWeather     ServiceCategory = "WEATHER"       // Weather forecasts, alerts
    CategoryOther       ServiceCategory = "OTHER"
)
```

#### 2. Subscription Plans
```go
// SubscriptionPlan represents a subscription tier/package
type SubscriptionPlan struct {
    base.BaseModel

    // Plan Identification
    Name               string             `json:"name" gorm:"type:varchar(255);not null;uniqueIndex"`
    DisplayName        string             `json:"display_name" gorm:"type:varchar(255);not null"`
    Description        string             `json:"description" gorm:"type:text"`

    // Plan Type
    PlanType           PlanType           `json:"plan_type" gorm:"type:plan_type;not null"`
    Tier               string             `json:"tier" gorm:"type:varchar(50)"` // Basic, Premium, Enterprise

    // Pricing
    PricingModel       PricingModel       `json:"pricing_model" gorm:"type:pricing_model;not null"`
    Price              float64            `json:"price" gorm:"type:numeric(12,2)"`
    Currency           string             `json:"currency" gorm:"type:varchar(10);default:'INR'"`
    BillingCycle       BillingCycle       `json:"billing_cycle" gorm:"type:billing_cycle"`

    // Trial & Discounts
    TrialDays          int                `json:"trial_days" gorm:"default:0"`
    SetupFee           float64            `json:"setup_fee" gorm:"type:numeric(12,2);default:0"`

    // Status
    IsActive           bool               `json:"is_active" gorm:"default:true"`
    IsPublic           bool               `json:"is_public" gorm:"default:true"`

    // Included Services (many-to-many)
    Services           []Service          `json:"services,omitempty" gorm:"many2many:plan_services"`

    // Entitlements/Limits
    Entitlements       JSONB              `json:"entitlements" gorm:"type:jsonb;default:'{}'"`
    UsageLimits        JSONB              `json:"usage_limits" gorm:"type:jsonb;default:'{}'"`

    // Metadata
    Metadata           JSONB              `json:"metadata" gorm:"type:jsonb;default:'{}'"`
}

type PlanType string
const (
    PlanTypeFarmer     PlanType = "FARMER"      // Individual farmer subscription
    PlanTypeFPO        PlanType = "FPO"         // FPO organization subscription
    PlanTypeBulk       PlanType = "BULK"        // Bulk subscription for multiple farmers
)

type PricingModel string
const (
    PricingFree        PricingModel = "FREE"
    PricingFixed       PricingModel = "FIXED"
    PricingUsageBased  PricingModel = "USAGE_BASED"
    PricingTiered      PricingModel = "TIERED"
)

type BillingCycle string
const (
    BillingOneTime     BillingCycle = "ONE_TIME"
    BillingMonthly     BillingCycle = "MONTHLY"
    BillingQuarterly   BillingCycle = "QUARTERLY"
    BillingYearly      BillingCycle = "YEARLY"
    BillingSeasonal    BillingCycle = "SEASONAL"  // Kharif/Rabi specific
)
```

#### 3. Farmer Subscriptions
```go
// FarmerSubscription represents an active subscription
type FarmerSubscription struct {
    base.BaseModel

    // Subscriber
    FarmerID           string             `json:"farmer_id" gorm:"type:varchar(255);not null;index"`
    AAAUserID          string             `json:"aaa_user_id" gorm:"type:varchar(255);not null;index"`
    AAAOrgID           string             `json:"aaa_org_id" gorm:"type:varchar(255);not null;index"`

    // Subscription Details
    PlanID             string             `json:"plan_id" gorm:"type:varchar(255);not null;index"`
    Plan               *SubscriptionPlan  `json:"plan,omitempty" gorm:"foreignKey:PlanID"`

    // Lifecycle
    Status             SubscriptionStatus `json:"status" gorm:"type:subscription_status;not null"`
    StartDate          time.Time          `json:"start_date" gorm:"not null"`
    EndDate            *time.Time         `json:"end_date"`
    TrialEndDate       *time.Time         `json:"trial_end_date"`
    RenewalDate        *time.Time         `json:"renewal_date"`

    // Auto-renewal
    AutoRenew          bool               `json:"auto_renew" gorm:"default:false"`
    CancellationDate   *time.Time         `json:"cancellation_date"`
    CancellationReason *string            `json:"cancellation_reason" gorm:"type:text"`

    // Billing
    BillingAmount      float64            `json:"billing_amount" gorm:"type:numeric(12,2)"`
    Currency           string             `json:"currency" gorm:"type:varchar(10);default:'INR'"`
    PaymentStatus      PaymentStatus      `json:"payment_status" gorm:"type:payment_status"`

    // Attribution
    SubscribedBy       string             `json:"subscribed_by" gorm:"type:varchar(255)"` // Admin user ID
    SubscribedByRole   string             `json:"subscribed_by_role" gorm:"type:varchar(100)"`

    // Usage Tracking
    UsageStats         JSONB              `json:"usage_stats" gorm:"type:jsonb;default:'{}'"`
    Metadata           JSONB              `json:"metadata" gorm:"type:jsonb;default:'{}'"`

    // Relationships
    Farmer             *farmer.Farmer     `json:"farmer,omitempty" gorm:"foreignKey:FarmerID"`
}

type SubscriptionStatus string
const (
    StatusTrial        SubscriptionStatus = "TRIAL"
    StatusActive       SubscriptionStatus = "ACTIVE"
    StatusSuspended    SubscriptionStatus = "SUSPENDED"
    StatusCancelled    SubscriptionStatus = "CANCELLED"
    StatusExpired      SubscriptionStatus = "EXPIRED"
    StatusPending      SubscriptionStatus = "PENDING_APPROVAL"
)

type PaymentStatus string
const (
    PaymentPending     PaymentStatus = "PENDING"
    PaymentPaid        PaymentStatus = "PAID"
    PaymentFailed      PaymentStatus = "FAILED"
    PaymentRefunded    PaymentStatus = "REFUNDED"
)
```

#### 4. Service Usage Tracking
```go
// ServiceUsage tracks farmer service utilization
type ServiceUsage struct {
    base.BaseModel

    SubscriptionID     string             `json:"subscription_id" gorm:"type:varchar(255);not null;index"`
    ServiceID          string             `json:"service_id" gorm:"type:varchar(255);not null;index"`
    FarmerID           string             `json:"farmer_id" gorm:"type:varchar(255);not null;index"`

    // Usage Metrics
    UsageCount         int                `json:"usage_count" gorm:"default:0"`
    UsageDate          time.Time          `json:"usage_date" gorm:"index"`
    UsageDetails       JSONB              `json:"usage_details" gorm:"type:jsonb"`

    // Relationships
    Subscription       *FarmerSubscription `json:"subscription,omitempty" gorm:"foreignKey:SubscriptionID"`
    Service            *Service            `json:"service,omitempty" gorm:"foreignKey:ServiceID"`
}
```

## API Design

### Admin Service Management APIs

#### 1. Service Catalog Management
```yaml
# Create Service (Platform Admin Only)
POST /api/v1/admin/services
Permission: admin.service_create

# Update Service
PUT /api/v1/admin/services/:id
Permission: admin.service_update

# List All Services
GET /api/v1/admin/services
Permission: admin.service_list

# Get Service Details
GET /api/v1/admin/services/:id
Permission: admin.service_read

# Archive Service
DELETE /api/v1/admin/services/:id
Permission: admin.service_delete

# Toggle Service Activation
PUT /api/v1/admin/services/:id/toggle
Permission: admin.service_update
```

#### 2. Subscription Plan Management
```yaml
# Create Subscription Plan (Platform Admin)
POST /api/v1/admin/plans
Permission: admin.plan_create

# Update Plan
PUT /api/v1/admin/plans/:id
Permission: admin.plan_update

# List Plans
GET /api/v1/admin/plans
Permission: admin.plan_list

# Get Plan Details
GET /api/v1/admin/plans/:id
Permission: admin.plan_read

# Assign Services to Plan
POST /api/v1/admin/plans/:id/services
Permission: admin.plan_update

# Remove Service from Plan
DELETE /api/v1/admin/plans/:id/services/:service_id
Permission: admin.plan_update

# Archive Plan
DELETE /api/v1/admin/plans/:id
Permission: admin.plan_delete
```

#### 3. Farmer Subscription Management (FPO Admin)
```yaml
# Subscribe Farmer to Plan
POST /api/v1/admin/subscriptions
Permission: subscription.create
Body: {
  "farmer_id": "FMRR...",
  "plan_id": "PLAN...",
  "start_date": "2025-01-01",
  "auto_renew": true
}

# List Subscriptions (with filters)
GET /api/v1/admin/subscriptions?farmer_id=...&status=...&org_id=...
Permission: subscription.list

# Get Subscription Details
GET /api/v1/admin/subscriptions/:id
Permission: subscription.read

# Update Subscription (change plan, extend date)
PUT /api/v1/admin/subscriptions/:id
Permission: subscription.update

# Suspend Subscription
PUT /api/v1/admin/subscriptions/:id/suspend
Permission: subscription.suspend

# Activate Suspended Subscription
PUT /api/v1/admin/subscriptions/:id/activate
Permission: subscription.activate

# Cancel Subscription
PUT /api/v1/admin/subscriptions/:id/cancel
Permission: subscription.cancel

# Renew Subscription
POST /api/v1/admin/subscriptions/:id/renew
Permission: subscription.renew
```

#### 4. Bulk Subscription Operations
```yaml
# Bulk Subscribe Farmers (FPO Admin)
POST /api/v1/admin/subscriptions/bulk
Permission: subscription.bulk_create
Body: {
  "plan_id": "PLAN...",
  "farmer_ids": ["FMRR001", "FMRR002", ...],
  "start_date": "2025-01-01",
  "duration_months": 12
}

# Bulk Update Subscriptions
PUT /api/v1/admin/subscriptions/bulk
Permission: subscription.bulk_update

# Bulk Cancel Subscriptions
PUT /api/v1/admin/subscriptions/bulk/cancel
Permission: subscription.bulk_cancel
```

### Farmer-Facing APIs (Read-Only)

```yaml
# Get My Subscriptions
GET /api/v1/subscriptions/me
Permission: subscription.read_own

# Get My Active Services
GET /api/v1/services/available
Permission: service.list_available

# Check Service Entitlement
GET /api/v1/entitlements/check?service=...
Permission: entitlement.check

# Get Usage History
GET /api/v1/usage/me
Permission: usage.read_own
```

## Authorization Strategy

### Role-Based Access Control (RBAC)

#### Platform Admin Roles
```yaml
platform_admin:
  permissions:
    - admin.service_*          # Full service catalog management
    - admin.plan_*             # Full plan management
    - subscription.*           # View all subscriptions
    - usage.*                  # View all usage
    - report.*                 # Analytics and reports
```

#### FPO Admin Roles
```yaml
fpo_admin:
  permissions:
    - service.list             # View available services
    - plan.list                # View available plans
    - subscription.create      # Subscribe farmers in their FPO
    - subscription.read        # View subscriptions in their FPO
    - subscription.update      # Update subscriptions in their FPO
    - subscription.suspend     # Suspend subscriptions
    - subscription.cancel      # Cancel subscriptions
    - subscription.renew       # Renew subscriptions
    - subscription.bulk_*      # Bulk operations for their FPO
    - usage.read              # View usage for their FPO farmers
```

#### Kisan Sathi Roles
```yaml
kisan_sathi:
  permissions:
    - service.list_available   # View available services
    - subscription.read        # View subscriptions of assigned farmers
    - usage.read              # View usage of assigned farmers
```

#### Farmer Roles
```yaml
farmer:
  permissions:
    - service.list_available   # View services they have access to
    - subscription.read_own    # View their own subscriptions
    - entitlement.check       # Check their entitlements
    - usage.read_own          # View their own usage
```

### Implementation with AAA Service

```go
// Permission definitions in internal/auth/permissions.go
var RoutePermissionMap = map[string]Permission{
    // Admin - Service Management
    "POST /api/v1/admin/services":              {Resource: "admin", Action: "service_create"},
    "PUT /api/v1/admin/services/:id":           {Resource: "admin", Action: "service_update"},
    "GET /api/v1/admin/services":               {Resource: "admin", Action: "service_list"},
    "GET /api/v1/admin/services/:id":           {Resource: "admin", Action: "service_read"},
    "DELETE /api/v1/admin/services/:id":        {Resource: "admin", Action: "service_delete"},

    // Admin - Plan Management
    "POST /api/v1/admin/plans":                 {Resource: "admin", Action: "plan_create"},
    "PUT /api/v1/admin/plans/:id":              {Resource: "admin", Action: "plan_update"},
    "GET /api/v1/admin/plans":                  {Resource: "admin", Action: "plan_list"},
    "GET /api/v1/admin/plans/:id":              {Resource: "admin", Action: "plan_read"},
    "DELETE /api/v1/admin/plans/:id":           {Resource: "admin", Action: "plan_delete"},

    // Admin - Subscription Management (FPO-scoped)
    "POST /api/v1/admin/subscriptions":         {Resource: "subscription", Action: "create"},
    "GET /api/v1/admin/subscriptions":          {Resource: "subscription", Action: "list"},
    "GET /api/v1/admin/subscriptions/:id":      {Resource: "subscription", Action: "read"},
    "PUT /api/v1/admin/subscriptions/:id":      {Resource: "subscription", Action: "update"},
    "PUT /api/v1/admin/subscriptions/:id/suspend": {Resource: "subscription", Action: "suspend"},
    "PUT /api/v1/admin/subscriptions/:id/cancel": {Resource: "subscription", Action: "cancel"},

    // Farmer-facing
    "GET /api/v1/subscriptions/me":             {Resource: "subscription", Action: "read_own"},
    "GET /api/v1/services/available":           {Resource: "service", Action: "list_available"},
    "GET /api/v1/entitlements/check":           {Resource: "entitlement", Action: "check"},
}
```

### Org-Scoped Authorization

```go
// Middleware to enforce FPO-level data isolation
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
    // Extract admin user from context
    adminUserID := c.GetString("aaa_user_id")
    adminOrgID := c.GetString("aaa_org_id")

    var req CreateSubscriptionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Verify farmer belongs to admin's organization
    farmer, err := h.farmerService.GetByID(c.Request.Context(), req.FarmerID)
    if err != nil {
        c.JSON(404, gin.H{"error": "Farmer not found"})
        return
    }

    if farmer.AAAOrgID != adminOrgID {
        c.JSON(403, gin.H{"error": "Cannot subscribe farmer from different organization"})
        return
    }

    // Check permission with AAA service
    allowed, err := h.aaaClient.CheckPermission(c.Request.Context(), &aaapb.CheckPermissionRequest{
        Subject:  adminUserID,
        Resource: "subscription",
        Action:   "create",
        Object:   req.FarmerID,
        OrgID:    adminOrgID,
    })

    if err != nil || !allowed.Allowed {
        c.JSON(403, gin.H{"error": "Permission denied"})
        return
    }

    // Proceed with subscription creation
    subscription, err := h.subscriptionService.CreateSubscription(c.Request.Context(), req)
    // ...
}
```

## Entitlement Checking System

### Usage Pattern
```go
// Check if farmer has access to a service
func (s *SubscriptionService) CheckEntitlement(ctx context.Context, farmerID, serviceKey string) (bool, error) {
    // Get active subscription
    subscription, err := s.repo.GetActiveSubscription(ctx, farmerID)
    if err != nil {
        return false, err
    }

    if subscription == nil || subscription.Status != StatusActive {
        return false, nil
    }

    // Check if subscription's plan includes the service
    plan, err := s.planRepo.GetByID(ctx, subscription.PlanID)
    if err != nil {
        return false, err
    }

    // Check if service is in plan
    hasService := false
    for _, svc := range plan.Services {
        if svc.Name == serviceKey {
            hasService = true
            break
        }
    }

    if !hasService {
        return false, nil
    }

    // Check usage limits (if any)
    if plan.UsageLimits != nil {
        // Check monthly/daily limits
        usage, err := s.usageRepo.GetCurrentUsage(ctx, subscription.ID, serviceKey)
        if err != nil {
            return false, err
        }

        // Validate against limits
        if exceeded := s.checkUsageLimits(plan.UsageLimits, usage); exceeded {
            return false, nil
        }
    }

    return true, nil
}
```

### Integration in Existing Services
```go
// Example: Farm Advisory Service
func (h *AdvisoryHandler) GetAdvice(c *gin.Context) {
    farmerID := c.Param("farmer_id")

    // Check entitlement
    hasAccess, err := h.subscriptionService.CheckEntitlement(c.Request.Context(), farmerID, "advisory")
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to check entitlement"})
        return
    }

    if !hasAccess {
        c.JSON(403, gin.H{
            "error": "Service not available",
            "message": "Please upgrade your subscription to access advisory services",
            "upgrade_url": "/api/v1/plans?service=advisory",
        })
        return
    }

    // Track usage
    go h.subscriptionService.TrackUsage(context.Background(), farmerID, "advisory", map[string]interface{}{
        "action": "get_advice",
        "timestamp": time.Now(),
    })

    // Proceed with service logic
    // ...
}
```

## Database Schema

### Migration Strategy
```sql
-- 006_subscription_services.sql

-- Create ENUMs
CREATE TYPE service_category AS ENUM (
    'ADVISORY', 'INPUTS', 'MARKETPLACE', 'FINANCE',
    'LOGISTICS', 'ANALYTICS', 'TRAINING', 'WEATHER', 'OTHER'
);

CREATE TYPE plan_type AS ENUM ('FARMER', 'FPO', 'BULK');
CREATE TYPE pricing_model AS ENUM ('FREE', 'FIXED', 'USAGE_BASED', 'TIERED');
CREATE TYPE billing_cycle AS ENUM ('ONE_TIME', 'MONTHLY', 'QUARTERLY', 'YEARLY', 'SEASONAL');
CREATE TYPE subscription_status AS ENUM ('TRIAL', 'ACTIVE', 'SUSPENDED', 'CANCELLED', 'EXPIRED', 'PENDING_APPROVAL');
CREATE TYPE payment_status AS ENUM ('PENDING', 'PAID', 'FAILED', 'REFUNDED');

-- Services table
CREATE TABLE services (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    category service_category NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    requires_approval BOOLEAN DEFAULT FALSE,
    is_premium BOOLEAN DEFAULT FALSE,
    features JSONB DEFAULT '{}',
    limits JSONB DEFAULT '{}',
    icon VARCHAR(500),
    documentation TEXT,
    terms_of_service TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255)
);

-- Subscription plans table
CREATE TABLE subscription_plans (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    plan_type plan_type NOT NULL,
    tier VARCHAR(50),
    pricing_model pricing_model NOT NULL,
    price NUMERIC(12,2),
    currency VARCHAR(10) DEFAULT 'INR',
    billing_cycle billing_cycle,
    trial_days INTEGER DEFAULT 0,
    setup_fee NUMERIC(12,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    is_public BOOLEAN DEFAULT TRUE,
    entitlements JSONB DEFAULT '{}',
    usage_limits JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255)
);

-- Plan-Service junction table
CREATE TABLE plan_services (
    plan_id VARCHAR(255) NOT NULL REFERENCES subscription_plans(id),
    service_id VARCHAR(255) NOT NULL REFERENCES services(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (plan_id, service_id)
);

-- Farmer subscriptions table
CREATE TABLE farmer_subscriptions (
    id VARCHAR(255) PRIMARY KEY,
    farmer_id VARCHAR(255) NOT NULL REFERENCES farmers(id),
    aaa_user_id VARCHAR(255) NOT NULL,
    aaa_org_id VARCHAR(255) NOT NULL,
    plan_id VARCHAR(255) NOT NULL REFERENCES subscription_plans(id),
    status subscription_status NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP,
    trial_end_date TIMESTAMP,
    renewal_date TIMESTAMP,
    auto_renew BOOLEAN DEFAULT FALSE,
    cancellation_date TIMESTAMP,
    cancellation_reason TEXT,
    billing_amount NUMERIC(12,2),
    currency VARCHAR(10) DEFAULT 'INR',
    payment_status payment_status,
    subscribed_by VARCHAR(255),
    subscribed_by_role VARCHAR(100),
    usage_stats JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255)
);

-- Service usage tracking table
CREATE TABLE service_usage (
    id VARCHAR(255) PRIMARY KEY,
    subscription_id VARCHAR(255) NOT NULL REFERENCES farmer_subscriptions(id),
    service_id VARCHAR(255) NOT NULL REFERENCES services(id),
    farmer_id VARCHAR(255) NOT NULL REFERENCES farmers(id),
    usage_count INTEGER DEFAULT 0,
    usage_date TIMESTAMP NOT NULL,
    usage_details JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_services_category ON services(category);
CREATE INDEX idx_services_is_active ON services(is_active);

CREATE INDEX idx_plans_type ON subscription_plans(plan_type);
CREATE INDEX idx_plans_tier ON subscription_plans(tier);
CREATE INDEX idx_plans_is_active ON subscription_plans(is_active);

CREATE INDEX idx_farmer_subs_farmer_id ON farmer_subscriptions(farmer_id);
CREATE INDEX idx_farmer_subs_org_id ON farmer_subscriptions(aaa_org_id);
CREATE INDEX idx_farmer_subs_plan_id ON farmer_subscriptions(plan_id);
CREATE INDEX idx_farmer_subs_status ON farmer_subscriptions(status);
CREATE INDEX idx_farmer_subs_start_date ON farmer_subscriptions(start_date);
CREATE INDEX idx_farmer_subs_end_date ON farmer_subscriptions(end_date);

CREATE INDEX idx_usage_subscription_id ON service_usage(subscription_id);
CREATE INDEX idx_usage_service_id ON service_usage(service_id);
CREATE INDEX idx_usage_farmer_id ON service_usage(farmer_id);
CREATE INDEX idx_usage_date ON service_usage(usage_date);
```

## Implementation Roadmap

### Phase 1: Foundation (Week 1-2)
**Goal**: Core domain models and database setup

**Tasks**:
1. Create entity models (Service, SubscriptionPlan, FarmerSubscription, ServiceUsage)
2. Create database migration with ENUMs and tables
3. Setup repositories using BaseFilterableRepository pattern
4. Add permission mappings for admin routes
5. Create basic CRUD services
6. Unit tests for domain logic

**Deliverables**:
- `internal/entities/service/` (entities)
- `internal/entities/subscription/` (entities)
- `internal/repo/service/`, `internal/repo/subscription/` (repositories)
- `internal/db/migrations/006_subscription_services.sql`
- Unit tests

### Phase 2: Admin APIs - Service & Plan Management (Week 3)
**Goal**: Platform admin can manage service catalog and plans

**Tasks**:
1. Implement ServiceService with CRUD operations
2. Implement PlanService with CRUD operations
3. Create admin handlers for services
4. Create admin handlers for plans
5. Implement plan-service association APIs
6. Add validation and error handling
7. Integration tests

**Deliverables**:
- `internal/services/service_service.go`
- `internal/services/plan_service.go`
- `internal/handlers/admin_service_handler.go`
- `internal/handlers/admin_plan_handler.go`
- `internal/routes/admin_subscription_routes.go`
- Swagger documentation
- Integration tests

### Phase 3: Subscription Management (Week 4)
**Goal**: FPO admins can subscribe farmers to plans

**Tasks**:
1. Implement SubscriptionService
2. Create subscription lifecycle methods (activate, suspend, cancel, renew)
3. Implement org-scoped authorization checks
4. Create admin subscription handlers
5. Add bulk subscription operations
6. Integration tests with AAA service

**Deliverables**:
- `internal/services/subscription_service.go`
- `internal/handlers/admin_subscription_handler.go`
- Org-scoped middleware
- Integration tests

### Phase 4: Entitlement System (Week 5)
**Goal**: Service access control and usage tracking

**Tasks**:
1. Implement entitlement checking logic
2. Create usage tracking service
3. Add middleware for service gating
4. Implement usage limits enforcement
5. Create farmer-facing entitlement APIs
6. Integration tests

**Deliverables**:
- `internal/services/entitlement_service.go`
- `internal/services/usage_service.go`
- `middleware/entitlement.go`
- Farmer APIs
- Integration tests

### Phase 5: Analytics & Reporting (Week 6)
**Goal**: Usage analytics and subscription reports

**Tasks**:
1. Implement subscription analytics
2. Create usage reports
3. Add admin dashboards endpoints
4. Revenue/billing reports (preparation for billing integration)
5. Export capabilities

**Deliverables**:
- `internal/services/subscription_analytics_service.go`
- `internal/handlers/subscription_report_handler.go`
- Admin dashboard APIs

### Phase 6: Testing & Documentation (Week 7)
**Goal**: Comprehensive testing and documentation

**Tasks**:
1. End-to-end testing
2. Performance testing (bulk operations)
3. Security audit
4. API documentation (Swagger)
5. Admin user guide
6. Developer documentation

**Deliverables**:
- Complete test coverage
- Swagger docs
- User guides
- Architecture documentation

## Integration Points

### With Existing Modules

#### 1. Farmer Service Integration
```go
// Add subscription info to farmer responses
type FarmerProfileData struct {
    // ... existing fields
    ActiveSubscription *SubscriptionInfo `json:"active_subscription,omitempty"`
    AvailableServices  []string          `json:"available_services,omitempty"`
}
```

#### 2. Advisory/Input Services
```go
// Gate services with entitlement checks
func (h *AdvisoryHandler) GetCropAdvice(c *gin.Context) {
    // Check entitlement first
    if !h.entitlementService.CheckAccess(farmerID, "advisory") {
        return c.JSON(403, gin.H{"error": "Subscription required"})
    }
    // ... service logic
}
```

#### 3. Reporting Integration
```go
// Include subscription status in farmer reports
func (s *ReportService) GenerateFarmerPortfolio(farmerID string) (*Report, error) {
    subscription, _ := s.subscriptionService.GetActiveSubscription(farmerID)
    // Include in report
}
```

### External Integrations (Future)

1. **Payment Gateway**: Razorpay, PayU for subscription payments
2. **Billing System**: Invoice generation, payment tracking
3. **CRM**: Sync subscription data for customer management
4. **Analytics**: Export to data warehouse for BI
5. **Notification Service**: Renewal reminders, expiry alerts

## Security Considerations

### 1. Authorization
- ✅ Multi-level: Platform admin > FPO admin > Farmer
- ✅ Org-scoped data isolation (FPO admins can only manage their farmers)
- ✅ AAA service integration for permission checks
- ✅ Resource-level authorization (check farmer belongs to org)

### 2. Data Protection
- Encryption at rest for sensitive subscription data
- Audit logging for all subscription changes
- PII protection (payment info, billing details)

### 3. Rate Limiting
- Protect bulk subscription APIs from abuse
- Usage tracking to prevent API quota exhaustion

### 4. Input Validation
- Strict validation for plan pricing, dates, limits
- Sanitize JSONB fields (features, entitlements)

## Success Metrics

### Business Metrics
1. **Adoption Rate**: % of farmers with active subscriptions
2. **Plan Distribution**: Breakdown by tier (Basic/Premium/Enterprise)
3. **Revenue**: Monthly recurring revenue from subscriptions
4. **Churn Rate**: Subscription cancellations
5. **Service Utilization**: Usage per service

### Technical Metrics
1. **API Performance**: P95 response time < 200ms
2. **Availability**: 99.9% uptime
3. **Entitlement Check Latency**: < 50ms
4. **Bulk Operation Throughput**: > 1000 subscriptions/minute

## Risk Mitigation

### Technical Risks
1. **Performance**: Cache entitlement checks, use read replicas
2. **Data Consistency**: Use transactions for subscription lifecycle
3. **Migration**: Blue-green deployment for database changes

### Business Risks
1. **Pricing Complexity**: Start simple, iterate based on feedback
2. **Admin Adoption**: Provide admin training and onboarding
3. **Service Integration**: Phased rollout, feature flags

## Next Steps

1. **Review & Approval**: Get stakeholder sign-off on strategy
2. **Team Assignment**: Allocate developers to phases
3. **Environment Setup**: Staging environment for testing
4. **Kick-off Phase 1**: Begin with foundation work

## Appendix

### Example Service Definitions
```json
{
  "services": [
    {
      "name": "crop_advisory",
      "display_name": "Crop Advisory",
      "category": "ADVISORY",
      "features": {
        "expert_consultation": true,
        "crop_recommendations": true,
        "pest_disease_alerts": true
      },
      "limits": {
        "consultations_per_month": 5
      }
    },
    {
      "name": "market_access",
      "display_name": "Market Access",
      "category": "MARKETPLACE",
      "features": {
        "buyer_connections": true,
        "price_discovery": true,
        "contract_farming": true
      }
    }
  ]
}
```

### Example Subscription Plan
```json
{
  "plan": {
    "name": "premium_farmer",
    "display_name": "Premium Farmer Plan",
    "plan_type": "FARMER",
    "tier": "Premium",
    "pricing_model": "FIXED",
    "price": 999.00,
    "currency": "INR",
    "billing_cycle": "YEARLY",
    "trial_days": 30,
    "services": ["crop_advisory", "market_access", "weather_alerts"],
    "entitlements": {
      "priority_support": true,
      "expert_consultations": 10,
      "data_export": true
    },
    "usage_limits": {
      "api_calls_per_day": 1000
    }
  }
}
```
