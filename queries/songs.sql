-- name: GetSong :one
SELECT * FROM songs
WHERE id = $1 LIMIT 1;

-- name: ListSongs :many
SELECT * FROM songs
ORDER BY created_at DESC;

-- name: CreateSong :one
INSERT INTO songs (
    title, artists, album, duration_seconds
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateSong :one
UPDATE songs
SET title = $2, artists = $3, album = $4, duration_seconds = $5
WHERE id = $1
RETURNING *;

-- name: DeleteSong :exec
DELETE FROM songs
WHERE id = $1;
