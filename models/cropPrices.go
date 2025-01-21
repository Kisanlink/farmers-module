package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// CommodityPrice represents the price details of a commodity, including the name, price, and other relevant information
type CommodityPrice struct {
	CommodityPriceID              primitive.ObjectID `bson:"_id,omitempty"`         // MongoDB ObjectId for the commodity price entry
	CommodityName   string             `bson:"comodityName"`          // Name of the commodity (e.g., Soya Beans)
	
	Price           float64            `bson:"price"`                 // Current price of the commodity
	PaymentTerms    int                `bson:"paymentTerms"`          // Payment terms (e.g., days)
	
	PreviousPrice   float64            `bson:"previousPrice"`         // Price of the commodity from the previous period
	CreatedBy       string             `bson:"createdBy"`             // User who created this price entry
	CreatedAt       time.Time          `bson:"createdAt"`             // Timestamp of when the price entry was created
	ModifiedBy      string             `bson:"modifiedBy"`            // User who last modified the price entry
	ModifiedAt      time.Time          `bson:"modifiedAt"`            // Timestamp of when the price entry was last modified
	IsDeleted       bool               `bson:"isDeleted"`             // Soft delete flag for the price entry
	Class           string             `bson:"_class"`                // Class name for MongoDB (if applicable)
}

