package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)


type CropDetails struct {
	FarmID         int                 `bson:"farmId"`           // Unique farm ID
	FarmerID       int                 `bson:"farmerId"`         // Reference to the farmer's ID who owns the farm
	CropID         primitive.ObjectID  `bson:"cropId"`           // Reference to the Crop (Commodity) ID
	
	Acres          float64             `bson:"acres"`            // Acres of land for the farm
	Tons           float64             `bson:"tons"`             // Amount harvested in tons
	ExpectedYield  float64             `bson:"expectedYield"`    // Expected yield for the farm
	Price          float64             `bson:"price"`            // Price of the crop
	HarvestDate    time.Time           `bson:"harvestDate"`      // Date of harvest
	Verified       bool                `bson:"verified"`         // Verification status of the crop
	IsActive       bool                `bson:"isActive"`         // Active status of the farm
	IsDeleted      bool                `bson:"isDeleted"`        // Soft delete flag for the farm
	CreatedBy      string              `bson:"createdBy"`        // Who created the farm record
	CreatedAt      time.Time           `bson:"createdAt"`        // When the farm record was created
	ModifiedBy     string              `bson:"modifiedBy"`       // Who modified the farm record
	ModifiedAt     time.Time           `bson:"modifiedAt"`       // When the farm record was last modified
	AreaManagerID  primitive.ObjectID  `bson:"areaManagerId"`    // Reference to Area Manager's ID (from AreaManager struct)
}



