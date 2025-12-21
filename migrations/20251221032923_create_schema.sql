-- +goose Up
-- +goose StatementBegin

-- Create artists table
CREATE TABLE artists (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create songs table (without artist column)
CREATE TABLE songs (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    album TEXT,
    duration_seconds INT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create song_artists junction table for many-to-many relationship
CREATE TABLE song_artists (
    song_id INT NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    artist_id INT NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    PRIMARY KEY (song_id, artist_id)
);

-- Create fingerprints table
CREATE TABLE fingerprints (
    hash BIGINT NOT NULL,
    song_id INT NOT NULL REFERENCES songs(id),
    time_offset INT NOT NULL
);

-- Create indexes for performance
CREATE INDEX idx_artists_name ON artists(name);
CREATE INDEX idx_song_artists_song_id ON song_artists(song_id);
CREATE INDEX idx_song_artists_artist_id ON song_artists(artist_id);
CREATE INDEX idx_fingerprints_hash ON fingerprints(hash);
CREATE INDEX idx_fingerprints_song_time ON fingerprints(song_id, time_offset);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indexes
DROP INDEX IF EXISTS idx_fingerprints_song_time;
DROP INDEX IF EXISTS idx_fingerprints_hash;
DROP INDEX IF EXISTS idx_song_artists_artist_id;
DROP INDEX IF EXISTS idx_song_artists_song_id;
DROP INDEX IF EXISTS idx_artists_name;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS fingerprints;
DROP TABLE IF EXISTS song_artists;
DROP TABLE IF EXISTS songs;
DROP TABLE IF EXISTS artists;

-- +goose StatementEnd

