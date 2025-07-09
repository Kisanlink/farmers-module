package handlers

import (
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

type KisansathiHandler struct {
	svc services.KisansathiServiceInterface
}

func NewKisansathiHandler(s services.KisansathiServiceInterface) *KisansathiHandler {
	return &KisansathiHandler{svc: s}
}

// GET /kisansathis?fpo_reg_no=123
func (h *KisansathiHandler) List(c *gin.Context) {
	fpoRegNo := c.Query("fpo_reg_no")

	users, err := h.svc.ListKisansathis(fpoRegNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    http.StatusInternalServerError,
			"success":   false,
			"message":   "Failed to fetch Kisansathi users",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
			"data":      nil,
		})
		return
	}
	if len(users) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":    http.StatusOK,
			"success":   true,
			"message":   "No Kisansathi users found",
			"error":     nil,
			"timestamp": time.Now().UTC(),
			"data":      []interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    http.StatusOK,
		"success":   true,
		"message":   "Kisansathi users fetched successfully",
		"error":     nil,
		"timestamp": time.Now().UTC(),
		"data":      users,
	})
}
