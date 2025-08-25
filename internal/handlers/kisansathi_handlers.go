package handlers

import (
	"net/http"
	"strings"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AssignKisanSathi handles W4: Assign KisanSathi to farmer
// @Summary Assign KisanSathi to farmer
// @Description Assign a KisanSathi user to a specific farmer
// @Tags kisansathi
// @Accept json
// @Produce json
// @Param assignment body requests.AssignKisanSathiRequest true "KisanSathi assignment data"
// @Success 200 {object} responses.SwaggerKisanSathiAssignmentResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Router /kisansathi/assign [post]
func AssignKisanSathi(service services.FarmerLinkageService, logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.AssignKisanSathiRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Failed to bind request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		// Set request ID if not provided
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}

		logger.Info("Assigning KisanSathi to farmer",
			zap.String("request_id", req.RequestID),
			zap.String("aaa_user_id", req.AAAUserID),
			zap.String("kisan_sathi_user_id", req.KisanSathiUserID),
			zap.String("aaa_org_id", req.AAAOrgID))

		// Call the service
		assignmentData, err := service.AssignKisanSathi(c.Request.Context(), &req)
		if err != nil {
			logger.Error("Failed to assign KisanSathi", zap.Error(err), zap.String("request_id", req.RequestID))

			statusCode := http.StatusInternalServerError
			if isValidationError(err) {
				statusCode = http.StatusBadRequest
			} else if isPermissionError(err) {
				statusCode = http.StatusForbidden
			} else if isNotFoundError(err) {
				statusCode = http.StatusNotFound
			}

			c.JSON(statusCode, gin.H{
				"error":          err.Error(),
				"request_id":     req.RequestID,
				"correlation_id": req.RequestID,
			})
			return
		}

		// Convert to proper response format
		assignment, ok := assignmentData.(*responses.KisanSathiAssignmentData)
		if !ok {
			logger.Error("Invalid response format from service", zap.String("request_id", req.RequestID))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":          "invalid response format",
				"request_id":     req.RequestID,
				"correlation_id": req.RequestID,
			})
			return
		}

		response := responses.NewKisanSathiAssignmentResponse(assignment, "KisanSathi assigned successfully")
		response.SetRequestID(req.RequestID)

		logger.Info("KisanSathi assigned successfully",
			zap.String("request_id", req.RequestID),
			zap.String("aaa_user_id", req.AAAUserID),
			zap.String("kisan_sathi_user_id", req.KisanSathiUserID),
			zap.String("aaa_org_id", req.AAAOrgID))

		c.JSON(http.StatusOK, response)
	}
}

// ReassignOrRemoveKisanSathi handles W5: Reassign or remove KisanSathi
// @Summary Reassign or remove KisanSathi
// @Description Reassign a KisanSathi to a different farmer or remove the assignment
// @Tags kisansathi
// @Accept json
// @Produce json
// @Param assignment body requests.ReassignKisanSathiRequest true "KisanSathi reassignment data"
// @Success 200 {object} responses.SwaggerKisanSathiAssignmentResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Router /kisansathi/reassign [put]
func ReassignOrRemoveKisanSathi(service services.FarmerLinkageService, logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.ReassignKisanSathiRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Failed to bind request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		// Set request ID if not provided
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}

		action := "reassign"
		if req.NewKisanSathiUserID == nil {
			action = "remove"
		}

		logger.Info("Reassigning or removing KisanSathi",
			zap.String("request_id", req.RequestID),
			zap.String("action", action),
			zap.String("aaa_user_id", req.AAAUserID),
			zap.String("aaa_org_id", req.AAAOrgID))

		// Call the service
		assignmentData, err := service.ReassignOrRemoveKisanSathi(c.Request.Context(), &req)
		if err != nil {
			logger.Error("Failed to reassign/remove KisanSathi", zap.Error(err), zap.String("request_id", req.RequestID))

			statusCode := http.StatusInternalServerError
			if isValidationError(err) {
				statusCode = http.StatusBadRequest
			} else if isPermissionError(err) {
				statusCode = http.StatusForbidden
			} else if isNotFoundError(err) {
				statusCode = http.StatusNotFound
			}

			c.JSON(statusCode, gin.H{
				"error":          err.Error(),
				"request_id":     req.RequestID,
				"correlation_id": req.RequestID,
			})
			return
		}

		// Convert to proper response format
		assignment, ok := assignmentData.(*responses.KisanSathiAssignmentData)
		if !ok {
			logger.Error("Invalid response format from service", zap.String("request_id", req.RequestID))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":          "invalid response format",
				"request_id":     req.RequestID,
				"correlation_id": req.RequestID,
			})
			return
		}

		message := "KisanSathi reassigned successfully"
		if req.NewKisanSathiUserID == nil {
			message = "KisanSathi removed successfully"
		}

		response := responses.NewKisanSathiAssignmentResponse(assignment, message)
		response.SetRequestID(req.RequestID)

		logger.Info("KisanSathi operation completed successfully",
			zap.String("request_id", req.RequestID),
			zap.String("action", action),
			zap.String("aaa_user_id", req.AAAUserID),
			zap.String("aaa_org_id", req.AAAOrgID))

		c.JSON(http.StatusOK, response)
	}
}

