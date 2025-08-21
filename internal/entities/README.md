# Base Request and Response Models

This directory contains the base request and response models for the farmers module, providing a consistent structure for all API operations.

## Structure

```
entities/
├── requests/
│   ├── base.go              # Base request models
│   ├── farmer/
│   │   └── farmer.go        # Farmer-specific request models
│   └── farm/
│       └── farm.go          # Farm-specific request models
├── responses/
│   ├── base.go              # Base response models
│   ├── farmer/
│   │   └── farmer.go        # Farmer-specific response models
│   └── farm/
│       └── farm.go          # Farm-specific response models
└── README.md                 # This file
```

## Base Models

### BaseRequest
Common fields for all API requests:
- `RequestID`: Unique identifier for request tracking
- `Timestamp`: When the request was made
- `UserID`: ID of the user making the request
- `OrgID`: ID of the organization
- `Metadata`: Additional key-value pairs
- `RequestType`: Type of request being made

### BaseResponse
Common fields for all API responses:
- `RequestID`: Matches the request ID for tracking
- `Timestamp`: When the response was generated
- `Status`: Success/error status
- `Message`: Human-readable message
- `ErrorCode`: Error code if applicable
- `Metadata`: Additional key-value pairs
- `RequestType`: Type of response

### PaginationRequest/PaginationResponse
Handles pagination for list operations:
- `Page`: Current page number
- `PageSize`: Number of items per page
- `TotalPages`: Total number of pages
- `TotalCount`: Total number of items
- `HasNext/HasPrev`: Navigation indicators

### FilterRequest
Extends pagination with filtering capabilities:
- `Filters`: Map of field-value filters
- `SortBy`: Field to sort by
- `SortDir`: Sort direction (asc/desc)

## Entity-Specific Models

### Farmer Models
- **Requests**: Create, Update, Delete, Get, List operations
- **Responses**: Single farmer, list of farmers, farmer links
- **Features**: Profile data, address information, preferences

### Farm Models
- **Requests**: CRUD operations, spatial queries, overlap checks
- **Responses**: Farm data, geometry information, overlap results
- **Features**: Geographic boundaries, area calculations

## Usage Examples

### Creating a Request
```go
// Create a new farmer request
req := farmer.NewCreateFarmerRequest()
req.SetUserContext("user123", "org456")
req.SetRequestType("create_farmer")
req.AAAUserID = "user123"
req.AAAOrgID = "org456"
req.Profile.FirstName = "John"
req.Profile.LastName = "Doe"
```

### Creating a Response
```go
// Create a success response
resp := farmer.NewFarmerResponse(farmerData, "Farmer created successfully")
resp.SetRequestID(requestID)
resp.SetStatus("success")
```

### Using Pagination
```go
// Create a paginated request
req := requests.NewPaginationRequest(1, 20)
req.SetUserContext("user123", "org456")

// Create a paginated response
resp := responses.NewListResponse(data, 1, 20, 150)
```

## Validation

All request models include validation tags using the `validate` package:
- `required`: Field must be present
- `min/max`: Numeric constraints
- `oneof`: Enumeration constraints

## Extending

To add new entities:

1. Create a new subdirectory under `requests/` and `responses/`
2. Define entity-specific request/response types
3. Extend the appropriate base models
4. Add constructor functions with `New*` prefix
5. Implement `SetRequestID` methods for tracking

## Benefits

- **Consistency**: All APIs follow the same structure
- **Tracking**: Request ID correlation for debugging
- **Pagination**: Standardized list operations
- **Filtering**: Flexible query capabilities
- **Validation**: Built-in request validation
- **Extensibility**: Easy to add new entities
