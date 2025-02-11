package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

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
	// Parse farmer ID from URL path
	farmerIDstr := c.Param("id")
	if farmerIDstr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Farmer ID is required"})
		return
	}

	farmerID, err := strconv.ParseInt(farmerIDstr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID"})
		return
	}

	// Optional filters
	orderStatus := c.Query("status")  // Example: ?status=Delivered
	startDate := c.Query("startDate") // Example: ?startDate=2024-01-01
	endDate := c.Query("endDate")     // Example: ?endDate=2024-02-01

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

// GetOrdersByFarmerIDWithFilters retrieves orders placed by a farmer using farmerID with paymentmode and orderstatus filters
func (oc *OrderController) GetOrdersByFarmerIDWithFilters(c *gin.Context) {
	// Parse farmer ID from URL path
	farmerIDstr := c.Param("id")
	if farmerIDstr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Farmer ID is required"})
		return
	}

	farmerID, err := strconv.ParseInt(farmerIDstr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID"})
		return
	}

	// Query parameters
	paymentMode := c.Query("paymentmode") // Example: ?paymentmode=online
	orderStatus := c.Query("orderstatus") // Example: ?orderstatus=delivered

	// Context and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch orders with filtering
	orders, err := oc.OrderRepository.GetOrdersWithFilters(ctx, farmerID, paymentMode, orderStatus)
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
	// Parse farmer ID from URL path
	farmerIDstr := c.Param("id")
	if farmerIDstr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Farmer ID is required"})
		return
	}

	farmerID, err := strconv.ParseInt(farmerIDstr, 10, 64)
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
