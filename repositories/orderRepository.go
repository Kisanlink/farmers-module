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

// GetOrders retrieves all orders placed by a farmer using farmerID
func (repo *OrderRepository) GetOrders(ctx context.Context, farmerID int64) ([]models.Order, error) {
	// Debug: Log the farmer ID being queried
	log.Printf("DEBUG: Starting query for orders placed by farmer with ID: %d", farmerID)

	// Perform the query using Find to get multiple orders for the given farmer ID
	cursor, err := repo.Collection.Find(ctx, bson.M{"farmerId": farmerID})
	if err != nil {
		log.Printf("ERROR: Failed to retrieve orders for farmerId %d: %v", farmerID, err)
		return nil, err
	}
	defer cursor.Close(ctx) // Ensure the cursor is closed once done iterating

	// Declare a slice to hold the orders
	var orders []models.Order

	// Iterate through the cursor and decode each order into the orders slice
	for cursor.Next(ctx) {
		var order models.Order
		if err := cursor.Decode(&order); err != nil {
			log.Printf("ERROR: Failed to decode order: %v", err)
			return nil, err
		}
		orders = append(orders, order)
	}

	// Check for errors while iterating through the cursor
	if err := cursor.Err(); err != nil {
		log.Printf("ERROR: Cursor error: %v", err)
		return nil, err
	}

	// Debug: Successfully retrieved multiple orders
	log.Printf("DEBUG: Successfully retrieved %d orders for farmerId %d", len(orders), farmerID)

	// Return the slice of orders
	return orders, nil
}


// GetCreditOrdersByFarmerID retrieves credit orders placed by a farmer using farmerID
func (repo *OrderRepository) GetCreditOrdersByFarmerID(ctx context.Context, farmerID int64) ([]models.Order, error) {
	// Perform the query using Find to get orders with "Credit" status and the given farmer ID
	cursor, err := repo.Collection.Find(ctx, bson.M{"farmerId": farmerID, "orderStatus": "Credit"})
	if err != nil {
		log.Printf("ERROR: Failed to retrieve credit orders for farmerId %d: %v", farmerID, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	// Declare a slice to hold the orders
	var orders []models.Order

	// Iterate through the cursor and decode each order into the orders slice
	for cursor.Next(ctx) {
		var order models.Order
		if err := cursor.Decode(&order); err != nil {
			log.Printf("ERROR: Failed to decode order: %v", err)
			return nil, err
		}
		orders = append(orders, order)
	}

	// Check for errors while iterating through the cursor
	if err := cursor.Err(); err != nil {
		log.Printf("ERROR: Cursor error: %v", err)
		return nil, err
	}

	// Return the slice of orders
	return orders, nil
}
