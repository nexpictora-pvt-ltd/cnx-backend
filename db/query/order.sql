-- name: CreateOrder :one
INSERT INTO orders (
  order_id,
  user_id,
  service_ids,
  order_status 
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetOrder :many
SELECT * FROM orders
WHERE order_id = $1;

-- name: ListOrders :many
SELECT * FROM orders
ORDER BY order_id DESC
LIMIT $1
OFFSET $2;

-- name: ListAllOrders :many
SELECT * FROM orders
ORDER BY id DESC;

-- name: ListAllOrdersByUserId :many
SELECT * FROM orders
ORDER BY user_id DESC;

-- name: UpdateOrder :one
UPDATE orders 
SET order_status = $2
WHERE order_id = $1
RETURNING *;

-- name: CancelOrder :one
UPDATE orders 
SET order_status = $2
WHERE order_id = $1
RETURNING *;

-- name: UpdateOrderStatus :one
UPDATE orders 
SET order_status = $2
WHERE order_id = $1
RETURNING *;

-- name: UpdateOrderDelivery :one
UPDATE orders 
SET order_delivered = $2,
order_delivery_time = $3
WHERE order_id = $1
RETURNING *;

-- name: DeleteOrder :exec
DELETE FROM orders WHERE order_id = $1;