// CreateKisanSathiUser handles creating a new KisanSathi user
// @Summary Create KisanSathi user
// @Description Create a new KisanSathi user with role assignment
// @Tags kisansathi
// @Accept json
// @Produce json
// @Param user body requests.CreateKisanSathiUserRequest true "KisanSathi user data"
// @Success 201 {object} responses.SwaggerKisanSathiUserResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 409 {object} responses.SwaggerErrorResponse
// @Router /kisansathi/create-user [post]
func CreateKisanSathiUser(service services.FarmerLinkageService, logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.CreateKisanSathiUserRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Failed to bind request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		// Set request ID if not provided
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}

		logger.Info("Creating KisanSathi user",
			zap.String("request_id", req.RequestID),
			zap.String("username", req.Username),
			zap.String("phone_number", req.PhoneNumber),
			zap.String("email", req.Email))

		// Call the service
		userData, err := service.CreateKisanSathiUser(c.Request.Context(), &req)
		if err != nil {
			logger.Error("Failed to create KisanSathi user", zap.Error(err), zap.String("request_id", req.RequestID))

			statusCode := http.StatusInternalServerError
			if isValidationError(err) {
				statusCode = http.StatusBadRequest
			} else if isPermissionError(err) {
				statusCode = http.StatusForbidden
			} else if strings.Contains(strings.ToLower(err.Error()), "already exists") {
				statusCode = http.StatusConflict
			}

			c.JSON(statusCode, gin.H{
				"error":          err.Error(),
				"request_id":     req.RequestID,
				"correlation_id": req.RequestID,
			})
			return
		}

		// Convert to proper response format
		user, ok := userData.(*responses.KisanSathiUserData)
		if !ok {
			logger.Error("Invalid response format from service", zap.String("request_id", req.RequestID))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":          "invalid response format",
				"request_id":     req.RequestID,
				"correlation_id": req.RequestID,
			})
			return
		}

		response := responses.NewKisanSathiUserResponse(user, "KisanSathi user created successfully")
		response.SetRequestID(req.RequestID)

		logger.Info("KisanSathi user created successfully",
			zap.String("request_id", req.RequestID),
			zap.String("user_id", user.ID),
			zap.String("username", user.Username))

		c.JSON(http.StatusCreated, response)
	}
}

// GetKisanSathiAssignment handles getting KisanSathi assignment
// @Summary Get KisanSathi assignment
// @Description Retrieve the KisanSathi assignment for a specific farmer
// @Tags kisansathi
// @Accept json
// @Produce json
// @Param farmer_id path string true "Farmer ID"
// @Param org_id path string true "Organization ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /kisansathi/assignment/{farmer_id}/{org_id} [get]
func GetKisanSathiAssignment(service services.FarmerLinkageService, logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		farmerID := c.Param("farmer_id")
		orgID := c.Param("org_id")

		if farmerID == "" || orgID == "" {
			logger.Error("Missing required parameters")
			c.JSON(http.StatusBadRequest, gin.H{"error": "farmer_id and org_id are required"})
			return
		}

		logger.Info("Getting KisanSathi assignment",
			zap.String("farmer_id", farmerID),
			zap.String("org_id", orgID))

		// Call the service to get farmer linkage which includes KisanSathi assignment
		linkageData, err := service.GetFarmerLinkage(c.Request.Context(), farmerID, orgID)
		if err != nil {
			logger.Error("Failed to get KisanSathi assignment", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Info("KisanSathi assignment retrieved successfully",
			zap.String("farmer_id", farmerID),
			zap.String("org_id", orgID))

		c.JSON(http.StatusOK, gin.H{
			"message": "KisanSathi assignment retrieved successfully",
			"data":    linkageData,
		})
	}
}
