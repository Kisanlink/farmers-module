package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

type FarmHandler struct {
	farmService services.FarmServiceInterface
	userService services.UserServiceInterface
}

func NewFarmHandler(
	farmService services.FarmServiceInterface,
	userService services.UserServiceInterface,
) *FarmHandler {
	return &FarmHandler{
		farmService: farmService,
		userService: userService,
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
	actorId := c.GetHeader("user-id")
	if actorId == "" {
		sendStandardError(c, http.StatusUnauthorized,
			"Please include your user ID in headers",
			"missing user-id header")
		return
	}

	//Step 1: User verification via service layer
	exists, isKisansathi, err := h.userService.VerifyUserAndType(actorId)
	if err != nil {
		sendStandardError(c, http.StatusInternalServerError,
			"Something went wrong on our end",
			"user verification failed: "+err.Error())
		return
	}
	if !exists {
		sendStandardError(c, http.StatusUnauthorized,
			"Your account isn't registered",
			"user not found in farmer/kisansathi records")
		return
	}

	// Parse request body
	var farmRequest FarmRequest
	if err := c.ShouldBindJSON(&farmRequest); err != nil {
		sendStandardError(c, http.StatusBadRequest,
			"Invalid farm details provided",
			"request body parsing failed: "+err.Error())
		return
	}

	// Determine required action based on user type
	requiredAction := "read"
	if isKisansathi {
		requiredAction = "read"
	}

	// Get user details to check actions
	userResp, err := services.GetUserByIdClient(c.Request.Context(), actorId)
	if err != nil {
		sendStandardError(c, http.StatusInternalServerError,
			"Failed to verify user actions", err.Error())
		return
	}

	// Verify the required action exists in user's allowed actions
	// Verify the required action exists in user's allowed actions
	hasAction := false
	if userResp != nil && userResp.Data != nil && userResp.Data.UsageRight != nil {
		for _, permission := range userResp.Data.UsageRight.Permissions {
			if permission != nil && permission.Action == requiredAction {
				hasAction = true
				break
			}
		}
	}

	if !hasAction {
		sendStandardError(c, http.StatusForbidden,
			"Action not permitted",
			fmt.Sprintf("missing required action: %s", requiredAction))
		return
	}
	// Convert to proper GeoJSON structure
	geoJSONPolygon := models.GeoJSONPolygon{
		Type:        "Polygon",
		Coordinates: farmRequest.Location,
	}

	// Validate coordinates
	if len(geoJSONPolygon.Coordinates) == 0 || len(geoJSONPolygon.Coordinates[0]) < 4 {
		sendStandardError(c, http.StatusBadRequest,
			"A polygon requires at least 4 points",
			"insufficient polygon points")
		return
	}

	//Call Service layer to create farm
	farm, err := h.farmService.CreateFarm(
		farmRequest.FarmerId,
		geoJSONPolygon,
		farmRequest.Area,
		farmRequest.Locality,
		farmRequest.Pincode,
		farmRequest.OwnerId,
	)

	if err != nil {
		handleFarmCreationError(c, err)
		return
	}

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
		sendStandardError(c, http.StatusBadRequest,
			"Please provide a valid farm boundary",
			"Invalid farm location format")
	case strings.Contains(err.Error(), "overlap"):
		sendStandardError(c, http.StatusConflict,
			"This farm overlaps with an existing farm",
			"Farm location overlaps")
	default:
		sendStandardError(c, http.StatusInternalServerError,
			"Something went wrong while creating your farm",
			err.Error())
	}
}

func sendStandardError(c *gin.Context, status int, userMessage string, errorDetail string) {
	c.JSON(status, gin.H{
		"status":    status,
		"message":   userMessage,
		"error":     errorDetail,
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
	farmerId := c.Query("farmer_id")
	pincode := c.Query("pincode")
	createdAtFrom := c.Query("created_at_from")
	id := c.Query("id") // New ID parameter
	//createdAtTo := c.Query("created_at_from")
	//updatedAtFrom := c.Query("created_at_from")
	//updateAtTo:= c.Query("created_at_from") // New date parameter

	var parsedDate time.Time
	if createdAtFrom != "" {
		var err error
		// Try parsing with time and time zone
		parsedDate, err = time.Parse(time.RFC3339, createdAtFrom)
		if err != nil {
			// If parsing fails, try parsing as date only and default time to 12:00 AM
			parsedDate, err = time.Parse("2006-01-02", createdAtFrom)
			if err != nil {
				sendStandardError(c, http.StatusBadRequest,
					"Invalid date format",
					"Date must be in YYYY-MM-DD or RFC3339 format (e.g., 2025-04-08T00:00:00Z)")
				return
			}
		}
	}

	// Call service layer with the new date parameter
	farms, err := h.farmService.GetAllFarms(farmerId, pincode, parsedDate.Format(time.RFC3339), id)
	if err != nil {
		sendStandardError(c, http.StatusInternalServerError,
			"Failed to retrieve farms",
			err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    http.StatusOK,
		"message":   "Farms retrieved successfully",
		"data":      farms,
		"timestamp": time.Now().UTC(),
		"success":   true,
	})
}
