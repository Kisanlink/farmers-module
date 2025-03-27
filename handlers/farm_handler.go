package handlers

import (
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

type FarmRequest struct {
	KisansathiUserID *string `json:"kisansathi_user_id,omitempty"` 
	FarmerID        string  `json:"farmer_id" validate:"required"` 
	Location        [][][]float64 `json:"location" validate:"required,min=4"`
	Area            float64 `json:"area" validate:"required,gt=0"`
	Locality        string  `json:"locality" validate:"required"`
	CropType        string  `json:"crop_type" validate:"required"`
	IsVerified      bool    `json:"is_verified"`
	RequestedBy     string  `json:"-"`
}

func (h *FarmHandler) CreateFarmHandler(c *gin.Context) {
    
  //   // Step 0: Header validation
	// actorID := c.GetHeader("user-id")
	// if actorID == "" {
	// 	sendStandardError(c, http.StatusUnauthorized, 
	// 		"Please include your user ID in headers",
	// 		"missing user-id header")
	// 	return
	// }

	// //Step 1: User verification via service layer
	// exists, isKisansathi, err := h.userService.VerifyUserAndType(actorID)
	// if err != nil {
	// 	sendStandardError(c, http.StatusInternalServerError,
	// 		"Something went wrong on our end",
	// 		"user verification failed: "+err.Error())
	// 	return
	// }
	// if !exists {
	// 	sendStandardError(c, http.StatusUnauthorized,
	// 		"Your account isn't registered",
	// 		"user not found in farmer/kisansathi records")
	// 	return
	// }

    // Parse request body
    var farmRequest FarmRequest
    if err := c.ShouldBindJSON(&farmRequest); err != nil {
        sendStandardError(c, http.StatusBadRequest,
            "Invalid farm details provided",
            "request body parsing failed: "+err.Error())
        return
    }

  //   	requiredAction := "CREATE_UNVERIFIED_FARM"
	// if isKisansathi {
	// 	requiredAction = "CREATE_VERIFIED_FARM"
	// }

	// isAllowed, err := services.ValidateActionClient(c.Request.Context(), actorID, requiredAction)
	// if err != nil {
	// 	sendStandardError(c, http.StatusInternalServerError,
	// 		"Permission verification failed",
	// 		fmt.Sprintf("AAA service error: %v", err))
	// 	return
	// }
	// if !isAllowed {
	// 	sendStandardError(c, http.StatusForbidden,
	// 		"You don't have permission",
	// 		fmt.Sprintf("action %s not allowed", requiredAction))
	// 	return
	// }
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
        farmRequest.FarmerID,
        geoJSONPolygon,
        farmRequest.Area,
        farmRequest.Locality,
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