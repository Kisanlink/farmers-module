package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Withdraw struct {
    WithdrawID            string    `json:"_id" bson:"_id"`
    Amount        float64   `json:"amount" bson:"amount"`
    KisanSathiID  int       `json:"kisanSathiId" bson:"kisanSathiId"`
    Status        string    `json:"status" bson:"status"`
    Date          time.Time `json:"date" bson:"date"`
    ModifiedBy    string    `json:"modifiedBy" bson:"modifiedBy"`
    ModifiedAt    time.Time `json:"modifiedAt" bson:"modifiedAt"`
    IsDeleted     bool      `json:"isDeleted" bson:"isDeleted"`
}
