package main

import (
	"time"
)

// OrderItem represents an item in an order
type OrderItem struct {
	OrderID      int       `json:"orderId" bson:"orderId"`
	ProductID    int       `json:"productId" bson:"productId"`
	ProductName  string    `json:"productName" bson:"nameOfProduct"`
	ProductCost  float64   `json:"productCost" bson:"productCost"`
	Quantity     int       `json:"quantity" bson:"quantity"`
	MRP          float64   `json:"mrp" bson:"mrp"`
	ItemTax      float64   `json:"itemTax" bson:"itemTax"`
	Discount     float64   `json:"discount" bson:"discount"`
	DeliveryFee  float64   `json:"deliveryFee" bson:"deliveryFee"`
	DeliveryDate time.Time `json:"deliveryDate" bson:"deliveryDate"`
	OrderStatus  string    `json:"orderStatus" bson:"orderStatus"`
	Photo        string    `json:"photo" bson:"photo"`
}
