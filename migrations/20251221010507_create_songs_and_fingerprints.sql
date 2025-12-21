-- +goose Up
-- +goose StatementBegin
CREATE TABLE songs (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    artist TEXT NOT NULL,
    album TEXT,
    duration_seconds INT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE fingerprints (
    hash BIGINT NOT NULL,
    song_id INT NOT NULL REFERENCES songs(id),
    time_offset INT NOT NULL
);

-- Lookup by hash quickly
CREATE INDEX idx_fingerprints_hash ON fingerprints(hash);

-- composite index for song/time if needed
CREATE INDEX idx_fingerprints_song_time ON fingerprints(song_id, time_offset);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_fingerprints_song_time;
DROP INDEX IF EXISTS idx_fingerprints_hash;
DROP TABLE IF EXISTS fingerprints;
DROP TABLE IF EXISTS songs;
-- +goose StatementEnd

