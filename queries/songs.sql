-- name: GetSong :one
SELECT * FROM songs
WHERE id = $1 LIMIT 1;

-- name: GetSongByTitleAndArtists :one
SELECT * FROM songs
WHERE title = $1 AND artists = $2::jsonb
LIMIT 1;

-- name: ListSongs :many
SELECT * FROM songs
ORDER BY created_at DESC;

-- name: CreateSong :one
INSERT INTO songs (
    title, artists, album, duration_ms, image_url
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateSong :one
UPDATE songs
SET title = $2, artists = $3, album = $4, duration_ms = $5, image_url = $6
WHERE id = $1
RETURNING *;

-- name: DeleteSong :exec
DELETE FROM songs
WHERE id = $1;
