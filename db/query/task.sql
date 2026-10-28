-- name: CreateTask :one
INSERT INTO tasks (user_id, title)
VALUES ($1, $2)
RETURNING id, user_id, title, completed, created_at;

-- name: GetTasksByUserID :many
SELECT id, user_id, title, completed, created_at
FROM tasks
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: ToggleTask :one
UPDATE tasks
SET completed = NOT completed
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1 AND user_id = $2;