package main

import (
	// "log"
	// "net/http"
	"github.com/gin-gonic/gin"
	"github.com/Kisanlink/farmers-module/config" // Import the config package
	// Import the routes package
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080

	// Initialize MongoDB connection
	config.InitDB()

	// // Set up routes for the application
	// // routes.SetupRoutes()

	// // Start the web server on port 8080
	// log.Print("Starting server on :8080...")
	// err := http.ListenAndServe(":8080", nil)
	// if err != nil {
	// 	log.Fatal("Error starting server:", err)
}
