package main

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/services"
)

// This is a demonstration of the Farm Service functionality
// In a real application, this would be integrated with the full system
func main() {
	fmt.Println("Farm Management Service Demo")
	fmt.Println("============================")

	// Create a farm service instance (normally this would be dependency injected)
	service := &services.FarmServiceImpl{}

	// Demonstrate geometry validation
	fmt.Println("\n1. Geometry Validation Examples:")

	testGeometries := []struct {
		name string
		wkt  string
	}{
		{"Valid Square", "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))"},
		{"Valid Farm in Bangalore", "POLYGON((77.5946 12.9716, 77.6046 12.9716, 77.6046 12.9816, 77.5946 12.9816, 77.5946 12.9716))"},
		{"Invalid Point", "POINT(0 0)"},
		{"Empty Geometry", ""},
	}

	for _, test := range testGeometries {
		err := service.ValidateGeometry(context.Background(), test.wkt)
		status := "✓ Valid"
		if err != nil {
			status = fmt.Sprintf("✗ Invalid: %s", err.Error())
		}
		fmt.Printf("  %s: %s\n", test.name, status)
	}

	// Demonstrate request validation
	fmt.Println("\n2. Request Validation Examples:")

	testRequests := []struct {
		name string
		req  *requests.CreateFarmRequest
	}{
		{
			"Valid Request",
			&requests.CreateFarmRequest{
				AAAFarmerUserID: "farmer123",
				AAAOrgID:        "org123",
				Geometry: requests.GeometryData{
					WKT: "POLYGON((77.5946 12.9716, 77.6046 12.9716, 77.6046 12.9816, 77.5946 12.9816, 77.5946 12.9716))",
				},
				Metadata: map[string]string{
					"name": "Demo Farm",
					"crop": "Rice",
				},
			},
		},
		{
			"Missing Farmer ID",
			&requests.CreateFarmRequest{
				AAAOrgID: "org123",
				Geometry: requests.GeometryData{
					WKT: "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
				},
			},
		},
	}

	for _, test := range testRequests {
		// This would normally be a private method, but for demo purposes we'll show the concept
		var err error
		if test.req.AAAFarmerUserID == "" {
			err = fmt.Errorf("farmer user ID is required")
		} else if test.req.AAAOrgID == "" {
			err = fmt.Errorf("organization ID is required")
		} else if test.req.Geometry.WKT == "" {
			err = fmt.Errorf("geometry is required")
		}

		status := "✓ Valid"
		if err != nil {
			status = fmt.Sprintf("✗ Invalid: %s", err.Error())
		}
		fmt.Printf("  %s: %s\n", test.name, status)
	}

	// Demonstrate area filtering
	fmt.Println("\n3. Area Filtering Examples:")

	// This would normally come from the database
	sampleFarms := []*struct {
		Name   string
		AreaHa float64
	}{
		{"Small Farm", 0.5},
		{"Medium Farm", 2.0},
		{"Large Farm", 10.0},
		{"Very Large Farm", 50.0},
	}

	fmt.Println("  All farms:")
	for _, farm := range sampleFarms {
		fmt.Printf("    - %s: %.1f hectares\n", farm.Name, farm.AreaHa)
	}

	// Filter examples
	minArea := 1.0
	maxArea := 15.0
	fmt.Printf("\n  Farms between %.1f and %.1f hectares:\n", minArea, maxArea)
	for _, farm := range sampleFarms {
		if farm.AreaHa >= minArea && farm.AreaHa <= maxArea {
			fmt.Printf("    - %s: %.1f hectares\n", farm.Name, farm.AreaHa)
		}
	}

	fmt.Println("\n4. Key Features Implemented:")
	fmt.Println("  ✓ WKT geometry validation with PostGIS integration")
	fmt.Println("  ✓ SRID enforcement (WGS84 - EPSG:4326)")
	fmt.Println("  ✓ Polygon integrity checks (self-intersection detection)")
	fmt.Println("  ✓ Area calculation using PostGIS ST_Area function")
	fmt.Println("  ✓ Spatial filtering with bounding box queries")
	fmt.Println("  ✓ AAA service integration for authorization")
	fmt.Println("  ✓ Comprehensive error handling and validation")
	fmt.Println("  ✓ Request/response models with proper typing")
	fmt.Println("  ✓ Unit and integration tests")

	fmt.Println("\n5. API Endpoints Available:")
	fmt.Println("  POST   /farms           - Create a new farm")
	fmt.Println("  GET    /farms/{id}      - Get farm by ID")
	fmt.Println("  PUT    /farms/{id}      - Update farm")
	fmt.Println("  DELETE /farms/{id}      - Delete farm")
	fmt.Println("  GET    /farms           - List farms with filtering")

	fmt.Println("\nDemo completed successfully!")
}
