package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Crop represents a crop or commodity that a farmer might grow
type Crop struct {
	CropID                     primitive.ObjectID `bson:"_id,omitempty"`       // MongoDB ObjectId for the crop (commodity)
	CommodityName           string             `bson:"comodityName"`        // Name of the commodity/crop (e.g., Soya Beans)
	CommodityCategoryID     string             `bson:"comodityCategoryId"`  // ID of the commodity category
	CommodityCategoryName   string             `bson:"comodityCategoryName"`// Name of the commodity category (e.g., Cash Crops)
	Image                   string             `bson:"image"`               // Image URL for the crop
	GSTHSNCode              int                `bson:"gstHsnCode"`          // GST HSN Code for the crop
	HarvestPeriod           string             `bson:"harvestPeriod"`       // Period to harvest the crop
	SeedlingDays            string             `bson:"seedlingDays"`        // Seedling days for the crop
	VegetativeDays          string             `bson:"vegetativeDays"`      // Vegetative days for the crop
	ReproductionDays        string             `bson:"reproductionDays"`    // Reproduction days for the crop
	RipeningDays            string             `bson:"ripeningDays"`        // Ripening days for the crop
	HarvestingDays          string             `bson:"harvestingDays"`      // Harvesting days for the crop
}

// CropDetails struct links to Farms where crops are grown, and directly includes CropID
