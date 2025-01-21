package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)


type Wallet struct {
    WalletID              string    `json:"_id" bson:"_id"`
    KisanSathiID    int       `json:"kisanSathiId" bson:"kisanSathiId"`
    KisanSathiName  string    `json:"kisanSathiName" bson:"kisanSathiName"`
    AreaManagerID   string    `json:"areaManagerId" bson:"areaManagerId"`
    AreaManagerName string    `json:"areaManagerName" bson:"areaManagerName"`
    SalesValue      float64   `json:"salesValue" bson:"salesValue"`
    WalletAmount    float64   `json:"walletAmount" bson:"walletAmount"`
    StatusRequest   string    `json:"statusRequest" bson:"statusRequest"`
    CreatedBy       string    `json:"createdBy" bson:"createdBy"`
    CreatedAt       time.Time `json:"createdAt" bson:"createdAt"`
    ModifiedBy      string    `json:"modifiedBy" bson:"modifiedBy"`
    ModifiedAt      time.Time `json:"modifiedAt" bson:"modifiedAt"`
    IsDeleted       bool      `json:"isDeleted" bson:"isDeleted"`
    CreatedDate     time.Time `json:"createdDate" bson:"createdDate"`
}
