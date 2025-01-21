package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Farmer represents the farmer model in MongoDB
type Farmer struct {
	
	FarmerID         int                   `bson:"farmerId"`
	Image            string                `bson:"image"`
	FirstName        string                `bson:"firstName"`
	LastName         string                `bson:"lastName"`
	MobileNumber     int64                 `bson:"mobileNumber"`
	Acres            float64               `bson:"acres"`
	Age              int                   `bson:"age"`
	Address          string                `bson:"address"`
	Longtitude       float64               `bson:"longitude"`
	Latitude         float64               `bson:"lattitude"`
	KisansathiID     int                   `bson:"kisansathiId"`
	KisansathiName   string                `bson:"kisansathiName"`
	Verified         bool                  `bson:"verified"`
	City             string                `bson:"city"`
	State            string                `bson:"state"`
	District         string                `bson:"district"`
	NumberOfFarms    int                   `bson:"numberofFarms"`
	IsFavorite       bool                  `bson:"isFavorite"`
	IsActive         bool                  `bson:"isActive"`
	Roles            []string              `bson:"roles"`
	Pincode          string                `bson:"pincode"`
	AreaManagerID    primitive.ObjectID    `bson:"areaManagerId"`    // Reference to Area Manager's ID
	SalesCompleted   float64               `bson:"salesCompleted"`
	Shares           int                   `bson:"shares"`
	TotalWalletAmount float64              `bson:"totalWalletAmount"`
	
}


