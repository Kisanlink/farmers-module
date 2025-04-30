package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/config"
	"github.com/Kisanlink/farmers-module/utils"
)

// CreateFarmData calls the API endpoint to create farm data for the given farm_id.
func CreateFarmData(farm_id string) {
	// Create an HTTP client with a timeout.
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Prepare the API URL by substituting the farm_id.
	url := fmt.Sprintf("%s/api/v1/create-farm-data/%s", config.GetEnv("DIVYA_DRISHTI_ENDPOINT"), farm_id)

	// Define the payload as an empty map or a struct if needed.
	payload := map[string]interface{}{}
	json_data, err := json.Marshal(payload)
	if err != nil {
		utils.Log.Errorf("Failed to marshal farm data request: %v", err)
		return
	}

	// Send the HTTP POST request.
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		utils.Log.Errorf("Failed to call create farm data API: %v", err)
		return
	}
	defer resp.Body.Close()

	// Read and log the response.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Log.Errorf("Failed to read create farm data response: %v", err)
		return
	}

	utils.Log.Infof("Create farm data API called successfully. Response: %s", string(body))
}
