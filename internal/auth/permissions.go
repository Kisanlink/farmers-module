package auth

import (
	"fmt"
	"strings"
)

// Permission represents a resource-action pair
type Permission struct {
	Resource string
	Action   string
}

// String returns the string representation of the permission
func (p Permission) String() string {
	return fmt.Sprintf("%s.%s", p.Resource, p.Action)
}

// RoutePermissionMap maps HTTP routes to required permissions
var RoutePermissionMap = map[string]Permission{
	// Identity - Farmer management routes
	"POST /api/v1/identity/farmers":            {Resource: "farmer", Action: "create"},
	"GET /api/v1/identity/farmers":             {Resource: "farmer", Action: "list"},
	"GET /api/v1/identity/farmers/id/:id":      {Resource: "farmer", Action: "read"},
	"GET /api/v1/identity/farmers/user/:id":    {Resource: "farmer", Action: "read"},
	"GET /api/v1/identity/farmers/:id":         {Resource: "farmer", Action: "read"},
	"PUT /api/v1/identity/farmers/id/:id":      {Resource: "farmer", Action: "update"},
	"PUT /api/v1/identity/farmers/user/:id":    {Resource: "farmer", Action: "update"},
	"PUT /api/v1/identity/farmers/:id":         {Resource: "farmer", Action: "update"},
	"DELETE /api/v1/identity/farmers/id/:id":   {Resource: "farmer", Action: "delete"},
	"DELETE /api/v1/identity/farmers/user/:id": {Resource: "farmer", Action: "delete"},
	"DELETE /api/v1/identity/farmers/:id":      {Resource: "farmer", Action: "delete"},

	// Legacy farmer management routes (if any)
	"POST /api/v1/farmers":       {Resource: "farmer", Action: "create"},
	"GET /api/v1/farmers/:id":    {Resource: "farmer", Action: "read"},
	"PUT /api/v1/farmers/:id":    {Resource: "farmer", Action: "update"},
	"DELETE /api/v1/farmers/:id": {Resource: "farmer", Action: "delete"},
	"GET /api/v1/farmers":        {Resource: "farmer", Action: "list"},

	// Identity - Farmer linkage routes
	"POST /api/v1/identity/farmer/link":       {Resource: "farmer", Action: "link"},
	"DELETE /api/v1/identity/farmer/unlink":   {Resource: "farmer", Action: "unlink"},
	"GET /api/v1/identity/farmer/linkage/:id": {Resource: "farmer", Action: "read"},

	// Identity - KisanSathi routes
	"POST /api/v1/identity/kisansathi/assign":        {Resource: "farmer", Action: "assign_kisan_sathi"},
	"PUT /api/v1/identity/kisansathi/reassign":       {Resource: "farmer", Action: "assign_kisan_sathi"},
	"POST /api/v1/identity/kisansathi/create-user":   {Resource: "kisansathi", Action: "create"},
	"GET /api/v1/identity/kisansathi/assignment/:id": {Resource: "farmer", Action: "read"},

	// Identity - FPO routes
	"POST /api/v1/identity/fpo/create":       {Resource: "fpo", Action: "create"},
	"POST /api/v1/identity/fpo/register":     {Resource: "fpo", Action: "create"},
	"GET /api/v1/identity/fpo/reference/:id": {Resource: "fpo", Action: "read"},

	// Legacy FPO management routes
	"POST /api/v1/fpos":       {Resource: "fpo", Action: "create"},
	"GET /api/v1/fpos/:id":    {Resource: "fpo", Action: "read"},
	"PUT /api/v1/fpos/:id":    {Resource: "fpo", Action: "update"},
	"DELETE /api/v1/fpos/:id": {Resource: "fpo", Action: "delete"},
	"GET /api/v1/fpos":        {Resource: "fpo", Action: "list"},

	// Farmer linkage routes
	"POST /api/v1/farmer-links":            {Resource: "farmer", Action: "link"},
	"DELETE /api/v1/farmer-links":          {Resource: "farmer", Action: "unlink"},
	"PUT /api/v1/farmer-links/kisan-sathi": {Resource: "farmer", Action: "assign_kisan_sathi"},

	// Farm management routes
	"POST /api/v1/farms":       {Resource: "farm", Action: "create"},
	"GET /api/v1/farms/:id":    {Resource: "farm", Action: "read"},
	"PUT /api/v1/farms/:id":    {Resource: "farm", Action: "update"},
	"DELETE /api/v1/farms/:id": {Resource: "farm", Action: "delete"},
	"GET /api/v1/farms":        {Resource: "farm", Action: "list"},

	// Crop master data routes
	"POST /api/v1/crops":       {Resource: "crop", Action: "create"},
	"GET /api/v1/crops/:id":    {Resource: "crop", Action: "read"},
	"PUT /api/v1/crops/:id":    {Resource: "crop", Action: "update"},
	"DELETE /api/v1/crops/:id": {Resource: "crop", Action: "delete"},
	"GET /api/v1/crops":        {Resource: "crop", Action: "list"},

	// Crop variety routes
	"POST /api/v1/varieties":         {Resource: "crop", Action: "create"},
	"GET /api/v1/varieties/:id":      {Resource: "crop", Action: "read"},
	"PUT /api/v1/varieties/:id":      {Resource: "crop", Action: "update"},
	"DELETE /api/v1/varieties/:id":   {Resource: "crop", Action: "delete"},
	"GET /api/v1/varieties":          {Resource: "crop", Action: "list"},
	"GET /api/v1/crop-varieties/:id": {Resource: "crop", Action: "list"},

	// Crop cycle routes
	"POST /api/v1/crops/cycles":        {Resource: "cycle", Action: "start"},
	"GET /api/v1/crops/cycles/:id":     {Resource: "cycle", Action: "read"},
	"PUT /api/v1/crops/cycles/:id":     {Resource: "cycle", Action: "update"},
	"PUT /api/v1/crops/cycles/:id/end": {Resource: "cycle", Action: "end"},
	"DELETE /api/v1/crops/cycles/:id":  {Resource: "cycle", Action: "end"},
	"GET /api/v1/crops/cycles":         {Resource: "cycle", Action: "list"},

	// Farm activity routes
	"POST /api/v1/crops/activities":             {Resource: "activity", Action: "create"},
	"GET /api/v1/crops/activities/:id":          {Resource: "activity", Action: "read"},
	"PUT /api/v1/crops/activities/:id":          {Resource: "activity", Action: "update"},
	"PUT /api/v1/crops/activities/:id/complete": {Resource: "activity", Action: "complete"},
	"GET /api/v1/crops/activities":              {Resource: "activity", Action: "list"},

	// Lookup/Dropdown data routes
	"GET /api/v1/lookups/crops":           {Resource: "crop", Action: "list"},
	"GET /api/v1/lookups/varieties/:id":   {Resource: "crop", Action: "list"},
	"GET /api/v1/lookups/crop-categories": {Resource: "crop", Action: "list"},
	"GET /api/v1/lookups/crop-seasons":    {Resource: "crop", Action: "list"},

	// Data quality routes
	"POST /api/v1/data-quality/validate-geometry":       {Resource: "farm", Action: "audit"},
	"POST /api/v1/data-quality/reconcile-aaa-links":     {Resource: "admin", Action: "maintain"},
	"POST /api/v1/data-quality/rebuild-spatial-indexes": {Resource: "admin", Action: "maintain"},
	"POST /api/v1/data-quality/detect-farm-overlaps":    {Resource: "farm", Action: "audit"},

	// Reporting routes
	"GET /api/v1/reports/farmer-portfolio": {Resource: "report", Action: "read"},
	"GET /api/v1/reports/org-dashboard":    {Resource: "report", Action: "read"},

	// Administrative routes
	"POST /api/v1/admin/seed-roles": {Resource: "admin", Action: "maintain"},
	"GET /api/v1/health":            {Resource: "system", Action: "health"},
}

