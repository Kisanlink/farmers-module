package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Farm represents the farm model in MongoDB
type Farm struct {
	FarmID               primitive.ObjectID   `bson:"_id,omitempty"`    // MongoDB ObjectId for farm
	FarmerID         int                   `bson:"farmerId"`         // References the Farmer's ID
	CropID            string                `bson:"crop"`             // Type of crop (e.g., Soya Beans)

	Address          string                `bson:"address"`          // Address of the farm
	Dimensions       []Dimension           `bson:"dimensions"`       // Coordinates of the farm (latitude, longitude)
	City             string                `bson:"city"`             // City where the farm is located
	State            string                `bson:"state"`            // State of the farm
	District         string                `bson:"district"`         // District of the farm
	Pincode          string                `bson:"pincode"`          // Pincode for the farm
	Images           []string              `bson:"images"`           // Images associated with the farm
	AreaManagerID    primitive.ObjectID    `bson:"areaManagerId"`    // Reference to Area Manager's ID (from AreaManager struct)
	Class            string                `bson:"_class"`            // Class name for MongoDB
}

// Dimension represents the coordinates of the farm (longitude, latitude)
type Dimension struct {
	Longitude float64 `bson:"longitude"` // Longitude of a farm location
	Latitude  float64 `bson:"latitude"`  // Latitude of a farm location
}
