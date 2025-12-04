-- name: CreateTodo :one
INSERT INTO
    todos (
        id,
        created_at,
        updated_at,
        title,
        description,
        user_id
    )
VALUES (
        gen_random_uuid (),
        NOW(),
        NOW(),
        $1,
        $2,
        $3
    ) RETURNING *;

-- name: UpdateTodo :one
UPDATE todos
SET
    title = $3,
    description = $4,
    updated_at = NOW()
WHERE
    id = $1
    AND user_id = $2 RETURNING *;

-- name: GetTodoByID :one
SELECT * FROM todos WHERE id = $1;

-- name: DeleteTodo :exec
DELETE FROM todos WHERE id = $1 AND user_id = $2;

-- name: GetTodosByUserID :many
SELECT *
FROM todos
WHERE
    user_id = $1
ORDER BY updated_at DESC
LIMIT $2
OFFSET
    $3;

-- name: GetTodosCountByUserID :one
SELECT COUNT(*) FROM todos where user_id = $1;
