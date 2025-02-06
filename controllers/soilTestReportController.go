package controllers

import (
	"context"
	"net/http"
	"time"
	"strconv"
	
"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/gin-gonic/gin"
)

type SoilTestReportController struct {
	SoilTestReportRepo *repositories.SoilTestReportRepository
}

func NewSoilTestReportController(repo *repositories.SoilTestReportRepository) *SoilTestReportController {
	return &SoilTestReportController{
		SoilTestReportRepo: repo,
	}
}

// GetSoilTestReports retrieves soil test reports for a given farmer ID and/or farm ID
func (strc *SoilTestReportController) GetSoilTestReports(c *gin.Context) {
	// Extract query parameters for farmerId and farmId
	farmerIDStr := c.DefaultQuery("farmerId", "")
	farmIDStr := c.DefaultQuery("farmId", "")

	var farmerID, farmID int
	var err error

	// Convert farmerId to integer if present
	if farmerIDStr != "" {
		farmerID, err = strconv.Atoi(farmerIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID"})
			return
		}
	}

	// Convert farmId to integer if present
	if farmIDStr != "" {
		farmID, err = strconv.Atoi(farmIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farm ID"})
			return
		}
	}

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch reports from repository based on available query parameters
	var soilTestReports []models.SoilTestReport
	if farmID != 0 {
		soilTestReports, err = strc.SoilTestReportRepo.GetSoilTestReportsByFarmID(ctx, farmID)
	} else if farmerID != 0 {
		soilTestReports, err = strc.SoilTestReportRepo.GetSoilTestReportsByFarmerID(ctx, farmerID)
	} else {
		soilTestReports, err = strc.SoilTestReportRepo.GetAllSoilTestReports(ctx)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch soil test reports"})
		return
	}

	// If no reports found
	if len(soilTestReports) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No soil test reports found"})
		return
	}

	// Return soil test reports
	c.JSON(http.StatusOK, soilTestReports)
}
