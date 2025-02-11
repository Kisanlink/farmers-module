package repositories

import (
	"context"
	"log"

	"github.com/Kisanlink/farmers-module/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepository struct {
	Collection *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) *OrderRepository {
	return &OrderRepository{
		Collection: db.Collection("Orders"),
	}
}

func (repo *OrderRepository) GetOrders(ctx context.Context, farmerID int64, status, startDate, endDate string) ([]models.Order, error) {
	// Create the base query
	filter := bson.M{"farmerId": farmerID}

	// Apply optional filters
	if status != "" {
		filter["orderStatus"] = status
	}

	// Filter by date range if provided
	if startDate != "" && endDate != "" {
		filter["orderDate"] = bson.M{
			"$gte": startDate,
			"$lte": endDate,
		}
	}

	// Perform the query
	cursor, err := repo.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	for cursor.Next(ctx) {
		var order models.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// GetOrdersWithFilters fetches orders by farmer ID with paymentmode and orderstatus filters
func (repo *OrderRepository) GetOrdersWithFilters(ctx context.Context, farmerID int64, paymentMode, orderStatus string) ([]models.Order, error) {
	// Create the base query
	filter := bson.M{"farmerId": farmerID}

	// Apply filters
	if paymentMode != "" {
		filter["paymentMode"] = paymentMode
	}
	if orderStatus != "" {
		filter["orderStatus"] = orderStatus
	}

	// Perform the query
	cursor, err := repo.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	for cursor.Next(ctx) {
		var order models.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// GetCreditOrders fetches credit orders by farmer ID and optional status filter
func (repo *OrderRepository) GetCreditOrders(ctx context.Context, farmerID int64, status string) ([]models.Order, error) {
	filter := bson.M{"farmerId": farmerID, "orderStatus": "Credit"}

	if status != "" {
		filter["status"] = status
	}

	cursor, err := repo.Collection.Find(ctx, filter)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve credit orders: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	for cursor.Next(ctx) {
		var order models.Order
		if err := cursor.Decode(&order); err != nil {
			log.Printf("ERROR: Failed to decode order: %v", err)
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}
