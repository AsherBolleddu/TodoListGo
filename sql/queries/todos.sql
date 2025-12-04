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
