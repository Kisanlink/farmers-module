package main

import (
	"time"
)

// CompletedOrder represents an order that has been completed.
type CompletedOrder struct {
	OrderID      int       `json:"orderId" bson:"orderId"`
	OrderDate    time.Time `json:"orderDate" bson:"orderDate"`
	Amount       float64   `json:"amount" bson:"totalAmount"`
	CompletedDate time.Time `json:"completedDate" bson:"completedDate"`
	OrderStatus  string    `json:"orderStatus" bson:"orderStatus"`
	FarmerID   string    `json:"farmerName" bson:"farmerName"`
	OrderType    string    `json:"orderType" bson:"orderType"`
}
