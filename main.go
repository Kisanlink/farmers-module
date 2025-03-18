package main

import (
	"log"

	"github.com/Kisanlink/farmers-module/database"
	"github.com/Kisanlink/farmers-module/routes"
)

func main() {
	

	// Step 1: Initialize the database
	database.InitializeDatabase()

	// Step 2: Initialize the router
	router := routes.Setup()

	// Step 3: Start the server
	err := router.Run(":8080")
	if err != nil {
		log.Fatal("Error starting HTTP server:", err)
	}
}
