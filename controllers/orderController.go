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
	// Parse farmer ID from URL parameter
	farmerIDstr := c.Param("farmerId")
	farmerID, err := strconv.ParseInt(farmerIDstr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID"})
		return
	}

	// Context and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get orders from repository (expects multiple orders)
	orders, err := oc.OrderRepository.GetOrders(ctx, farmerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err})
		return
	}

	// Check if no orders were found
	if len(orders) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No orders found for this farmer"})
		return
	}

	// Prepare response array to hold order data
	var orderResponses []map[string]interface{}

	// Iterate through all orders and prepare the response
	for _, order := range orders {
		// Prepare the response map for each order
		orderResponse := map[string]interface{}{
	"orderID":           order.OrderID,
	"orderStatus":       order.OrderStatus,
	"orderDate":         order.OrderDate,
	"incentiveAmount":   order.IncentiveAmount,
	"deliveryFee":       order.DeliveryFee,
	"subTotal":          order.SubTotal,
	"totalAmount":       order.TotalAmount,
	"totalTax":          order.TotalTax,
	"totalMrp":          order.TotalMRP,
	"orderType":         order.OrderType,
	"landingPrice":      order.LandingPrice,
	"grossMargin":       order.GrossMargin,
	"paymentDone":       order.PaymentDone,
	"withInterestAmount":order.WithInterestAmount,
	"interest":          order.Interest,
	"isCashBackUsed":    order.IsCashBackUsed,
	"cashBackAmountUsed":order.CashBackAmountUsed,
}


		// Add the order response to the response array
		orderResponses = append(orderResponses, orderResponse)
	}

	// Return response with all orders
	c.JSON(http.StatusOK, orderResponses)
}
