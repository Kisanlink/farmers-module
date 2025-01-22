package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Farmer struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FarmerID          int64              `bson:"farmerId" json:"farmerId"`
	Image             string             `bson:"image" json:"image"`
	FirstName         string             `bson:"firstName" json:"firstName"`
	LastName          string             `bson:"lastName" json:"lastName"`
	MobileNumber      int64              `bson:"mobileNumber" json:"mobileNumber"`
	Acres             float64            `bson:"acres" json:"acres"`
	AddedAcres        float64            `bson:"addedAcres" json:"addedAcres"`
	IncentiveAmount   float64            `bson:"incentiveAmount" json:"incentiveAmount"`
	Age               int                `bson:"age" json:"age"`
	Address           string             `bson:"address" json:"address"`
	Longitude         float64            `bson:"longitude" json:"longitude"`
	Latitude          float64            `bson:"lattitude" json:"latitude"`
	KisansathiID      int64              `bson:"kisansathiId" json:"kisansathiId"`
	KisansathiName    string             `bson:"kisansathiName" json:"kisansathiName"`
	Verified          bool               `bson:"verified" json:"verified"`
	City              string             `bson:"city" json:"city"`
	State             string             `bson:"state" json:"state"`
	District          string             `bson:"district" json:"district"`
	NumberOfFarms     int                `bson:"numberofFarms" json:"numberOfFarms"`
	IsFavorite        bool               `bson:"isFavorite" json:"isFavorite"`
	IsActive          bool               `bson:"isActive" json:"isActive"`
	Roles             bson.Raw           `bson:"roles" json:"roles"`
	Pincode           string             `bson:"pincode" json:"pincode"`
	AreaManagerName   string             `bson:"areaManagerName" json:"areaManagerName"`
	AreaManagerID     string             `bson:"areaManagerId" json:"areaManagerId"`
	SalesCompleted    float64            `bson:"salesCompleted" json:"salesCompleted"`
	Shares            int                `bson:"shares" json:"shares"`
	TotalWalletAmount float64            `bson:"totalWalletAmount" json:"totalWalletAmount"`
	CreatedBy         string             `bson:"createdBy" json:"createdBy"`
	IsDeleted         bool               `bson:"isDeleted" json:"isDeleted"`
}
