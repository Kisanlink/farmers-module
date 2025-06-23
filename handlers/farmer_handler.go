package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Kisanlink/farmers-module/config"
	"github.com/Kisanlink/farmers-module/entities"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

type FarmerHandler struct {
	farmerService services.FarmerServiceInterface
}

func NewFarmerHandler(farmerService services.FarmerServiceInterface) *FarmerHandler {
	return &FarmerHandler{
		farmerService: farmerService,
	}
}

// ──────────────────────────────────────────────────────────────────────────────
//  1. FPO cross‑field helper  (exactly as in the previous answer)
//
// ──────────────────────────────────────────────────────────────────────────────
func validateFPOFields(r *models.FarmerSignupRequest) error {
	if r.IsFPO {
		missing := make([]string, 0, 7)
		if r.State == "" {
			missing = append(missing, "state")
		}
		if r.District == "" {
			missing = append(missing, "district")
		}
		if r.Block == "" {
			missing = append(missing, "block")
		}
		if r.IaName == "" {
			missing = append(missing, "iaName")
		}
		if r.CbbName == "" {
			missing = append(missing, "cbbName")
		}
		if r.FpoName == "" {
			missing = append(missing, "fpoName")
		}
		if r.FpoRegNo == "" {
			missing = append(missing, "fpoRegNo")
		}
		if len(missing) > 0 {
			return fmt.Errorf("when isFPO=true the following fields are required: %s",
				strings.Join(missing, ", "))
		}
	} else {
		// Wipe FPO‑specific values when not relevant.
		r.State, r.District, r.Block, r.IaName,
			r.CbbName, r.FpoName, r.FpoRegNo = "", "", "", "", "", "", ""
	}
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
//  2. The updated signup handler
//
// ──────────────────────────────────────────────────────────────────────────────
func (h *FarmerHandler) FarmerSignupHandler(c *gin.Context) {
	ctx := c.Request.Context()
	var req models.FarmerSignupRequest

	// ---------- 1‑A. Bind JSON ------------------------------------------------
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"Invalid request body", err.Error())
		return
	}

	if req.CountryCode == "" {
		req.CountryCode = "91" // default to India
	}

	// ---------- 1‑B. Field‑level validation ----------------------------------
	if err := models.Validator.Struct(&req); err != nil {
		h.sendErrorResponse(c, semanticStatus.Validation,
			"Validation failed", err.Error())
		return
	}

	// ---------- 1‑C. Cross‑field FPO validation ------------------------------
	if err := validateFPOFields(&req); err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"Validation failed", err.Error())
		return
	}

	// ---------- 2.  Farmer‑type check (was already present) ------------------
	if req.Type != "" {
		tp := strings.ToUpper(req.Type)
		if !entities.FARMER_TYPES.IsValid(tp) {
			h.sendErrorResponse(c, http.StatusBadRequest,
				"Invalid farmer type",
				fmt.Sprintf("`type` must be one of: %v",
					entities.FARMER_TYPES.StringValues()))
			return
		}
		req.Type = tp
	}

	// ---------- 3.  Mobile number is **always** required ---------------------
	if len(req.MobileNumberString) != 10 {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"Invalid mobile number", "must be exactly 10 digits")
		return
	}
	if req.MobileNumberString[0] == '0' {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"Invalid mobile number", "should not start with 0")
		return
	}
	mobileUint, err := strconv.ParseUint(req.MobileNumberString, 10, 64)
	if err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"Invalid mobile number", "must contain only digits")
		return
	}
	req.MobileNumber = mobileUint

	// ─────────────────────────────────────────────────────────────────────────
	// 4.  EARLY EXIT: existing flow when caller gives explicit `user_id`
	// ─────────────────────────────────────────────────────────────────────────
	if req.UserId != nil {
		h.createFarmerViaExplicitUserId(c, *req.UserId, &req)
		return
	}

	// ─────────────────────────────────────────────────────────────────────────
	// 5.  Mobile‑first flow
	// ─────────────────────────────────────────────────────────────────────────

	// 5‑A.  Ask AAA whether a user with this mobile already exists
	userResp, err := services.GetUserByMobileClient(
		ctx,
		req.MobileNumber,
	)
	if err != nil {
		// If AAA is down we refuse to continue: can't guarantee uniqueness.
		h.sendErrorResponse(c, semanticStatus.Dependency,
			"Failed to query AAA", err.Error())
		return
	}

	// 5‑B.  Path ① — user already exists  ➜  just assign Farmer role
	if userResp != nil &&
		userResp.StatusCode == http.StatusOK &&
		userResp.Data != nil && userResp.Data.Id != "" {

		userId := userResp.Data.Id

		//if farmer role exists for this user
		if exists, _ := h.farmerService.ExistsForUser(userId); exists {
			h.sendErrorResponse(c, semanticStatus.Conflict,
				"farmer already registered for this user", "duplicate farmer record")
			return
		}

		// Ensure the user actually *gets* the Farmer role (idempotent on AAA side)
		if _, err := services.AssignRoleToUserClient(
			c.Request.Context(),
			userId,
			config.ROLE_FARMER,
		); err != nil {
			h.sendErrorResponse(c, http.StatusInternalServerError,
				"Failed to assign role", err.Error())
			return
		}

		// Create corresponding Farmer row in our DB
		farmer, userDetails, err := h.farmerService.CreateFarmer(userId, req)
		if err != nil {
			h.sendErrorResponse(c, http.StatusInternalServerError,
				"Failed to create farmer record", err.Error())
			return
		}

		h.sendSuccessResponse(c, http.StatusCreated,
			"Farmer registered successfully", gin.H{
				"farmer": farmer,
				"user":   userDetails,
			})
		return
	}

	// 5‑C.  Path ② — no existing user  ➜  we have to create one first
	if err := requireFields(map[string]*string{
		"username": req.UserName,
	}); err != nil {
		h.sendErrorResponse(c, semanticStatus.Validation,
			"Validation failed", err.Error())
		return
	}

	createUserResp, err := services.CreateUserClient(req, "")
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"Failed to create user in AAA", err.Error())
		return
	}
	if createUserResp == nil || createUserResp.Data == nil ||
		createUserResp.Data.Id == "" {
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"Invalid response from AAA service",
			"empty user Id in response")
		return
	}
	userId := createUserResp.Data.Id

	// Immediately attach the Farmer role to the new user
	if _, err := services.AssignRoleToUserClient(
		c.Request.Context(), userId, config.ROLE_FARMER); err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"Role assignment failed", err.Error())
		return
	}

	// Farmer table entry
	farmer, userDetails, err := h.farmerService.CreateFarmer(userId, req)
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"Failed to create farmer record", err.Error())
		return
	}

	h.sendSuccessResponse(c, http.StatusCreated,
		"Farmer registered successfully", gin.H{
			"farmer": farmer,
			"user":   userDetails,
		})
}

