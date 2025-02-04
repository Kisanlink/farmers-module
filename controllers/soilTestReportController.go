package controllers

import (
	"context"
	"net/http"
	"time"
	"strconv"

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

// GetSoilTestReports retrieves soil test reports for a given farm ID
func (strc *SoilTestReportController) GetSoilTestReports(c *gin.Context) {
	// Extract farm ID from request parameters
	farmIDStr := c.Param("farmId")
	farmID, err := strconv.Atoi(farmIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farm ID"})
		return
	}

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch reports from repository
	soilTestReports, err := strc.SoilTestReportRepo.GetSoilTestReportsByFarmID(ctx, farmID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch soil test reports"})
		return
	}

	// If no reports found
	if len(soilTestReports) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No soil test reports found for this farm"})
		return
	}

	// Return soil test reports
	c.JSON(http.StatusOK, soilTestReports)
}
