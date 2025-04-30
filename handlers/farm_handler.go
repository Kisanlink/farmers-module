package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/Kisanlink/farmers-module/utils"
	"github.com/gin-gonic/gin"
)

type FarmHandler struct {
	FarmService services.FarmServiceInterface
	UserService services.UserServiceInterface
}

func NewFarmHandler(
	farm_service services.FarmServiceInterface,
	user_service services.UserServiceInterface,
) *FarmHandler {
	return &FarmHandler{
		FarmService: farm_service,
		UserService: user_service,
	}
}

type FarmRequest struct {
	KisansathiUserId *string       `json:"kisansathi_user_id,omitempty"`
	FarmerId         string        `json:"farmer_id" validate:"required"`
	Location         [][][]float64 `json:"location" validate:"required,min=4"`
	Area             float64       `json:"area" validate:"required,gt=0"`
	Locality         string        `json:"locality" validate:"required"`
	RequestedBy      string        `json:"-"`
	Pincode          int           `json:"pincode" validate:"required"`
	OwnerId          string        `json:"owner_id,omitempty"`
}

func (h *FarmHandler) CreateFarmHandler(c *gin.Context) {

	// Step 0: Header validation
	actor_id := c.GetHeader("user-id")
	if actor_id == "" {
		utils.Log.Warn("Missing user-id header")
		sendStandardError(c, http.StatusUnauthorized,
			"Please include your user ID in headers",
			"missing user-id header")
		return
	}

	// Step 1: User verification via service layer
	exists, is_kisansathi, err := h.UserService.VerifyUserAndType(actor_id)
	if err != nil {
		utils.Log.Errorf("User verification failed: %v", err)
		sendStandardError(c, http.StatusInternalServerError,
			"Something went wrong on our end",
			"user verification failed: "+err.Error())
		return
	}
	if !exists {
		utils.Log.Warnf("User not found: %s", actor_id)
		sendStandardError(c, http.StatusUnauthorized,
			"Your account isn't registered",
			"user not found in farmer/kisansathi records")
		return
	}

	// Parse request body
	var farmRequest FarmRequest
	if err := c.ShouldBindJSON(&farmRequest); err != nil {
		utils.Log.Errorf("Failed to parse request body: %v", err)
		sendStandardError(c, http.StatusBadRequest,
			"Invalid farm details provided",
			"request body parsing failed: "+err.Error())
		return
	}

	// Determine required action based on user type
	required_action := "read"
	if is_kisansathi {
		required_action = "read"
	}

	// Get user details to check actions
	user_resp, err := services.GetUserByIdClient(c.Request.Context(), actor_id)
	if err != nil {
		utils.Log.Errorf("GetUserByIdClient failed for user: %s, error: %v", actor_id, err)

		sendStandardError(c, http.StatusInternalServerError,
			"Failed to verify user actions", err.Error())
		return
	}

	// Verify the required action exists in user's allowed actions
	has_action := false
	if user_resp != nil && user_resp.Data != nil && user_resp.Data.RolePermissions != nil {
		for _, role_perms := range user_resp.Data.RolePermissions {
			for _, permission := range role_perms.Permissions {
				if permission != nil && permission.Action == required_action {
					has_action = true
					break
				}
			}
			if has_action {
				break
			}
		}
	}

	if !has_action {
		utils.Log.Warnf("User %s does not have permission: %s", actor_id, required_action)

		sendStandardError(c, http.StatusForbidden,
			"Action not permitted",
			fmt.Sprintf("missing required action: %s", required_action))
		return
	}
	// Convert to proper GeoJSON structure
	geoJSONPolygon := models.GeoJSONPolygon{
		Type:        "Polygon",
		Coordinates: farmRequest.Location,
	}

	// Validate coordinates
	if len(geoJSONPolygon.Coordinates) == 0 || len(geoJSONPolygon.Coordinates[0]) < 4 {
		utils.Log.Warnf("Invalid polygon coordinates: %+v", geoJSONPolygon.Coordinates)

		sendStandardError(c, http.StatusBadRequest,
			"A polygon requires at least 4 points",
			"insufficient polygon points")
		return
	}

	//Call Service layer to create farm
	farm, err := h.FarmService.CreateFarm(
		farmRequest.FarmerId,
		geoJSONPolygon,
		farmRequest.Area,
		farmRequest.Locality,
		farmRequest.Pincode,
		farmRequest.OwnerId,
	)

	if err != nil {
		utils.Log.Errorf("Farm creation failed: %v", err)

		handleFarmCreationError(c, err)
		return
	}

	// API call for divya drishti to create farm data
	// Start a goroutine to handle the CreateFarmData call asynchronously
	go func(farm_id string) {
		// You might want to add some error handling or logging here
		defer func() {
			if r := recover(); r != nil {
				utils.Log.Errorf("Recovered from panic in CreateFarmData goroutine: %v", r)
			}
		}()

		CreateFarmData(farm_id)
	}(farm.Id)

	c.JSON(http.StatusCreated, gin.H{
		"status":    http.StatusCreated,
		"message":   "Farm created successfully",
		"data":      farm,
		"timestamp": time.Now().UTC(),
		"success":   true,
	})
}

