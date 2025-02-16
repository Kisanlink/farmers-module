package main

import (
    "log"

    "github.com/Kisanlink/farmers-module/database"
    "github.com/Kisanlink/farmers-module/routes"
)

func main() {
    // Step 0: Setup the logger

    // Step 1: Initialize the router
    router := routes.Setup()
    

    // Step 2: Defer MongoDB client disconnection
    defer func() {
        if err := database.GetClient().Disconnect(nil); err != nil {
            log.Fatalf("Failed to disconnect MongoDB client: %v", err)
        }
    }()

    // Step 3: Start the server
    err := router.Run("0.0.0.0:80")
	if err != nil {
		log.Fatal("Error starting HTTPS server:", err)
	}
}
