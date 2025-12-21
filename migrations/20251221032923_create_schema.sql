
-- +goose Up
-- +goose StatementBegin

CREATE TABLE songs (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    artists JSONB NOT NULL DEFAULT '[]'::jsonb,
    album TEXT,
    duration_ms INT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE fingerprints (
    hash BIGINT NOT NULL,
    song_id INT NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    time_offset_ms INT NOT NULL,
    PRIMARY KEY (hash, time_offset_ms, song_id)
);

-- Indexes
CREATE INDEX idx_songs_artists ON songs USING GIN (artists);

-- You can keep this if you want faster song-specific queries/cleanup
CREATE INDEX idx_fingerprints_song_time ON fingerprints(song_id, time_offset_ms);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_fingerprints_song_time;
DROP INDEX IF EXISTS idx_songs_artists;

DROP TABLE IF EXISTS fingerprints;
DROP TABLE IF EXISTS songs;

-- +goose StatementEnd