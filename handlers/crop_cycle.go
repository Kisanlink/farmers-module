package handlers

import (
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

type CropCycleHandler struct {
	service services.CropCycleServiceInterface
}

func NewCropCycleHandler(service services.CropCycleServiceInterface) *CropCycleHandler {
	return &CropCycleHandler{service: service}
}

type CreateCropCycleRequest struct {
	FarmId           string    `json:"farm_id" binding:"required"`
	CropId           string    `json:"crop_id" binding:"required"`
	StartDate        time.Time `json:"start_date" binding:"required"`
	EndDate          time.Time `json:"end_date"`
	Acreage          float64   `json:"acreage"`
	ExpectedQuantity float64   `json:"expected_quantity"`
	Quantity         float64   `json:"quantity"`
	Report           string    `json:"report"`
}

func (h *CropCycleHandler) CreateCropCycle(c *gin.Context) {
	var req CreateCropCycleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Invalid input: " + err.Error(),
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	cycle, err := h.service.CreateCropCycle(
		req.FarmId,
		req.CropId,
		req.StartDate,
		req.EndDate,
		req.Acreage,
		req.ExpectedQuantity,
		req.Quantity,
		req.Report,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Could not create crop cycle: " + err.Error(),
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		StatusCode: http.StatusCreated,
		Success:    true,
		Message:    "Crop cycle created successfully",
		Data:       cycle,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *CropCycleHandler) GetCropCycles(c *gin.Context) {
	cropCycleID := c.Query("id")
	farmID := c.Query("farm_id")
	cropID := c.Query("crop_id")

	var (
		data interface{}
		err  error
	)

	switch {
	case cropCycleID != "":
		data, err = h.service.GetCropCycleByID(cropCycleID)
	case farmID != "" && cropID != "":
		data, err = h.service.GetCropCyclesByFarmAndCropID(farmID, cropID)
	case farmID != "":
		data, err = h.service.GetCropCyclesByFarmID(farmID)
	case cropID != "":
		data, err = h.service.GetCropCyclesByCropID(cropID)
	default:
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Please provide at least one of 'id', 'farm_id', or 'crop_id'",
			Error:      "Missing query parameter",
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to retrieve crop cycle(s): " + err.Error(),
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Crop cycle(s) retrieved successfully",
		Data:       data,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}
