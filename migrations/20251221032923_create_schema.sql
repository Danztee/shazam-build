-- +goose Up
-- +goose StatementBegin

-- Create songs table with artists as JSONB array
CREATE TABLE songs (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    artists JSONB NOT NULL DEFAULT '[]'::jsonb,
    album TEXT,
    duration_seconds INT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create fingerprints table
CREATE TABLE fingerprints (
    hash BIGINT NOT NULL,
    song_id INT NOT NULL REFERENCES songs(id),
    time_offset INT NOT NULL
);

-- Create indexes for performance
CREATE INDEX idx_songs_artists ON songs USING GIN (artists);
CREATE INDEX idx_fingerprints_hash ON fingerprints(hash);
CREATE INDEX idx_fingerprints_song_time ON fingerprints(song_id, time_offset);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indexes
DROP INDEX IF EXISTS idx_fingerprints_song_time;
DROP INDEX IF EXISTS idx_fingerprints_hash;
DROP INDEX IF EXISTS idx_songs_artists;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS fingerprints;
DROP TABLE IF EXISTS songs;

-- +goose StatementEnd

