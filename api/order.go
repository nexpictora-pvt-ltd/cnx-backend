package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/nexpictora-pvt-ltd/cnx-backend/db/sqlc"
	"github.com/nexpictora-pvt-ltd/cnx-backend/messaging"
	"github.com/nexpictora-pvt-ltd/cnx-backend/token"
	"github.com/nexpictora-pvt-ltd/cnx-backend/util"
)

type createOrderRequest struct {
	// OrderID     int64   `json:"order_id" binding:"required"`
	CustomerID  int     `json:"customer_id" binding:"required"`
	ServiceIDs  []int32 `json:"service_ids" binding:"required"`
	OrderStatus string  `json:"order_status" binding:"required"`
}

type orderResponse struct {
	OrderID           int64     `json:"order_id"`
	UserID            int       `json:"user_id"`
	Customer          string    `json:"customer_name"`
	OrderStatus       string    `json:"order_status"`
	OrderStarted      time.Time `json:"order_started"`
	OrderDelivered    bool      `json:"order_delivered"`
	OrderDeliveryTime time.Time `json:"order_delivery_time"`
	Services          []struct {
		ServiceID    int64  `json:"service_id"`
		ServiceName  string `json:"service_name"`
		ServicePrice int    `json:"service_price"`
		ServiceImage string `json:"service_image"`
	} `json:"services"`
}

func (server *Server) createOrder(ctx *gin.Context) {
	var req createOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	user, err := server.store.GetUserByEmail(ctx, authPayload.Email)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	if user.UserID != int32(req.CustomerID) {
		err := errors.New("user id does not belong to current customer...")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return

	}
	// Fix: Add a check to ensure that the service IDs array is not empty.
	if len(req.ServiceIDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "service_ids array cannot be empty"})
		return
	}

	var services []struct {
		ServiceID    int64  `json:"service_id"`
		ServiceName  string `json:"service_name"`
		ServicePrice int    `json:"service_price"`
		ServiceImage string `json:"service_image"`
	}
	var createdOrders []db.Order
	orderID := util.NewOrderID()
	// Fetch service details for the response
	for _, serviceID := range req.ServiceIDs {

		createdOrder, err := server.store.CreateOrder(ctx, db.CreateOrderParams{
			OrderID:     orderID,
			UserID:      int32(req.CustomerID),
			ServiceIds:  int32(serviceID),
			OrderStatus: req.OrderStatus,
			// Set other fields as needed
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		createdOrders = append(createdOrders, createdOrder)
		// Fetch service details for the response
		service, err := server.store.GetService(ctx, int32(serviceID))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		services = append(services, struct {
			ServiceID    int64  `json:"service_id"`
			ServiceName  string `json:"service_name"`
			ServicePrice int    `json:"service_price"`
			ServiceImage string `json:"service_image"`
		}{
			ServiceID:    int64(service.ServiceID),
			ServiceName:  service.ServiceName,
			ServicePrice: int(service.ServicePrice),
			ServiceImage: service.ServiceImage,
		})
	}

	// Construct the response object
	response := struct {
		OrderID           int64     `json:"order_id"`
		UserID            int       `json:"user_id"`
		OrderStatus       string    `json:"order_status"`
		OrderStarted      time.Time `json:"order_started"`
		OrderDelivered    bool      `json:"order_delivered"`
		OrderDeliveryTime time.Time `json:"order_delivery_time"`
		Services          []struct {
			ServiceID    int64  `json:"service_id"`
			ServiceName  string `json:"service_name"`
			ServicePrice int    `json:"service_price"`
			ServiceImage string `json:"service_image"`
		} `json:"services"`
	}{
		OrderID:           orderID,
		UserID:            req.CustomerID,
		OrderStatus:       req.OrderStatus,
		OrderStarted:      createdOrders[0].OrderStarted, // Assuming you want the order_started time of the first created order
		OrderDelivered:    createdOrders[0].OrderDelivered,
		OrderDeliveryTime: createdOrders[0].OrderDeliveryTime,
		Services:          services,
	}

	ctx.JSON(http.StatusOK, response)

	message := []byte("New order created: " + strconv.FormatInt(orderID, 10))

	// Publish the message to RabbitMQ
	publisher, err := messaging.NewPublisher("amqp://guest:guest@localhost:5672/")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	defer publisher.Close()

	err = publisher.PublishMessage("new-orders", message, ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
}

type updateOrderStatusRequest struct {
	OrderID     int64  `json:"order_id" binding:"required"`
	OrderStatus string `json:"order_status" binding:"required"`
}

func (server *Server) updateOrderStatus(ctx *gin.Context) {
	var req updateOrderStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	updateOrderStatusParam := db.UpdateOrderStatusParams{
		OrderID:     req.OrderID,
		OrderStatus: req.OrderStatus,
	}

	orderStatus, err := server.store.UpdateOrderStatus(ctx, updateOrderStatusParam)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, orderStatus)
}

type updateOrderDeliveredRequest struct {
	OrderID           int64     `json:"order_id" binding:"required"`
	OrderDelivered    bool      `json:"order_delivered" binding:"required"`
	OrderDeliveryTime time.Time `json:"order_delivery_time" binding:"required"`
}

func (server *Server) updateOrderDelivered(ctx *gin.Context) {
	var req updateOrderDeliveredRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	updateOrderDeliveryParam := db.UpdateOrderDeliveryParams{
		OrderID:           req.OrderID,
		OrderDelivered:    req.OrderDelivered,
		OrderDeliveryTime: req.OrderDeliveryTime,
	}

	orderStatus, err := server.store.UpdateOrderDelivery(ctx, updateOrderDeliveryParam)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, orderStatus)
}

