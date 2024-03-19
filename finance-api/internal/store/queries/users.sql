-- name: CurrentUser :one
SELECT id, name, email, roles FROM users LIMIT 1;