package handlers

import (
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FPOLifecycleHandler handles FPO lifecycle HTTP requests
type FPOLifecycleHandler struct {
	service *services.FPOLifecycleService
	logger  interfaces.Logger
}

// NewFPOLifecycleHandler creates a new FPO lifecycle handler
func NewFPOLifecycleHandler(service *services.FPOLifecycleService, logger interfaces.Logger) *FPOLifecycleHandler {
	return &FPOLifecycleHandler{
		service: service,
		logger:  logger,
	}
}

// SyncFromAAA synchronizes FPO reference from AAA service
// @Summary Sync FPO from AAA
// @Description Synchronize FPO reference from AAA service to local database
// @Tags FPO Lifecycle
// @Accept json
// @Produce json
// @Param aaa_org_id path string true "AAA Organization ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /identity/fpo/sync/{aaa_org_id} [post]
func (h *FPOLifecycleHandler) SyncFromAAA(c *gin.Context) {
	requestID := c.GetString("request_id")
	aaaOrgID := c.Param("aaa_org_id")

	h.logger.Info("Syncing FPO from AAA",
		zap.String("request_id", requestID),
		zap.String("aaa_org_id", aaaOrgID),
	)

	if aaaOrgID == "" {
		h.logger.Error("Missing AAA organization ID",
			zap.String("request_id", requestID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "AAA organization ID is required",
		})
		return
	}

	// Sync FPO from AAA
	fpoRef, err := h.service.SyncFPOFromAAA(c.Request.Context(), aaaOrgID)
	if err != nil {
		h.logger.Error("Failed to sync FPO from AAA",
			zap.String("request_id", requestID),
			zap.String("aaa_org_id", aaaOrgID),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	h.logger.Info("FPO synchronized successfully",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoRef.ID),
		zap.String("aaa_org_id", aaaOrgID),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "FPO synchronized successfully",
		"data": map[string]interface{}{
			"fpo_id":              fpoRef.ID,
			"aaa_org_id":          fpoRef.AAAOrgID,
			"name":                fpoRef.Name,
			"registration_number": fpoRef.RegistrationNo,
			"status":              fpoRef.Status.String(),
			"created_at":          fpoRef.CreatedAt.Format(time.RFC3339),
			"updated_at":          fpoRef.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// RetrySetup retries failed FPO setup
// @Summary Retry Failed Setup
// @Description Retry setup operations for a failed FPO
// @Tags FPO Lifecycle
// @Accept json
// @Produce json
// @Param id path string true "FPO ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /identity/fpo/{id}/retry-setup [post]
func (h *FPOLifecycleHandler) RetrySetup(c *gin.Context) {
	requestID := c.GetString("request_id")
	fpoID := c.Param("id")

	h.logger.Info("Retrying FPO setup",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
	)

	if fpoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "FPO ID is required",
		})
		return
	}

	err := h.service.RetryFailedSetup(c.Request.Context(), fpoID)
	if err != nil {
		h.logger.Error("Failed to retry FPO setup",
			zap.String("request_id", requestID),
			zap.String("fpo_id", fpoID),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	h.logger.Info("FPO setup retry initiated",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "FPO setup retry initiated successfully",
	})
}

// SuspendFPO suspends an FPO
// @Summary Suspend FPO
// @Description Suspend an active FPO
// @Tags FPO Lifecycle
// @Accept json
// @Produce json
// @Param id path string true "FPO ID"
// @Param request body requests.SuspendFPORequest true "Suspend Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /identity/fpo/{id}/suspend [put]
func (h *FPOLifecycleHandler) SuspendFPO(c *gin.Context) {
	requestID := c.GetString("request_id")
	fpoID := c.Param("id")

	var req requests.SuspendFPORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	h.logger.Info("Suspending FPO",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
		zap.String("reason", req.Reason),
	)

	err := h.service.SuspendFPO(c.Request.Context(), fpoID, req.Reason)
	if err != nil {
		h.logger.Error("Failed to suspend FPO",
			zap.String("request_id", requestID),
			zap.String("fpo_id", fpoID),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	h.logger.Info("FPO suspended successfully",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "FPO suspended successfully",
	})
}

// ReactivateFPO reactivates a suspended FPO
// @Summary Reactivate FPO
// @Description Reactivate a suspended FPO
// @Tags FPO Lifecycle
// @Accept json
// @Produce json
// @Param id path string true "FPO ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /identity/fpo/{id}/reactivate [put]
func (h *FPOLifecycleHandler) ReactivateFPO(c *gin.Context) {
	requestID := c.GetString("request_id")
	fpoID := c.Param("id")

	h.logger.Info("Reactivating FPO",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
	)

	err := h.service.ReactivateFPO(c.Request.Context(), fpoID)
	if err != nil {
		h.logger.Error("Failed to reactivate FPO",
			zap.String("request_id", requestID),
			zap.String("fpo_id", fpoID),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	h.logger.Info("FPO reactivated successfully",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "FPO reactivated successfully",
	})
}

// DeactivateFPO deactivates an FPO
// @Summary Deactivate FPO
// @Description Permanently deactivate an FPO
// @Tags FPO Lifecycle
// @Accept json
// @Produce json
// @Param id path string true "FPO ID"
// @Param request body requests.DeactivateFPORequest true "Deactivate Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /identity/fpo/{id}/deactivate [delete]
func (h *FPOLifecycleHandler) DeactivateFPO(c *gin.Context) {
	requestID := c.GetString("request_id")
	fpoID := c.Param("id")

	var req requests.DeactivateFPORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	h.logger.Info("Deactivating FPO",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
		zap.String("reason", req.Reason),
	)

	err := h.service.DeactivateFPO(c.Request.Context(), fpoID, req.Reason)
	if err != nil {
		h.logger.Error("Failed to deactivate FPO",
			zap.String("request_id", requestID),
			zap.String("fpo_id", fpoID),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	h.logger.Info("FPO deactivated successfully",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "FPO deactivated successfully",
	})
}

// GetHistory retrieves FPO audit history
// @Summary Get FPO History
// @Description Get audit history for an FPO
// @Tags FPO Lifecycle
// @Accept json
// @Produce json
// @Param id path string true "FPO ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /identity/fpo/{id}/history [get]
func (h *FPOLifecycleHandler) GetHistory(c *gin.Context) {
	requestID := c.GetString("request_id")
	fpoID := c.Param("id")

	h.logger.Info("Getting FPO history",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
	)

	history, err := h.service.GetFPOHistory(c.Request.Context(), fpoID)
	if err != nil {
		h.logger.Error("Failed to get FPO history",
			zap.String("request_id", requestID),
			zap.String("fpo_id", fpoID),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	// Convert to response format
	historyData := make([]map[string]interface{}, len(history))
	for i, entry := range history {
		historyData[i] = map[string]interface{}{
			"action":         entry.Action,
			"previous_state": entry.PreviousState,
			"new_state":      entry.NewState,
			"reason":         entry.Reason,
			"performed_by":   entry.PerformedBy,
			"performed_at":   entry.PerformedAt.Format(time.RFC3339),
			"details":        entry.Details,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    historyData,
	})
}

// GetFPOByAAAOrgID retrieves FPO by AAA organization ID (with auto-sync)
// @Summary Get FPO by AAA Org ID
// @Description Get FPO reference by AAA organization ID, syncs from AAA if not found locally
// @Tags FPO Lifecycle
// @Accept json
// @Produce json
// @Param aaa_org_id path string true "AAA Organization ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /identity/fpo/by-org/{aaa_org_id} [get]
func (h *FPOLifecycleHandler) GetFPOByAAAOrgID(c *gin.Context) {
	requestID := c.GetString("request_id")
	aaaOrgID := c.Param("aaa_org_id")

	h.logger.Info("Getting FPO by AAA org ID",
		zap.String("request_id", requestID),
		zap.String("aaa_org_id", aaaOrgID),
	)

	// Use GetOrSyncFPO to automatically sync if not found
	fpoRef, err := h.service.GetOrSyncFPO(c.Request.Context(), aaaOrgID)
	if err != nil {
		h.logger.Error("Failed to get FPO",
			zap.String("request_id", requestID),
			zap.String("aaa_org_id", aaaOrgID),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"fpo_id":              fpoRef.ID,
			"aaa_org_id":          fpoRef.AAAOrgID,
			"name":                fpoRef.Name,
			"registration_number": fpoRef.RegistrationNo,
			"status":              fpoRef.Status.String(),
			"business_config":     fpoRef.BusinessConfig,
			"metadata":            fpoRef.Metadata,
			"created_at":          fpoRef.CreatedAt.Format(time.RFC3339),
			"updated_at":          fpoRef.UpdatedAt.Format(time.RFC3339),
		},
	})
}