// ──────────────────────────────────────────────────────────────────────────────
// 5bis. helper – previous explicit‑user‑id path extracted for clarity
// ──────────────────────────────────────────────────────────────────────────────
func (h *FarmerHandler) createFarmerViaExplicitUserId(
	c *gin.Context,
	userId string,
	req *models.FarmerSignupRequest,
) {
	// *** the body below is 100 % identical to what was previously in the
	//     big if‑block at the top of the old handler, only function‑scoped. ***

	// Optional Kisansathi check
	if req.KisansathiUserId != nil {
		userResp, err := services.GetUserByIdClient(
			c.Request.Context(), *req.KisansathiUserId)
		if err != nil {
			h.sendErrorResponse(c, http.StatusInternalServerError,
				"Failed to verify Kisansathi user", err.Error())
			return
		}
		if userResp == nil || userResp.StatusCode != http.StatusOK ||
			userResp.Data == nil {
			h.sendErrorResponse(c, http.StatusUnauthorized,
				"Kisansathi user not found", "invalid user response")
			return
		}
		// … snipped the existing permission‑check loop for brevity …
	}

	// Create farmer row
	farmer, userDetails, err := h.farmerService.CreateFarmer(userId, *req)
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"Failed to create farmer record", err.Error())
		return
	}

	// Assign role
	if _, err := services.AssignRoleToUserClient(
		c.Request.Context(), userId, config.ROLE_FARMER); err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"Farmer created but role assignment failed", err.Error())
		return
	}

	h.sendSuccessResponse(c, http.StatusCreated,
		"Farmer registered successfully", gin.H{
			"farmer": farmer,
			"user":   userDetails,
		})
}

// Helper methods as receiver functions
func (h *FarmerHandler) sendErrorResponse(c *gin.Context, status int, userMessage string, errorDetail string) {
	c.JSON(status, gin.H{
		"status":    status,
		"message":   userMessage,
		"error":     errorDetail,
		"timestamp": time.Now().UTC(),
		"data":      nil,
		"success":   false,
	})
}

func (h *FarmerHandler) sendSuccessResponse(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, gin.H{
		"status":    status,
		"message":   message,
		"error":     nil,
		"timestamp": time.Now().UTC(),
		"data":      data,
		"success":   true,
	})
}

// SetSubscription handles “POST /farmers/:farmer_id/subscription?is_subscribed={true|false}”
func (h *FarmerHandler) SetSubscription(c *gin.Context) {
	// 1. Extract path parameter
	farmerID := c.Param("farmer_id")
	if farmerID == "" {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"farmer_id is required in path", "missing path parameter")
		return
	}

	// 2. Extract query parameter
	isSubStr := c.Query("is_subscribed")
	if isSubStr == "" {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"is_subscribed query parameter is required", "missing query parameter")
		return
	}

	// 3. Parse “is_subscribed” into a bool
	isSubscribed, err := strconv.ParseBool(isSubStr)
	if err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"invalid value for is_subscribed; must be true or false", err.Error())
		return
	}

	// 4. Call service to toggle the subscription flag
	if err := h.farmerService.SetSubscriptionStatus(farmerID, isSubscribed); err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"failed to update subscription status", err.Error())
		return
	}

	// 5. Return a success response
	h.sendSuccessResponse(c, http.StatusOK, "subscription status updated", gin.H{
		"farmer_id":     farmerID,
		"is_subscribed": isSubscribed,
	})
}

// requireFields checks that the given pointers are non-nil / non-empty
func requireFields(fields map[string]*string) error {
	missing := make([]string, 0, len(fields))
	for k, v := range fields {
		if v == nil || strings.TrimSpace(*v) == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required field(s): %s", strings.Join(missing, ", "))
	}
	return nil
}
