package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/Kisanlink/farmers-module/docs" // Import Swagger docs
	farmerReq "github.com/Kisanlink/farmers-module/internal/entities/requests"
	farmerResp "github.com/Kisanlink/farmers-module/internal/entities/responses/farmer"
	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gin-gonic/gin"
)

// @title Farmers Module API
// @version 1.0.0
// @description Farmers Module Service with Workflow-Based Architecture for Farm Management
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @Summary Get service info
// @Description Get information about the Farmers Module service
// @Tags service
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Summary Health check
// @Description Check the health status of the service
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func healthHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "service": "farmers-module"})
}

// @Summary Get identity endpoints info
// @Description Get information about identity and organization linkage endpoints
// @Tags identity
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /identity [get]
func identityHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Identity endpoints - W1-W3: Identity & Org Linkage"})
}

// @Summary Get KisanSathi endpoints info
// @Description Get information about KisanSathi assignment endpoints
// @Tags kisansathi
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /kisansathi [get]
func kisansathiHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "KisanSathi endpoints - W4-W5: KisanSathi Assignment"})
}

// @Summary Get farm endpoints info
// @Description Get information about farm management endpoints
// @Tags farms
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /farms [get]
func farmsHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Farm endpoints - W6-W9: Farm Management"})
}

// @Summary Get crop endpoints info
// @Description Get information about crop management endpoints
// @Tags crops
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /crops [get]
func cropsHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Crop endpoints - W10-W17: Crop Management"})
}

// @Summary Get admin endpoints info
// @Description Get information about admin and access control endpoints
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /admin [get]
func adminHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Admin endpoints - W18-W19: Access Control"})
}

// @Summary Get service info
// @Description Get information about the Farmers Module service
// @Tags service
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router / [get]
func rootHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Welcome to Farmers Module Server - Workflow-Based Architecture",
		"version": "1.0.0",
		"docs":    "/docs",
	})
}

