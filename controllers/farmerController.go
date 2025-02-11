package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"
  "github.com/Kisanlink/farmers-module/repositories"
	"github.com/Kisanlink/farmers-module/models"
  "github.com/gin-gonic/gin"
)

type FarmerController struct {
	Repository *repositories.FarmerRepository
}

func NewFarmerController(repo *repositories.FarmerRepository) *FarmerController {
	return &FarmerController{Repository: repo}
}
func (fc *FarmerController) CreateFarmer(c *gin.Context) {
	var farmer models.Farmer
	if err := c.ShouldBindJSON(&farmer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer data"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create farmer in the database
	err := fc.Repository.CreateFarmer(ctx, &farmer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create farmer"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Farmer created successfully",
		"data": gin.H{
			"farmerId": farmer.FarmerID, // Assuming FarmerID is a unique identifier
		},
	})
}

func (fc *FarmerController) UpdateFarmer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID"})
		return
	}

	var updatedFarmer models.Farmer
	if err := c.ShouldBindJSON(&updatedFarmer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer data"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Update the farmer in the database
	err = fc.Repository.UpdateFarmer(ctx, id, &updatedFarmer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update farmer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Farmer updated successfully",
	})
}

func (fc *FarmerController) SearchFarmers(c *gin.Context) {
	filter := c.DefaultQuery("filter", "")
	orderBy := c.DefaultQuery("orderBy", "firstName")
	pageNumber := c.DefaultQuery("pageNumber", "1")
	perPage := c.DefaultQuery("perPage", "10")

	page, _ := strconv.Atoi(pageNumber)
	perPageCount, _ := strconv.Atoi(perPage)

	// Convert perPageCount to int64 to match totalCount type
	perPageCountInt64 := int64(perPageCount)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	farmers, totalCount, err := fc.Repository.SearchFarmers(ctx, filter, orderBy, page, perPageCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch farmers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Successfully fetched farmers",
		"statusCode": 200,
		"data":       farmers,
		"pagination": gin.H{
			"total_count": totalCount,
			"per_page":    perPageCountInt64, // Use the int64 converted value
			"page_number": page,
			"total_pages": (totalCount + perPageCountInt64 - 1) / perPageCountInt64, // Adjust for the type mismatch
		},
	})
}
func (fc *FarmerController) GetFarmerPersonalDetailsByID(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID"})
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    farmer, err := fc.Repository.GetFarmerByID(ctx, id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err})
        return
    }

    response := map[string]interface{}{
        "id":                farmer.ID.Hex(),
        "farmedID":          farmer.FarmerID,
        "firstName":         farmer.FirstName,
        "lastName":          farmer.LastName,
        "city":              farmer.City,
        "state":             farmer.State,
        "age":               farmer.Age,
        "district":          farmer.District,
        "pincode":           farmer.Pincode,
        "mobileNumber":      farmer.MobileNumber,
        "kisansathiName":    farmer.KisansathiName,
        "shares":            farmer.Shares,
        "areaManagerName":   farmer.AreaManagerName,
        "areaManagerId":     farmer.AreaManagerID,
        "totalWalletAmount": farmer.TotalWalletAmount,
    }

    c.JSON(http.StatusOK, response)
}

