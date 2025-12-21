package songs

import (
	"context"
	"errors"
	"fmt"

	repo "github.com/Danztee/shazam-build/internal/database/queries"
	"github.com/Danztee/shazam-build/internal/spotify"

	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	AddSong(ctx context.Context, song createSongPayload) (repo.Song, error)
}

type svc struct {
	repo          repo.Querier
	spotifyClient spotify.Service
}

func NewService(repo repo.Querier, spotifyClient spotify.Service) Service {
	return &svc{
		repo:          repo,
		spotifyClient: spotifyClient,
	}
}

func (s *svc) AddSong(ctx context.Context, payload createSongPayload) (repo.Song, error) {
	// Extract URL type and ID from Spotify URL
	urlInfo, err := s.spotifyClient.ExtractURLInfo(payload.SpotifyUrl)
	if err != nil {
		return repo.Song{}, fmt.Errorf("invalid spotify URL: %w", err)
	}

	// Get Spotify access token
	token, err := s.spotifyClient.GetAccessToken()
	if err != nil {
		return repo.Song{}, fmt.Errorf("failed to get access token: %w", err)
	}

	// Handle album URLs
	if urlInfo.Type == "album" {
		return s.addAlbumTracks(ctx, token, urlInfo.ID)
	}

	// Handle track URLs
	track, err := s.spotifyClient.GetTrack(ctx, token, urlInfo.ID)
	if err != nil {
		return repo.Song{}, fmt.Errorf("failed to get track info: %w", err)
	}

	// Create song in database
	song, err := s.createSongFromTrack(ctx, track)
	if err != nil {
		return repo.Song{}, fmt.Errorf("failed to create song: %w", err)
	}

	return song, nil
}

// addAlbumTracks fetches all tracks from an album and adds them to the database
func (s *svc) addAlbumTracks(ctx context.Context, token, albumID string) (repo.Song, error) {
	// Get album tracks
	tracks, _, err := s.spotifyClient.GetAlbumTracks(ctx, token, albumID)
	if err != nil {
		return repo.Song{}, fmt.Errorf("failed to get album tracks: %w", err)
	}

	if len(tracks) == 0 {
		return repo.Song{}, errors.New("album has no tracks")
	}

	// Create all songs in database
	var lastSong repo.Song
	for _, track := range tracks {
		song, err := s.createSongFromTrack(ctx, track)
		if err != nil {
			return repo.Song{}, fmt.Errorf("failed to create song %s: %w", track.Name, err)
		}
		lastSong = song
	}

	return lastSong, nil
}

// createSongFromTrack creates a song in the database from a Spotify track
func (s *svc) createSongFromTrack(ctx context.Context, track *spotify.Track) (repo.Song, error) {
	artistName := ""
	if len(track.Artists) > 0 {
		artistName = track.Artists[0].Name
	}

	albumName := pgtype.Text{Valid: false}
	if track.Album.Name != "" {
		albumName = pgtype.Text{String: track.Album.Name, Valid: true}
	}

	durationSeconds := pgtype.Int4{Valid: false}
	if track.DurationMs > 0 {
		durationSeconds = pgtype.Int4{Int32: int32(track.DurationMs / 1000), Valid: true}
	}

	createParams := repo.CreateSongParams{
		Title:           track.Name,
		Artist:          artistName,
		Album:           albumName,
		DurationSeconds: durationSeconds,
	}

	song, err := s.repo.CreateSong(ctx, createParams)
	if err != nil {
		return repo.Song{}, fmt.Errorf("failed to create song: %w", err)
	}

	return song, nil
}
