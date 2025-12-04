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
