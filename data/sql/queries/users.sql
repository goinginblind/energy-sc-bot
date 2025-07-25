-- name: GetUserByPhone :one
SELECT *
  FROM users
 WHERE phone = $1
 LIMIT 1;

-- name: GetUserByEmail :one
SELECT *
  FROM users
 WHERE email = $1
 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (telegram_id, phone, email)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpsertUserByContact :one
INSERT INTO users (telegram_id, phone, email)
VALUES ($1, $2, $3)
ON CONFLICT (phone, email) DO UPDATE
  SET telegram_id = COALESCE(EXCLUDED.telegram_id, users.telegram_id),
      phone       = COALESCE(EXCLUDED.phone, users.phone),
      email       = COALESCE(EXCLUDED.email, users.email)
RETURNING *;