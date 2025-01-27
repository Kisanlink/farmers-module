package models

import (
	
	"go.mongodb.org/mongo-driver/bson/primitive"
	
)

type Farm struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FarmID              int                `bson:"farmId" json:"farmId"`
	FarmerID            int                `bson:"farmerId" json:"farmerId"`
	FarmerName          string             `bson:"farmerName" json:"farmerName"`
	VariantID           string             `bson:"variantId" json:"variantId"`
	Status              string             `bson:"status" json:"status"`
	ExpectedYield       float64            `bson:"expectedYield" json:"expectedYield"`
	Price               float64            `bson:"price" json:"price"`
	CropCategory        string             `bson:"cropCategory" json:"cropCategory"`
	Crop                string             `bson:"crop" json:"crop"`
	CropID              string             `bson:"cropId" json:"cropId"`
	CropImage           string             `bson:"cropImage" json:"cropImage"`
	CropVariety         string             `bson:"cropVariety" json:"cropVariety"`
	Acres               float64            `bson:"acres" json:"acres"`
	Tons                float64            `bson:"tons" json:"tons"`
	HarvestDate         Date                `bson:"harvestDate" json:"harvestDate"`
	FarmerMobileNumber  int64              `bson:"farmerMobileNumber" json:"farmerMobileNumber"`
	Address             string             `bson:"address" json:"address"`
	Dimensions          []Dimension        `bson:"dimensions" json:"dimensions"`
	Locality            string             `bson:"locality" json:"locality"`
	Landmark            string             `bson:"landmark" json:"landmark"`
	Pincode             string             `bson:"pincode" json:"pincode"`
	City                string             `bson:"city" json:"city"`
	State               string             `bson:"state" json:"state"`
	District            string             `bson:"district" json:"district"`
	Images              []string           `bson:"images" json:"images"`
	KisansathiID        int                `bson:"kisansathiId" json:"kisansathiId"`
	KisansathiName      string             `bson:"kisansathiName" json:"kisansathiName"`
	Verified            bool               `bson:"verified" json:"verified"`
	FarmerVerified      bool               `bson:"farmerVerified" json:"farmerVerified"`
	IsActive            bool               `bson:"isActive" json:"isActive"`
	Question1           string             `bson:"question1" json:"question1"`
	Question2           string             `bson:"question2" json:"question2"`
	Question3           string             `bson:"question3" json:"question3"`
	Answer1             string             `bson:"answer1" json:"answer1"`
	Answer2             string             `bson:"answer2" json:"answer2"`
	Answer3             string             `bson:"answer3" json:"answer3"`
	CommodityCategoryID string             `bson:"comodityCategoryId" json:"commodityCategoryId"`
	AreaManagerID       string             `bson:"areaManagerId" json:"areaManagerId"`
	AreaManagerName     string             `bson:"areaManagerName" json:"areaManagerName"`
	CreatedBy           string             `bson:"createdBy" json:"createdBy"`
  CreatedAt           Date               `bson:"createdAt"`
	ModifiedBy          string             `bson:"modifiedBy" json:"modifiedBy"`
  ModifiedAt          Date               `bson:"modifiedAt"`
	IsDeleted           bool               `bson:"isDeleted" json:"isDeleted"`
	Class               string             `bson:"_class" json:"_class"`
}

type Dimension struct {
	Longitude float64 `bson:"longitude" json:"longitude"`
	Latitude  float64 `bson:"latitude" json:"latitude"`
}
type Date struct {
	OrigYear          int     `bson:"orig_year" json:"orig_year"`
	OrigMonth         int     `bson:"orig_month" json:"orig_month"`
	OrigDay           int     `bson:"orig_day" json:"orig_day"`
	OrigHour          int     `bson:"orig_hour" json:"orig_hour"`
	OrigMinute        int     `bson:"orig_minute" json:"orig_minute"`
	OrigSecond        int     `bson:"orig_second" json:"orig_second"`
	OrigTimezone      int     `bson:"orig_timezone" json:"orig_timezone"`
	Year              int     `bson:"year" json:"year"`
	Month             int     `bson:"month" json:"month"`
	Day               int     `bson:"day" json:"day"`
	Timezone          int     `bson:"timezone" json:"timezone"`
	Hour              int     `bson:"hour" json:"hour"`
	Minute            int     `bson:"minute" json:"minute"`
	Second            int     `bson:"second" json:"second"`
	FractionalSecond string  `bson:"fractionalSecond" json:"fractionalSecond"`
	Class             string  `bson:"_class" json:"class"`
}