type getOrderRequest struct {
	OrderID int64 `uri:"order_id" binding:"required"`
}

func (server *Server) getOrder(ctx *gin.Context) {
	var req getOrderRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	order, err := server.store.GetOrder(ctx, int64(req.OrderID))
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, order)

}

type listOrdersRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// func (server *Server) listOrders(ctx *gin.Context) {
// 	var req listOrdersRequest
// 	if err := ctx.ShouldBindQuery(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}

// 	arg := db.ListOrdersParams{
// 		Limit:  req.PageSize,
// 		Offset: (req.PageID - 1) * req.PageSize,
// 	}
// 	services, err := server.store.ListOrders(ctx, arg)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}
// 	ctx.JSON(http.StatusOK, services)
// }

func (server *Server) listOrders(ctx *gin.Context) {
	var req listOrdersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Calculate the offset based on the provided page size and page number
	offset := int(req.PageID-1) * int(req.PageSize)

	arg := db.ListOrdersParams{
		Limit:  req.PageSize + int32(offset), // Fetch more records than required to ensure we have enough unique OrderIDs
		Offset: 0,                            // Start fetching from the beginning
	}
	orders, err := server.store.ListOrders(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Create a map to group orders by OrderID
	orderMap := make(map[int64]orderResponse)

	// Iterate through the list of orders and group them by OrderID
	for _, order := range orders {
		user, err := server.store.GetUser(ctx, order.UserID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		service, err := server.store.GetService(ctx, int32(order.ServiceIds))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		orderResp, exists := orderMap[order.OrderID]
		if !exists {
			orderResp = orderResponse{
				OrderID:           order.OrderID,
				UserID:            int(order.UserID),
				Customer:          user.Name,
				OrderStatus:       order.OrderStatus,
				OrderStarted:      order.OrderStarted,
				OrderDelivered:    order.OrderDelivered,
				OrderDeliveryTime: order.OrderDeliveryTime,
				Services: []struct {
					ServiceID    int64  `json:"service_id"`
					ServiceName  string `json:"service_name"`
					ServicePrice int    `json:"service_price"`
					ServiceImage string `json:"service_image"`
				}{
					{
						ServiceID:    int64(service.ServiceID),
						ServiceName:  service.ServiceName,
						ServicePrice: int(service.ServicePrice),
						ServiceImage: service.ServiceImage,
					},
				},
			}
		} else {
			// If the OrderID already exists, append the service to the existing list of services
			orderResp.Services = append(orderResp.Services, struct {
				ServiceID    int64  `json:"service_id"`
				ServiceName  string `json:"service_name"`
				ServicePrice int    `json:"service_price"`
				ServiceImage string `json:"service_image"`
			}{
				ServiceID:    int64(service.ServiceID),
				ServiceName:  service.ServiceName,
				ServicePrice: int(service.ServicePrice),
				ServiceImage: service.ServiceImage,
			})
		}

		orderMap[order.OrderID] = orderResp
	}

	// Convert the map to a slice to maintain consistent ordering
	var response []orderResponse
	for _, orderResp := range orderMap {
		response = append(response, orderResp)
	}

	// Apply pagination to the final response slice
	startIndex := offset
	endIndex := startIndex + int(req.PageSize)
	if startIndex >= len(response) {
		response = []orderResponse{} // If offset is greater than or equal to the number of orders, return an empty slice
	} else if endIndex > len(response) {
		response = response[startIndex:] // If endIndex is beyond the number of orders, return orders from startIndex to the end
	} else {
		response = response[startIndex:endIndex] // Otherwise, return orders from startIndex to endIndex
	}

	ctx.JSON(http.StatusOK, response)
}

func (server *Server) listAllOrders(ctx *gin.Context) {
	orders, err := server.store.ListAllOrders(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Create a map to group orders by OrderID
	orderMap := make(map[int64][]orderResponse)

	// Iterate through the list of orders and group them by OrderID
	for _, order := range orders {
		user, err := server.store.GetUser(ctx, order.UserID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		service, err := server.store.GetService(ctx, int32(order.ServiceIds))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		orderResp := orderResponse{
			OrderID:           order.OrderID,
			UserID:            int(order.UserID),
			Customer:          user.Name,
			OrderStatus:       order.OrderStatus,
			OrderStarted:      order.OrderStarted,
			OrderDelivered:    order.OrderDelivered,
			OrderDeliveryTime: order.OrderDeliveryTime,
			Services: []struct {
				ServiceID    int64  `json:"service_id"`
				ServiceName  string `json:"service_name"`
				ServicePrice int    `json:"service_price"`
				ServiceImage string `json:"service_image"`
			}{
				{
					ServiceID:    int64(service.ServiceID),
					ServiceName:  service.ServiceName,
					ServicePrice: int(service.ServicePrice),
					ServiceImage: service.ServiceImage,
				},
			},
		}

		// Check if the OrderID is already in the map
		if _, exists := orderMap[order.OrderID]; exists {
			// If the OrderID exists, append the service to the existing array of services
			orderMap[order.OrderID][0].Services = append(orderMap[order.OrderID][0].Services, orderResp.Services[0])
		} else {
			// If the OrderID does not exist, create a new entry in the map with the order
			orderMap[order.OrderID] = []orderResponse{orderResp}
		}
	}

	// Convert the map to a slice of orderResponse for the final response
	var response []orderResponse
	for _, v := range orderMap {
		response = append(response, v[0])
	}

	ctx.JSON(http.StatusOK, response)
}

// func (server *Server) createOrder(ctx *gin.Context) {
// 	var req createOrderRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}

// 	// Fix: Add a check to ensure that the service IDs array is not empty.
// 	if len(req.ServiceIDs) == 0 {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "service_ids array cannot be empty"})
// 		return
// 	}
// 	serviceIDs := pq.Int64Array(req.ServiceIDs)
// 	var createdOrders []db.Order
// 	// Iterate through the service IDs and create individual orders
// 	for _, serviceID := range serviceIDs {

// 		createdOrder, err := server.store.CreateOrder(ctx, db.CreateOrderParams{
// 			OrderID:     int64(req.OrderID),
// 			UserID:      int32(req.CustomerID),
// 			ServiceIds:  int32(serviceID),
// 			OrderStatus: req.OrderStatus,
// 			// Set other fields as needed
// 		})
// 		if err != nil {
// 			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 			return
// 		}
// 		createdOrders = append(createdOrders, createdOrder)
// 	}
// 	ctx.JSON(http.StatusOK, createdOrders)
// }

// func (server *Server) createOrder(ctx *gin.Context) {
// 	var req createOrderRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}

// 	// Fix: Add a check to ensure that the service IDs array is not empty.
// 	if len(req.ServiceIDs) == 0 {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "service_ids array cannot be empty"})
// 		return
// 	}

// 	serviceIDs := pq.Int64Array(req.ServiceIDs)
// 	var createdOrdersWithServices []struct {
// 		ID                int64      `json:"id"`
// 		OrderID           int64      `json:"order_id"`
// 		UserID            int        `json:"user_id"`
// 		OrderStatus       string     `json:"order_status"`
// 		OrderStarted      time.Time  `json:"order_started"`
// 		OrderDelivered    bool       `json:"order_delivered"`
// 		OrderDeliveryTime time.Time  `json:"order_delivery_time"`
// 		Services          db.Service `json:"services"`
// 	}

// 	// Iterate through the service IDs and create individual orders
// 	for _, serviceID := range serviceIDs {
// 		createdOrder, err := server.store.CreateOrder(ctx, db.CreateOrderParams{
// 			OrderID:     int64(req.OrderID),
// 			UserID:      int32(req.CustomerID),
// 			ServiceIds:  int32(serviceID),
// 			OrderStatus: req.OrderStatus,
// 			// Set other fields as needed
// 		})
// 		if err != nil {
// 			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 			return
// 		}

// 		// Fetch service details
// 		service, err := server.store.GetService(ctx, int32(serviceID))
// 		if err != nil {
// 			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 			return
// 		}

// 		createdOrdersWithServices = append(createdOrdersWithServices, struct {
// 			ID                int64      `json:"id"`
// 			OrderID           int64      `json:"order_id"`
// 			UserID            int        `json:"user_id"`
// 			OrderStatus       string     `json:"order_status"`
// 			OrderStarted      time.Time  `json:"order_started"`
// 			OrderDelivered    bool       `json:"order_delivered"`
// 			OrderDeliveryTime time.Time  `json:"order_delivery_time"`
// 			Services          db.Service `json:"services"`
// 		}{
// 			ID:                int64(createdOrder.ID),
// 			OrderID:           createdOrder.OrderID,
// 			UserID:            int(createdOrder.UserID),
// 			OrderStatus:       createdOrder.OrderStatus,
// 			OrderStarted:      createdOrder.OrderStarted,
// 			OrderDelivered:    createdOrder.OrderDelivered,
// 			OrderDeliveryTime: createdOrder.OrderDeliveryTime,
// 			Services:          service,
// 		})
// 	}

// 	ctx.JSON(http.StatusOK, createdOrdersWithServices)
// }
