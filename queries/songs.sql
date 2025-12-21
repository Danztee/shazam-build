-- name: GetSong :one
SELECT * FROM songs
WHERE id = $1 LIMIT 1;

-- name: ListSongs :many
SELECT * FROM songs
ORDER BY created_at DESC;

-- name: CreateSong :one
INSERT INTO songs (
    title, album, duration_seconds
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: UpdateSong :one
UPDATE songs
SET title = $2, album = $3, duration_seconds = $4
WHERE id = $1
RETURNING *;

-- name: DeleteSong :exec
DELETE FROM songs
WHERE id = $1;