func handleFarmCreationError(c *gin.Context, err error) {
	switch {
	case strings.Contains(err.Error(), "invalid location"):
		utils.Log.Warn("Invalid farm location format")

		sendStandardError(c, http.StatusBadRequest,
			"Please provide a valid farm boundary",
			"Invalid farm location format")
	case strings.Contains(err.Error(), "overlap"):
		utils.Log.Warn("Farm location overlaps with existing farm")

		sendStandardError(c, http.StatusConflict,
			"This farm overlaps with an existing farm",
			"Farm location overlaps")
	default:
		utils.Log.Error("Unhandled farm creation error: ", err)

		sendStandardError(c, http.StatusInternalServerError,
			"Something went wrong while creating your farm",
			err.Error())
	}
}

func sendStandardError(c *gin.Context, status int, user_message string, error_detail string) {
	c.JSON(status, gin.H{
		"status":    status,
		"message":   user_message,
		"error":     error_detail,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      nil,
		"success":   false,
	})
}

// Add these methods to FarmHandler struct

// GetFarmsHandler retrieves farms with optional filters
// GetFarmsHandler retrieves all farms
func (h *FarmHandler) GetFarmsHandler(c *gin.Context) {
	// Extract query parameters
	farmer_id := c.Query("farmer_id")
	pincode := c.Query("pincode")
	created_at_from := c.Query("created_at_from")
	id := c.Query("id") // New ID parameter
	//createdAtTo := c.Query("created_at_from")
	//updatedAtFrom := c.Query("created_at_from")
	//updateAtTo:= c.Query("created_at_from") // New date parameter

	var parsed_date time.Time
	if created_at_from != "" {
		var err error
		// Try parsing with time and time zone
		parsed_date, err = time.Parse(time.RFC3339, created_at_from)
		if err != nil {
			// If parsing fails, try parsing as date only and default time to 12:00 AM
			parsed_date, err = time.Parse("2006-01-02", created_at_from)
			if err != nil {
				utils.Log.Warnf("Invalid date format: %s", created_at_from)
				sendStandardError(c, http.StatusBadRequest,
					"Invalid date format",
					"Date must be in YYYY-MM-DD or RFC3339 format (e.g., 2025-04-08T00:00:00Z)")
				return
			}
		}
	}

	// Call service layer with the new date parameter
	farms, err := h.FarmService.GetAllFarms(farmer_id, pincode, parsed_date.Format(time.RFC3339), id)
	if err != nil {
		utils.Log.Errorf("Failed to retrieve farms: %v", err)
		sendStandardError(c, http.StatusInternalServerError,
			"Failed to retrieve farms",
			err.Error())
		return
	}

	utils.Log.Infof("Farms retrieved successfully for farmer: %s", farmer_id)

	c.JSON(http.StatusOK, gin.H{
		"status":    http.StatusOK,
		"message":   "Farms retrieved successfully",
		"data":      farms,
		"timestamp": time.Now().UTC(),
		"success":   true,
	})
}

func (h *FarmHandler) GetFarmByFarmID(c *gin.Context) {
	// Retrieve farm_id from URL parameters
	farm_id := c.Param("farm_id")
	if farm_id == "" {
		utils.Log.Warn("Empty farm_id parameter")

		sendStandardError(c, http.StatusBadRequest,
			"Farm ID is required",
			"empty farm id parameter")
		return
	}

	// Call the service layer method to retrieve the farm by ID
	farm, err := h.FarmService.GetFarmByID(farm_id)
	if err != nil {
		// If farm not found, return a 404 error
		if err.Error() == "farm not found" {
			utils.Log.Warnf("Farm not found: %s", farm_id)

			sendStandardError(c, http.StatusNotFound,
				"Farm does not exist",
				"Farm with the provided ID does not exist")
			return
		}

		// Handle any other errors
		utils.Log.Errorf("Error retrieving farm: %v", err)

		sendStandardError(c, http.StatusInternalServerError,
			"Internal server error",
			err.Error())
		return
	}
	utils.Log.Infof("Farm retrieved successfully: %s", farm_id)

	// Respond with the farm data if found
	c.JSON(http.StatusOK, gin.H{
		"status":    http.StatusOK,
		"message":   "Farm retrieved successfully",
		"data":      farm,
		"timestamp": time.Now().UTC(),
		"success":   true,
	})
}
