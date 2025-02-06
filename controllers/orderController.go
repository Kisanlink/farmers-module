package controllers

import (
	"context"
	"net/http"
	"time"
	"strconv"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/gin-gonic/gin"
)

type OrderController struct {
	OrderRepository *repositories.OrderRepository
}

func NewOrderController(orderRepo *repositories.OrderRepository) *OrderController {
	return &OrderController{
		OrderRepository: orderRepo,
	}
}

// GetOrdersByFarmerID retrieves all orders placed by a farmer using farmerID
func (oc *OrderController) GetOrdersByFarmerID(c *gin.Context) {
	// Parse query parameters
	farmerIDstr := c.Query("farmerId")
	if farmerIDstr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "farmerId query parameter is required"})
		return
	}

	farmerID, err := strconv.ParseInt(farmerIDstr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID"})
		return
	}

	// Optional filters
	orderStatus := c.Query("status")   // Example: ?status=Delivered
	startDate := c.Query("startDate")  // Example: ?startDate=2024-01-01
	endDate := c.Query("endDate")      // Example: ?endDate=2024-02-01

	// Context and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch orders with optional filtering
	orders, err := oc.OrderRepository.GetOrders(ctx, farmerID, orderStatus, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// No orders found
	if len(orders) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No orders found"})
		return
	}

	c.JSON(http.StatusOK, orders)
}


// GetCreditOrdersByFarmerID retrieves credit orders placed by a farmer using query parameters
func (oc *OrderController) GetCreditOrdersByFarmerID(c *gin.Context) {
    // Parse query parameters
    farmerIDStr := c.Query("farmerId")
    if farmerIDStr == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "farmerId query parameter is required"})
        return
    }

    farmerID, err := strconv.ParseInt(farmerIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID"})
        return
    }

    status := c.Query("status") // Optional filter

    // Context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Fetch credit orders from repository
    creditOrders, err := oc.OrderRepository.GetCreditOrders(ctx, farmerID, status)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    if len(creditOrders) == 0 {
        c.JSON(http.StatusNotFound, gin.H{"message": "No credit orders found"})
        return
    }

    c.JSON(http.StatusOK, creditOrders)
}
