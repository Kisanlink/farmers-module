package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)


type FarmerCashbacks struct {
    FarmerCashbacksID            string    `json:"_id" bson:"_id"`
    OrderNumber   int       `json:"orderNumber" bson:"orderNumber"`
    ProductID     int       `json:"productId" bson:"productId"`
    FarmerID      int       `json:"farmerId" bson:"farmerId"`
    AreaManagerID string    `json:"areaManagerId" bson:"areaManagerId"`
    KisansathiID  int       `json:"kisansathiId" bson:"kisansathiId"`
    WalletAmount  float64   `json:"walletAmount" bson:"walletAmount"`
    CreatedBy     string    `json:"createdBy" bson:"createdBy"`
    CreatedAt     time.Time `json:"createdAt" bson:"createdAt"`
    ModifiedBy    string    `json:"modifiedBy" bson:"modifiedBy"`
    ModifiedAt    time.Time `json:"modifiedAt" bson:"modifiedAt"`
    IsDeleted     bool      `json:"isDeleted" bson:"isDeleted"`
    Type          string    `json:"type" bson:"type"`
}
