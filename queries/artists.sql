-- name: CreateOrGetArtist :one
INSERT INTO artists (name)
VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = artists.name
RETURNING *;

-- name: GetArtistByName :one
SELECT * FROM artists
WHERE name = $1 LIMIT 1;

-- name: LinkSongToArtist :exec
INSERT INTO song_artists (song_id, artist_id)
VALUES ($1, $2)
ON CONFLICT (song_id, artist_id) DO NOTHING;

-- name: GetArtistsBySongID :many
SELECT a.* FROM artists a
INNER JOIN song_artists sa ON a.id = sa.artist_id
WHERE sa.song_id = $1
ORDER BY a.name;

-- name: GetSongWithArtists :one
SELECT 
    s.*,
    COALESCE(
        json_agg(
            json_build_object('id', a.id, 'name', a.name, 'created_at', a.created_at)
        ) FILTER (WHERE a.id IS NOT NULL),
        '[]'::json
    ) as artists
FROM songs s
LEFT JOIN song_artists sa ON s.id = sa.song_id
LEFT JOIN artists a ON sa.artist_id = a.id
WHERE s.id = $1
GROUP BY s.id;

-- name: GetSongsWithArtists :many
SELECT 
    s.*,
    COALESCE(
        json_agg(
            json_build_object('id', a.id, 'name', a.name, 'created_at', a.created_at)
        ) FILTER (WHERE a.id IS NOT NULL),
        '[]'::json
    ) as artists
FROM songs s
LEFT JOIN song_artists sa ON s.id = sa.song_id
LEFT JOIN artists a ON sa.artist_id = a.id
GROUP BY s.id
ORDER BY s.created_at DESC;

