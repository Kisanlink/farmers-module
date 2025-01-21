package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// SoilTestReport represents a soil test report for a specific farm
type SoilTestReport struct {
	SoilTestReportID               primitive.ObjectID `bson:"_id,omitempty"`       // MongoDB ObjectId for the report
	ReportDate       time.Time          `bson:"reportDate"`          // Date of the soil test report
	FarmID           int                `bson:"farmId"`              // ID of the associated farm
	FarmerID         int                `bson:"farmerId"`            // ID of the associated farmer
	KisanSathiID     int                `bson:"kisanSathiId"`        // ID of the associated Kisan Sathi
	AreaManagerID    primitive.ObjectID `bson:"areaManagerId"`       // ID of the associated Area Manager
	Reports          []string           `bson:"reports"`             // List of report URLs (images, data, etc.)
	CreatedBy        string             `bson:"createdBy"`           // Who created the report
	CreatedAt        time.Time          `bson:"createdAt"`           // When the report was created
	ModifiedBy       string             `bson:"modifiedBy"`          // Who modified the report
	ModifiedAt       time.Time          `bson:"modifiedAt"`          // When the report was last modified
	IsDeleted        bool               `bson:"isDeleted"`           // Soft delete flag for the report
	Class            string             `bson:"_class"`              // Class name for MongoDB (if applicable)
}






