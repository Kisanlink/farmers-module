package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Kisanlink/farmers-module/repositories"

	"github.com/gin-gonic/gin"
)

type FarmerController struct {
	Repository *repositories.FarmerRepository
}

func NewFarmerController(repo *repositories.FarmerRepository) *FarmerController {
	return &FarmerController{Repository: repo}
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
	// Exclude certain fields dynamically
	response := map[string]interface{}{
		"id":        farmer.ID.Hex(),
		"farmedID":  farmer.FarmerID,
		"firstName": farmer.FirstName,
		"lastName":  farmer.LastName,
		"city":      farmer.City,
		"state":     farmer.State,
		"age":       farmer.Age,
		"district":  farmer.District, 
		"pincode":   farmer.Pincode,
		"mobileNumber":farmer.MobileNumber,
		"kisansathiName": farmer.KisansathiName,
		"shares":     farmer.Shares,
		"areaManagerName"	:farmer.AreaManagerName,
	  "areaManagerId":farmer.AreaManagerID, 
	 "totalWalletAmount":farmer.TotalWalletAmount, 
		}

	c.JSON(http.StatusOK, response)
}
