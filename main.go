package main

import (
	"log"
	"os"

	"github.com/Kisanlink/farmers-module/database"
	"github.com/Kisanlink/farmers-module/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Initialize the database
	database.InitializeDatabase()

	// Initialize the router
	router := routes.Setup()

	// Add root route
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Farmer Module Server",
		})
	})

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable not set")
	}

	// Start the server
	log.Printf("Starting server on :%s", port)
	err = router.Run(":" + port)
	if err != nil {
		log.Fatal("Error starting HTTP server:", err)
	}
}
