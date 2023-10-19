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

-- name: GetAdminByEmail :one
SELECT * FROM admins
WHERE email = $1 LIMIT 1;

-- name: ListAdmins :many
SELECT * FROM admins
ORDER BY admin_id
LIMIT $1
OFFSET $2;

-- name: DeleteAdmin :exec
DELETE FROM admins WHERE admin_id = $1;