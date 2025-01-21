package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type TransactionHistory struct {
    TransactionHistoryID               string    `json:"_id" bson:"_id"`
    WalletAmount     float64   `json:"walletAmount" bson:"walletAmount"`
    IncentivesAmount float64   `json:"incentivesAmount" bson:"incentivesAmount"`
    OrderPayment     float64   `json:"orderPayment" bson:"orderPayment"`
    Date             time.Time `json:"date" bson:"date"`
    KisanSathiID     int       `json:"kisanSathiId" bson:"kisanSathiId"`
    Description      string    `json:"description" bson:"description"`
    TransactionType  string    `json:"transactionType" bson:"transactionType"`
    CreatedBy        string    `json:"createdBy" bson:"createdBy"`
    CreatedAt        time.Time `json:"createdAt" bson:"createdAt"`
    ModifiedBy       string    `json:"modifiedBy" bson:"modifiedBy"`
    ModifiedAt       time.Time `json:"modifiedAt" bson:"modifiedAt"`
    IsDeleted        bool      `json:"isDeleted" bson:"isDeleted"`
}
