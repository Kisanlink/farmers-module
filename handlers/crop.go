package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/Kisanlink/farmers-module/entities"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

type CropHandler struct {
	service services.CropServiceInterface
}

func NewCropHandler(service services.CropServiceInterface) *CropHandler {
	return &CropHandler{service: service}
}

// CreateCrop handles POST /crops
func (h *CropHandler) CreateCrop(c *gin.Context) {
	var crop models.Crop
	if err := c.ShouldBindJSON(&crop); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Invalid input data",
			Data:       nil,
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	if err := h.service.CreateCrop(&crop); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to create crop",
			Data:       nil,
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		StatusCode: http.StatusCreated,
		Success:    true,
		Message:    "Crop created successfully",
		Data:       crop,
		Error:      nil,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}

// GetAllCrops handles GET /crops
func (h *CropHandler) GetAllCrops(c *gin.Context) {
	// Get query parameters
	name := c.Query("name")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	crops, total, err := h.service.GetAllCrops(name, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to get crops",
			Data:       nil,
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	// Create pagination response
	pagination := map[string]interface{}{
		"current_page": page,
		"page_size":    pageSize,
		"total_items":  total,
		"total_pages":  int(math.Ceil(float64(total) / float64(pageSize))),
	}

	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Crops retrieved successfully",
		Data: map[string]interface{}{
			"crops":      crops,
			"pagination": pagination,
		},
		Error:     nil,
		TimeStamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// GetCropByID handles GET /crops/:id
func (h *CropHandler) GetCropByID(c *gin.Context) {
	id := c.Param("id")

	crop, err := h.service.GetCropByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.Response{
			StatusCode: http.StatusNotFound,
			Success:    false,
			Message:    "Crop not found",
			Data:       nil,
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Crop retrieved successfully",
		Data:       crop,
		Error:      nil,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}

// UpdateCrop handles PUT /crops/:id
func (h *CropHandler) UpdateCrop(c *gin.Context) {
	id := c.Param("id")

	var crop models.Crop
	if err := c.ShouldBindJSON(&crop); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Invalid input data",
			Data:       nil,
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	// Validate Category if provided
	if crop.Category != "" {
		if !entities.CROP_CATEGORIES.IsValid(string(crop.Category)) {
			c.JSON(http.StatusBadRequest, models.Response{
				StatusCode: http.StatusBadRequest,
				Success:    false,
				Message:    "Invalid crop category",
				Data:       nil,
				Error: fmt.Sprintf("Invalid category: %s. Valid values are: %v",
					crop.Category, entities.CROP_CATEGORIES),
				TimeStamp: time.Now().UTC().Format(time.RFC3339),
			})
			return
		}
	}

	// Validate Unit if provided
	if crop.Unit != "" {
		if !entities.CROP_UNITS.IsValid(string(crop.Unit)) {
			c.JSON(http.StatusBadRequest, models.Response{
				StatusCode: http.StatusBadRequest,
				Success:    false,
				Message:    "Invalid crop unit",
				Data:       nil,
				Error: fmt.Sprintf("Invalid unit: %s. Valid values are: %v",
					crop.Unit, entities.CROP_UNITS),
				TimeStamp: time.Now().UTC().Format(time.RFC3339),
			})
			return
		}
	}

	// Ensure the ID in the path matches the ID in the body
	crop.Id = id

	if err := h.service.UpdateCrop(&crop); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to update crop",
			Data:       nil,
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Crop updated successfully",
		Data:       crop,
		Error:      nil,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}

// DeleteCrop handles DELETE /crops/:id
func (h *CropHandler) DeleteCrop(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteCrop(id); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to delete crop",
			Data:       nil,
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Crop deleted successfully",
		Data:       nil,
		Error:      nil,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}
