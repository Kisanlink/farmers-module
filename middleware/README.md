# Middleware Configuration

This directory contains middleware configurations for the Farmers Module API.

## CORS Configuration

The CORS (Cross-Origin Resource Sharing) middleware is configured to allow all requests from any origin.

### Available Middleware Functions

#### `CORSMiddleware()`
- Allows all origins (`*`)
- Supports all HTTP methods: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
- Includes comprehensive headers support
- Credentials are disabled (required when using `*` for origins)

#### `CORSMiddlewareWithCredentials(allowedOrigins []string)`
- Allows specific origins with credentials support
- Use this when you need to send cookies or authentication headers
- Default origins: `http://localhost:3000`, `http://localhost:5173`, `https://farmers.kisanlink.in`

#### `SetupMiddlewares(router *gin.Engine)`
- Applies CORS and logging middleware
- Use this for basic setup

#### `SetupMiddlewaresWithAuth(router *gin.Engine)`
- Applies CORS, logging, and authentication middleware
- Use this when authentication is required

#### `SetupMiddlewaresWithCredentials(router *gin.Engine, allowedOrigins []string)`
- Applies CORS with credentials, logging, and other middlewares
- Use this when you need credentials support

### Usage

```go
// Basic setup (allows all requests)
router := gin.Default()
middleware.SetupMiddlewares(router)

// With authentication
router := gin.Default()
middleware.SetupMiddlewaresWithAuth(router)

// With credentials support
router := gin.Default()
allowedOrigins := []string{"http://localhost:3000", "https://yourdomain.com"}
middleware.SetupMiddlewaresWithCredentials(router, allowedOrigins)
```

### Headers Supported

The CORS configuration supports the following headers:
- Origin
- Content-Length
- Content-Type
- Authorization
- aaa-auth-token (for gRPC authentication)
- X-Requested-With
- Accept
- Accept-Encoding
- Accept-Language
- Cache-Control
- Connection
- DNT
- Host
- Pragma
- Referer
- User-Agent

### Security Notes

- Using `AllowOrigins: ["*"]` with `AllowCredentials: true` is not allowed by browsers
- For production with credentials, specify exact origins instead of using `*`
- The current configuration is suitable for development and testing environments 