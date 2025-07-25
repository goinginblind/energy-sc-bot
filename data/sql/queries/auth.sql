-- name: CreateOTP :one
INSERT INTO otp_requests (user_id, otp_code, expires_at)
VALUES ($1, $2, now() + interval '5 minutes')
RETURNING *;

-- name: GetValidOTP :one
SELECT * 
FROM otp_requests
WHERE user_id = $1
   AND otp_code = $2
   AND expires_at > now();

-- name: DeleteOTP :exec
DELETE FROM otp_requests WHERE user_id = $1;

-- name: CreateSession :one
INSERT INTO sessions (user_id, token, expires_at)
VALUES ($1, $2, now() + interval '3 hours')
RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions WHERE token = $1 AND expires_at > now();

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token = $1;