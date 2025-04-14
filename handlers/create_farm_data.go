package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/config"
)

// CreateFarmData calls the API endpoint to create farm data for the given farmID.
func CreateFarmData(farmID string) {
	// Create an HTTP client with a timeout.
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Prepare the API URL by substituting the farmID.
	url := fmt.Sprintf("%s/api/v1/create-farm-data/%s", config.GetEnv("DIVYA_DRISHTI_ENDPOIN"), farmID)

	// Define the payload as an empty map or a struct if needed.
	payload := map[string]interface{}{}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal farm data request: %v", err)
		return
	}

	// Send the HTTP POST request.
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to call create farm data API: %v", err)
		return
	}
	defer resp.Body.Close()

	// Read and log the response.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read create farm data response: %v", err)
		return
	}

	log.Printf("Create farm data API called successfully. Response: %s", string(body))
}
