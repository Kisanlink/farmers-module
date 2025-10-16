package constants

// AAA Role Names
// These constants define the role names used in the AAA (Authentication, Authorization, Auditing) service.
// IMPORTANT: These must match exactly with the role names configured in the AAA service.
// Any changes to these values must be coordinated with the AAA service team.
const (
	// RoleFarmer is assigned to all farmer users upon registration
	// Grants permissions to manage own farms, crop cycles, and activities
	RoleFarmer = "farmer"

	// RoleKisanSathi is assigned to field agents (KisanSathis)
	// Grants permissions to manage farmers assigned to them and their farms
	// Note: Uses lowercase to match AAA service configuration
	RoleKisanSathi = "kisansathi"

	// RoleFPOCEO is assigned to FPO (Farmer Producer Organization) chief executive officers
	// Grants permissions to manage the entire FPO organization and all its farmers
	// Note: Both "CEO" and "fpo_ceo" may be used depending on AAA service configuration
	RoleFPOCEO = "CEO"

	// RoleFPOManager is assigned to FPO managers (non-CEO leadership roles)
	// Grants permissions to manage FPO operations and access farmer data
	RoleFPOManager = "fpo_manager"

	// RoleAdmin is assigned to system administrators
	// Grants full permissions across all organizations and resources
	RoleAdmin = "admin"

	// RoleReadOnly is assigned to users who need read-only access
	// Grants read permissions without write/update/delete capabilities
	RoleReadOnly = "readonly"
)

// RoleDisplayNames maps internal role names to human-readable display names
// Used for UI rendering and user-facing messages
var RoleDisplayNames = map[string]string{
	RoleFarmer:     "Farmer",
	RoleKisanSathi: "KisanSathi",
	RoleFPOCEO:     "FPO CEO",
	RoleFPOManager: "FPO Manager",
	RoleAdmin:      "Administrator",
	RoleReadOnly:   "Read-Only User",
}

// RoleDescriptions provides detailed descriptions of each role's purpose and permissions
var RoleDescriptions = map[string]string{
	RoleFarmer:     "Individual agricultural practitioner with permissions to manage own farms and crop cycles",
	RoleKisanSathi: "Field agent assigned to support farmers with data collection and advisory services",
	RoleFPOCEO:     "Chief executive officer of a Farmer Producer Organization with full organizational management permissions",
	RoleFPOManager: "Manager within an FPO with permissions to manage operations and access organizational data",
	RoleAdmin:      "System administrator with full permissions across all organizations and resources",
	RoleReadOnly:   "User with read-only access to resources within their organization",
}

// AllRoles returns a slice of all defined role names
// Useful for validation and iteration
func AllRoles() []string {
	return []string{
		RoleFarmer,
		RoleKisanSathi,
		RoleFPOCEO,
		RoleFPOManager,
		RoleAdmin,
		RoleReadOnly,
	}
}

// IsValidRole checks if a given role name is a known valid role
func IsValidRole(roleName string) bool {
	for _, role := range AllRoles() {
		if role == roleName {
			return true
		}
	}
	return false
}

// GetRoleDisplayName returns the display name for a given role
// Returns the role name itself if no display name is defined
func GetRoleDisplayName(roleName string) string {
	if displayName, exists := RoleDisplayNames[roleName]; exists {
		return displayName
	}
	return roleName
}

// GetRoleDescription returns the description for a given role
// Returns an empty string if no description is defined
func GetRoleDescription(roleName string) string {
	if description, exists := RoleDescriptions[roleName]; exists {
		return description
	}
	return ""
}
