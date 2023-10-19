// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.21.0
// source: order.sql

package db

import (
	"context"
	"time"
)

const createOrder = `-- name: CreateOrder :one
INSERT INTO orders (
  order_id,
  user_id,
  service_ids,
  order_status 
) VALUES (
  $1, $2, $3, $4
) RETURNING id, order_id, user_id, service_ids, order_status, order_started, order_delivered, order_delivery_time
`

type CreateOrderParams struct {
	OrderID     int64  `json:"order_id"`
	UserID      int32  `json:"user_id"`
	ServiceIds  int32  `json:"service_ids"`
	OrderStatus string `json:"order_status"`
}

func (q *Queries) CreateOrder(ctx context.Context, arg CreateOrderParams) (Order, error) {
	row := q.db.QueryRowContext(ctx, createOrder,
		arg.OrderID,
		arg.UserID,
		arg.ServiceIds,
		arg.OrderStatus,
	)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.UserID,
		&i.ServiceIds,
		&i.OrderStatus,
		&i.OrderStarted,
		&i.OrderDelivered,
		&i.OrderDeliveryTime,
	)
	return i, err
}

const deleteOrder = `-- name: DeleteOrder :exec
DELETE FROM orders WHERE order_id = $1
`

func (q *Queries) DeleteOrder(ctx context.Context, orderID int64) error {
	_, err := q.db.ExecContext(ctx, deleteOrder, orderID)
	return err
}

const getOrder = `-- name: GetOrder :many
SELECT id, order_id, user_id, service_ids, order_status, order_started, order_delivered, order_delivery_time FROM orders
WHERE order_id = $1
`

func (q *Queries) GetOrder(ctx context.Context, orderID int64) ([]Order, error) {
	rows, err := q.db.QueryContext(ctx, getOrder, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Order{}
	for rows.Next() {
		var i Order
		if err := rows.Scan(
			&i.ID,
			&i.OrderID,
			&i.UserID,
			&i.ServiceIds,
			&i.OrderStatus,
			&i.OrderStarted,
			&i.OrderDelivered,
			&i.OrderDeliveryTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listAllOrders = `-- name: ListAllOrders :many
SELECT id, order_id, user_id, service_ids, order_status, order_started, order_delivered, order_delivery_time FROM orders
ORDER BY id DESC
`

func (q *Queries) ListAllOrders(ctx context.Context) ([]Order, error) {
	rows, err := q.db.QueryContext(ctx, listAllOrders)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Order{}
	for rows.Next() {
		var i Order
		if err := rows.Scan(
			&i.ID,
			&i.OrderID,
			&i.UserID,
			&i.ServiceIds,
			&i.OrderStatus,
			&i.OrderStarted,
			&i.OrderDelivered,
			&i.OrderDeliveryTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listOrders = `-- name: ListOrders :many
SELECT id, order_id, user_id, service_ids, order_status, order_started, order_delivered, order_delivery_time FROM orders
ORDER BY order_id DESC
LIMIT $1
OFFSET $2
`

type ListOrdersParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListOrders(ctx context.Context, arg ListOrdersParams) ([]Order, error) {
	rows, err := q.db.QueryContext(ctx, listOrders, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Order{}
	for rows.Next() {
		var i Order
		if err := rows.Scan(
			&i.ID,
			&i.OrderID,
			&i.UserID,
			&i.ServiceIds,
			&i.OrderStatus,
			&i.OrderStarted,
			&i.OrderDelivered,
			&i.OrderDeliveryTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateOrder = `-- name: UpdateOrder :one
UPDATE orders 
SET order_status = $2
WHERE order_id = $1
RETURNING id, order_id, user_id, service_ids, order_status, order_started, order_delivered, order_delivery_time
`

type UpdateOrderParams struct {
	OrderID     int64  `json:"order_id"`
	OrderStatus string `json:"order_status"`
}

func (q *Queries) UpdateOrder(ctx context.Context, arg UpdateOrderParams) (Order, error) {
	row := q.db.QueryRowContext(ctx, updateOrder, arg.OrderID, arg.OrderStatus)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.UserID,
		&i.ServiceIds,
		&i.OrderStatus,
		&i.OrderStarted,
		&i.OrderDelivered,
		&i.OrderDeliveryTime,
	)
	return i, err
}

const updateOrderDelivery = `-- name: UpdateOrderDelivery :one
UPDATE orders 
SET order_delivered = $2,
order_delivery_time = $3
WHERE order_id = $1
RETURNING id, order_id, user_id, service_ids, order_status, order_started, order_delivered, order_delivery_time
`

type UpdateOrderDeliveryParams struct {
	OrderID           int64     `json:"order_id"`
	OrderDelivered    bool      `json:"order_delivered"`
	OrderDeliveryTime time.Time `json:"order_delivery_time"`
}

func (q *Queries) UpdateOrderDelivery(ctx context.Context, arg UpdateOrderDeliveryParams) (Order, error) {
	row := q.db.QueryRowContext(ctx, updateOrderDelivery, arg.OrderID, arg.OrderDelivered, arg.OrderDeliveryTime)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.UserID,
		&i.ServiceIds,
		&i.OrderStatus,
		&i.OrderStarted,
		&i.OrderDelivered,
		&i.OrderDeliveryTime,
	)
	return i, err
}

const updateOrderStatus = `-- name: UpdateOrderStatus :one
UPDATE orders 
SET order_status = $2
WHERE order_id = $1
RETURNING id, order_id, user_id, service_ids, order_status, order_started, order_delivered, order_delivery_time
`

type UpdateOrderStatusParams struct {
	OrderID     int64  `json:"order_id"`
	OrderStatus string `json:"order_status"`
}

func (q *Queries) UpdateOrderStatus(ctx context.Context, arg UpdateOrderStatusParams) (Order, error) {
	row := q.db.QueryRowContext(ctx, updateOrderStatus, arg.OrderID, arg.OrderStatus)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.UserID,
		&i.ServiceIds,
		&i.OrderStatus,
		&i.OrderStarted,
		&i.OrderDelivered,
		&i.OrderDeliveryTime,
	)
	return i, err
}