func main() {
	// Initialize router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API v1 group
	api := router.Group("/api/v1")
	{
		// Health check endpoint
		api.GET("/health", healthHandler)

		// Identity & Organization Linkage (W1-W3)
		identity := api.Group("/identity")
		{
			identity.GET("", identityHandler)

			// Simple farmer management endpoints (in-memory for now)
			farmers := identity.Group("/farmers")
			{
				// @Summary Create a new farmer
				// @Description Create a new farmer profile
				// @Tags identity
				// @Accept json
				// @Produce json
				// @Param request body farmerReq.CreateFarmerRequest true "Create Farmer Request"
				// @Success 201 {object} farmerResp.FarmerResponse
				// @Router /identity/farmers [post]
				farmers.POST("", func(c *gin.Context) {
					var req farmerReq.CreateFarmerRequest
					if err := c.ShouldBindJSON(&req); err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
						return
					}

					// Create sample response using the proper response structure
					profileData := &farmerResp.FarmerProfileData{
						AAAUserID:        req.AAAUserID,
						AAAOrgID:         req.AAAOrgID,
						KisanSathiUserID: req.KisanSathiUserID,
						FirstName:        req.Profile.FirstName,
						LastName:         req.Profile.LastName,
						PhoneNumber:      req.Profile.PhoneNumber,
						Email:            req.Profile.Email,
						DateOfBirth:      req.Profile.DateOfBirth,
						Gender:           req.Profile.Gender,
						Address: farmerResp.AddressData{
							StreetAddress: req.Profile.Address.StreetAddress,
							City:          req.Profile.Address.City,
							State:         req.Profile.Address.State,
							PostalCode:    req.Profile.Address.PostalCode,
							Country:       req.Profile.Address.Country,
							Coordinates:   req.Profile.Address.Coordinates,
						},
						Preferences: req.Profile.Preferences,
						Metadata:    req.Profile.Metadata,
						CreatedAt:   "2024-01-01T00:00:00Z",
						UpdatedAt:   "2024-01-01T00:00:00Z",
					}

					response := farmerResp.NewFarmerResponse(profileData, "Farmer created successfully")
					response.SetRequestID(req.RequestID)
					c.JSON(http.StatusCreated, response)
				})

				// @Summary List farmers
				// @Description List farmers with filtering and pagination
				// @Tags identity
				// @Accept json
				// @Produce json
				// @Param request body farmerReq.ListFarmersRequest true "List Farmers Request"
				// @Success 200 {object} farmerResp.FarmerListResponse
				// @Router /identity/farmers [get]
				farmers.GET("", func(c *gin.Context) {
					var req farmerReq.ListFarmersRequest
					if err := c.ShouldBindQuery(&req); err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
						return
					}

					// Create sample response using the proper response structure
					sampleFarmers := []*farmerResp.FarmerProfileData{
						{
							AAAUserID:        "sample-user-1",
							AAAOrgID:         req.AAAOrgID,
							KisanSathiUserID: &req.KisanSathiUserID,
							FirstName:        "John",
							LastName:         "Doe",
							PhoneNumber:      "+1234567890",
							Email:            "john.doe@example.com",
							CreatedAt:        "2024-01-01T00:00:00Z",
							UpdatedAt:        "2024-01-01T00:00:00Z",
						},
					}

					response := farmerResp.NewFarmerListResponse(sampleFarmers, req.Page, req.PageSize, 1)
					response.SetRequestID(req.RequestID)
					c.JSON(http.StatusOK, response)
				})

				// @Summary Get farmer by ID
				// @Description Retrieve a farmer profile by AAA user ID and org ID
				// @Tags identity
				// @Accept json
				// @Produce json
				// @Param aaa_user_id path string true "AAA User ID"
				// @Param aaa_org_id path string true "AAA Org ID"
				// @Success 200 {object} farmerResp.FarmerProfileResponse
				// @Router /identity/farmers/{aaa_user_id}/{aaa_org_id} [get]
				farmers.GET("/:aaa_user_id/:aaa_org_id", func(c *gin.Context) {
					aaaUserID := c.Param("aaa_user_id")
					aaaOrgID := c.Param("aaa_org_id")

					// Create sample response using the proper response structure
					profileData := &farmerResp.FarmerProfileData{
						AAAUserID:        aaaUserID,
						AAAOrgID:         aaaOrgID,
						KisanSathiUserID: nil,
						FirstName:        "John",
						LastName:         "Doe",
						PhoneNumber:      "+1234567890",
						Email:            "john.doe@example.com",
						DateOfBirth:      "1990-01-01",
						Gender:           "MALE",
						Address: farmerResp.AddressData{
							StreetAddress: "123 Main St",
							City:          "Sample City",
							State:         "Sample State",
							PostalCode:    "12345",
							Country:       "Sample Country",
						},
						Preferences: map[string]string{
							"language": "en",
							"timezone": "UTC",
						},
						Metadata: map[string]string{
							"source": "manual_entry",
						},
						CreatedAt: "2024-01-01T00:00:00Z",
						UpdatedAt: "2024-01-01T00:00:00Z",
					}

					response := farmerResp.NewFarmerProfileResponse(profileData, "Farmer profile retrieved successfully")
					response.SetRequestID("req-" + aaaUserID + "-" + aaaOrgID)
					c.JSON(http.StatusOK, response)
				})
			}
		}

		// KisanSathi Assignment (W4-W5)
		kisansathi := api.Group("/kisansathi")
		{
			kisansathi.GET("", kisansathiHandler)
		}

		// Farm Management (W6-W9)
		farms := api.Group("/farms")
		{
			farms.GET("", farmsHandler)
		}

		// Crop Management (W10-W17)
		crops := api.Group("/crops")
		{
			crops.GET("", cropsHandler)
		}

		// Admin & Access Control (W18-W19)
		admin := api.Group("/admin")
		{
			admin.GET("", adminHandler)
		}
	}

	// Add root route
	router.GET("/", rootHandler)

	// Serve OpenAPI and Scalar-powered docs UI
	router.StaticFile("/docs/swagger.json", "docs/swagger.json")
	router.StaticFile("/docs/swagger.yaml", "docs/swagger.yaml")
	router.GET("/docs", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		specURL := scheme + "://" + c.Request.Host + "/docs/swagger.json"

		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecURL:  specURL,
			DarkMode: true,
			CustomOptions: scalar.CustomOptions{
				PageTitle: "Farmers Module API Reference",
			},
		})
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to render API docs: %v", err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
	})

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	log.Printf("Starting Farmers Module server on :%s", port)
	log.Println("Available workflow groups:")
	log.Println("  - /api/v1/identity     (W1-W3: Identity & Org Linkage)")
	log.Println("  - /api/v1/kisansathi   (W4-W5: KisanSathi Assignment)")
	log.Println("  - /api/v1/farms        (W6-W9: Farm Management)")
	log.Println("  - /api/v1/crops        (W10-W17: Crop Management)")
	log.Println("  - /api/v1/admin        (W18-W19: Access Control)")
	log.Println("  - /docs                (API Documentation)")

	err := router.Run(":" + port)
	if err != nil {
		log.Fatal("Error starting HTTP server:", err)
	}
}
