package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
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

func (h *FarmHandler) CreateFarmHandler(c *gin.Context) {
	// Step 0: Header validation
	actorID := c.GetHeader("user-id")
	if actorID == "" {
		sendStandardError(c, http.StatusUnauthorized, 
			"Please include your user ID in headers",
			"missing user-id header")
		return
	}

	// Step 1: User verification
	exists, isKisansathi, err := h.userService.VerifyUserAndType(actorID)
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

	// Step 2: Parse body
	var farmRequest models.FarmRequest
	if err := c.ShouldBindJSON(&farmRequest); err != nil {
		sendStandardError(c, http.StatusBadRequest,
			"Invalid farm details provided",
			"request body parsing failed: "+err.Error())
		return
	}

	// Validate location polygon
	if len(farmRequest.Location) < 4 {
		sendStandardError(c, http.StatusBadRequest,
			"A polygon requires at least 4 points (first and last should be same)",
			"insufficient polygon points")
		return
	}

	// Check if first and last points are same (closed polygon)
	first := farmRequest.Location[0]
	last := farmRequest.Location[len(farmRequest.Location)-1]
	if first[0] != last[0] || first[1] != last[1] {
		sendStandardError(c, http.StatusBadRequest,
			"Polygon must be closed (first and last points should be identical)",
			"unclosed polygon")
		return
	}

	// Validate coordinates
	for _, point := range farmRequest.Location {
		if len(point) != 2 {
			sendStandardError(c, http.StatusBadRequest,
				"Each location point must have exactly 2 values (latitude, longitude)",
				"invalid coordinate format")
			return
		}
		lat, lon := point[0], point[1]
		if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
			sendStandardError(c, http.StatusBadRequest,
				"Invalid coordinates (latitude must be -90 to 90, longitude -180 to 180)",
				"invalid coordinate range")
			return
		}
	}

	// Step 3: Permission check
	requiredAction := "CREATE_UNVERIFIED_FARM"
	if isKisansathi {
		requiredAction = "CREATE_VERIFIED_FARM"
	}

	isAllowed, err := services.ValidateActionClient(c.Request.Context(), actorID, requiredAction)
	if err != nil {
		sendStandardError(c, http.StatusInternalServerError,
			"Permission verification failed",
			fmt.Sprintf("AAA service error: %v", err))
		return
	}
	if !isAllowed {
		sendStandardError(c, http.StatusForbidden,
			"You don't have permission",
			fmt.Sprintf("action %s not allowed", requiredAction))
		return
	}

	// Step 4: Farm creation
	farm, err := h.farmService.CreateFarm(
		c.Request.Context(),
		farmRequest.FarmerID,
		farmRequest.Location,
		farmRequest.Area,
		farmRequest.Locality,
		farmRequest.CropType,
		isKisansathi,
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