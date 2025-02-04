package models

import (
	

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// XMLGregorianCalendar represents the date-time format in your MongoDB document.
type XMLGregorianCalendar struct {
	Year              int    `bson:"year"`
	Month             int    `bson:"month"`
	Day               int    `bson:"day"`
	Timezone          int    `bson:"timezone"`
	Hour              int    `bson:"hour"`
	Minute            int    `bson:"minute"`
	Second            int    `bson:"second"`
	FractionalSecond  string `bson:"fractionalSecond"`
	Class             string `bson:"_class"`
}

// SoilTestReport represents the structure of your MongoDB document.
type SoilTestReport struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReportDate    XMLGregorianCalendar `bson:"reportDate" json:"reportDate"`
	FarmID        int                 `bson:"farmId" json:"farmId"`
	FarmerID      int                 `bson:"farmerId" json:"farmerId"`
	KisanSathiID  int                 `bson:"kisanSathiId" json:"kisanSathiId"`
	AreaManagerID string              `bson:"areaManagerId" json:"areaManagerId"`
	Reports       []string            `bson:"reports" json:"reports"`
	CreatedBy     string              `bson:"createdBy" json:"createdBy"`
	CreatedAt     XMLGregorianCalendar `bson:"createdAt" json:"createdAt"`
	ModifiedBy    string              `bson:"modifiedBy" json:"modifiedBy"`
	ModifiedAt    XMLGregorianCalendar `bson:"modifiedAt" json:"modifiedAt"`
	IsDeleted     bool                 `bson:"isDeleted" json:"isDeleted"`
	Class         string               `bson:"_class" json:"class"`
}
