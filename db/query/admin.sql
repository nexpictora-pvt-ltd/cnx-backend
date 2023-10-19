-- name: CreateAdmin :one
INSERT INTO admins (
  name,
  email,
  phone,
  address,
  hashed_password 
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: AddAdmin :one
INSERT INTO admins (
  name,
  email,
  phone,
  address,
  hashed_password 
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: DeleteAdmin :exec
DELETE FROM admins WHERE admin_id = $1;