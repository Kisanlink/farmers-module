package main

import (
	"time"
)

// CurrentOrder represents an order that is still being processed or pending.
type CurrentOrder struct {
	OrderID        int       `json:"orderId" bson:"orderId"`
	OrderDate      time.Time `json:"orderDate" bson:"orderDate"`
	Amount         float64   `json:"amount" bson:"totalAmount"`
	DueDate        time.Time `json:"dueDate" bson:"dueDate"`
	Interest       float64   `json:"interest" bson:"interest"`
	OrderStatus    string    `json:"orderStatus" bson:"orderStatus"`
	PaymentDone    bool      `json:"paymentDone" bson:"paymentDone"`
	FarmerID     string    `json:"farmerName" bson:"farmerName"`
	
	OrderType      string    `json:"orderType" bson:"orderType"`
	ShippingAddress string   `json:"shippingAddress" bson:"shippingAddress"`
}
