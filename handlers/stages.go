package handlers

import (
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StageHandler struct {
	service services.StageServiceInterface
}

func NewStageHandler(service services.StageServiceInterface) *StageHandler {
	return &StageHandler{service: service}
}

// CreateStage handles POST /stages
func (h *StageHandler) CreateStage(c *gin.Context) {
	var stage models.Stage
	if err := c.ShouldBindJSON(&stage); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Invalid input data",
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	if err := h.service.CreateStage(&stage); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to create stage",
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		StatusCode: http.StatusCreated,
		Success:    true,
		Message:    "Stage created successfully",
		Data:       stage,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}

// GetAllStages handles GET /stages
func (h *StageHandler) GetAllStages(c *gin.Context) {
	stages, err := h.service.GetAllStages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to get stages",
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Stages retrieved successfully",
		Data:       stages,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}

// GetStageByID handles GET /stages/:id
func (h *StageHandler) GetStageByID(c *gin.Context) {
	id := c.Param("id")
	stage, err := h.service.GetStageByID(id)

	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "Failed to retrieve stage"
		if err == gorm.ErrRecordNotFound {
			statusCode = http.StatusNotFound
			message = "Stage not found"
		}
		c.JSON(statusCode, models.Response{
			StatusCode: statusCode,
			Success:    false,
			Message:    message,
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Stage retrieved successfully",
		Data:       stage,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}

// UpdateStage handles PUT /stages/:id
func (h *StageHandler) UpdateStage(c *gin.Context) {
	id := c.Param("id")
	var stage models.Stage
	if err := c.ShouldBindJSON(&stage); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Invalid input data",
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	stage.Id = id // Ensure we are updating the correct record
	if err := h.service.UpdateStage(&stage); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to update stage",
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Stage updated successfully",
		Data:       stage,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}

// DeleteStage handles DELETE /stages/:id
func (h *StageHandler) DeleteStage(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteStage(id); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to delete stage",
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Stage deleted successfully",
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}
