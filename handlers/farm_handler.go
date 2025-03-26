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

// handlers/farm_handler.go

func (h *FarmHandler) CreateFarmHandler(c *gin.Context) {
    var farmRequest models.FarmRequest
    if err := c.ShouldBindJSON(&farmRequest); err != nil {
        sendStandardError(c, http.StatusBadRequest,
            "Invalid farm details provided",
            "request body parsing failed: "+err.Error())
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

    // Auto-close polygon if not closed
    ring := geoJSONPolygon.Coordinates[0]
    first, last := ring[0], ring[len(ring)-1]
    if first[0] != last[0] || first[1] != last[1] {
        geoJSONPolygon.Coordinates[0] = append(ring, ring[0])
    }

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