// GetPermissionForRoute returns the required permission for a given HTTP method and path
func GetPermissionForRoute(method, path string) (Permission, bool) {
	// Normalize the route by replacing path parameters with placeholders
	normalizedPath := normalizePath(path)
	routeKey := fmt.Sprintf("%s %s", method, normalizedPath)

	permission, exists := RoutePermissionMap[routeKey]
	return permission, exists
}

// normalizePath converts actual paths to route patterns
// e.g., "/api/v1/identity/farmers/123" -> "/api/v1/identity/farmers/:id"
func normalizePath(path string) string {
	// Remove trailing slash if present (except for root "/")
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	// Split path into segments
	segments := strings.Split(path, "/")

	// Handle identity routes: /api/v1/identity/resource/...
	if len(segments) >= 5 && segments[1] == "api" && segments[2] == "v1" && segments[3] == "identity" {
		resource := segments[4] // farmers, fpo, etc.

		if len(segments) == 5 {
			// Pattern: /api/v1/identity/farmers -> /api/v1/identity/farmers (no normalization needed)
			return path
		}
		if len(segments) == 6 {
			// Pattern: /api/v1/identity/farmers/123 -> /api/v1/identity/farmers/:id
			return fmt.Sprintf("/api/v1/identity/%s/:id", resource)
		}
		if len(segments) == 7 {
			subPath := segments[5] // id, user, etc.
			if subPath == "id" || subPath == "user" || subPath == "reference" {
				// Pattern: /api/v1/identity/farmers/id/123 -> /api/v1/identity/farmers/id/:id
				return fmt.Sprintf("/api/v1/identity/%s/%s/:id", resource, subPath)
			}
			// Pattern: /api/v1/identity/farmers/123/action -> /api/v1/identity/farmers/:id/action
			return fmt.Sprintf("/api/v1/identity/%s/:id/%s", resource, segments[6])
		}
		if len(segments) == 8 {
			subPath := segments[5] // id, user, etc.
			// Pattern: /api/v1/identity/farmers/id/123/action -> /api/v1/identity/farmers/id/:id/action
			return fmt.Sprintf("/api/v1/identity/%s/%s/:id/%s", resource, subPath, segments[7])
		}
	}

	// Handle other routes
	if len(segments) >= 4 {
		// Check for common API patterns
		if len(segments) == 5 && segments[1] == "api" && segments[2] == "v1" {
			// Pattern: /api/v1/resource/id
			return fmt.Sprintf("/api/v1/%s/:id", segments[3])
		}
		if len(segments) == 6 && segments[1] == "api" && segments[2] == "v1" {
			// Pattern: /api/v1/resource/id/action
			return fmt.Sprintf("/api/v1/%s/:id/%s", segments[3], segments[5])
		}
	}

	return path
}

// IsPublicRoute checks if a route is public and doesn't require authentication
func IsPublicRoute(method, path string) bool {
	publicRoutes := []string{
		"GET /health",
		"GET /api/v1/health",
		"GET /docs",
		"GET /docs/swagger.json",
		"GET /swagger",
		"GET /",
	}

	routeKey := fmt.Sprintf("%s %s", method, path)
	for _, publicRoute := range publicRoutes {
		if routeKey == publicRoute || strings.HasPrefix(path, "/docs") || strings.HasPrefix(path, "/swagger") {
			return true
		}
	}

	return false
}
