package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	
)

// CommodityPrice model definition for storing crop prices
type CommodityPrice struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	CommodityName     string             `bson:"comodityName" json:"comodityName"`
	Image             string             `bson:"image" json:"image"`
	Price             float64            `bson:"price" json:"price"`
	PaymentTerms      int                `bson:"paymentTerms" json:"paymentTerms"`
	Question1         string             `bson:"question1" json:"question1"`
	Question2         string             `bson:"question2" json:"question2"`
	Question3         string             `bson:"question3" json:"question3"`
	Answer1           int                `bson:"answer1" json:"answer1"`
	Answer2           int                `bson:"answer2" json:"answer2"`
	Answer3           int                `bson:"answer3" json:"answer3"`
	PreviousPrice     float64            `bson:"previousPrice" json:"previousPrice"`
	CreatedBy         string             `bson:"createdBy" json:"createdBy"`
	CreatedAt         Date               `bson:"createdAt"`
	ModifiedBy        string             `bson:"modifiedBy" json:"modifiedBy"`
	ModifiedAt        Date          `bson:"modifiedAt"`
	IsDeleted         bool               `bson:"isDeleted" json:"isDeleted"`
	Class             string             `bson:"_class" json:"class"`
}

