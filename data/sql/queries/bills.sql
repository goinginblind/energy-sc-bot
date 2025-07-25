-- name: GetBillsByUserID :many
SELECT id, user_id, pdf_url, amount, status, issued_at, due_date
  FROM bills
 WHERE user_id = $1
 ORDER BY issued_at DESC;

-- name: GetBillByID :one
SELECT id, user_id, pdf_url, amount, status, issued_at, due_date
  FROM bills
 WHERE id = $1 AND user_id = $2;