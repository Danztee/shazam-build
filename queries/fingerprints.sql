-- name: GetFingerprintsByHash :many
SELECT * FROM fingerprints
WHERE hash = $1;

-- name: GetFingerprintsBySongID :many
SELECT * FROM fingerprints
WHERE song_id = $1
ORDER BY time_offset_ms ASC;

-- name: CreateFingerprint :one
INSERT INTO fingerprints (
    hash, song_id, time_offset_ms
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetFingerprintsByHashes :many
SELECT * FROM fingerprints
WHERE hash = ANY($1::bigint[]);

-- name: DeleteFingerprintsBySongID :exec
DELETE FROM fingerprints
WHERE song_id = $1;

