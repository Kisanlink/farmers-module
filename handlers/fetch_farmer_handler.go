package handlers

import(
	"net/http"
	"time"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/Kisanlink/farmers-module/models"

)

func (h *FarmerHandler) FetchFarmersHandler(c *gin.Context) {
	// Parse query parameters into filter struct
	filter := models.FarmerFilter{
		UserID:           getStringPtr(c.Query("user_id")),
		UserName:         getStringPtr(c.Query("username")),
		Email:           getStringPtr(c.Query("email")),
		CountryCode:     getStringPtr(c.Query("country_code")),
		MobileNumber:    getUint64Ptr(c.Query("mobile_number")),
		AadhaarNumber:   getStringPtr(c.Query("aadhaar_number")),
		KisansathiUserID: getStringPtr(c.Query("kisansathi_user_id")),
		IsActive:        getBoolPtr(c.Query("is_active")),
		CreatedAfter:    getTimePtr(c.Query("created_after")),
		CreatedBefore:   getTimePtr(c.Query("created_before")),
	}

	farmers, err := h.farmerService.FetchFarmers(filter)
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch farmers", err.Error())
		return
	}
	h.sendSuccessResponse(c, http.StatusOK, "Farmers fetched successfully", farmers)
}

// Helper functions for query parameter parsing
func getStringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func getUint64Ptr(value string) *uint64 {
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

func getBoolPtr(value string) *bool {
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil
	}
	return &parsed
}

func getTimePtr(value string) *time.Time {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return &parsed
}