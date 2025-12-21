package songs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	repo "github.com/Danztee/shazam-build/internal/database/queries"
	"github.com/Danztee/shazam-build/internal/download"
	"github.com/Danztee/shazam-build/internal/spotify"

	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	AddSong(ctx context.Context, song createSongPayload) (repo.Song, error)
}

type svc struct {
	repo          repo.Querier
	spotifyClient spotify.Service
	downloadSvc   download.Service
}

func NewService(repo repo.Querier, spotifyClient spotify.Service, downloadSvc download.Service) Service {
	return &svc{
		repo:          repo,
		spotifyClient: spotifyClient,
		downloadSvc:   downloadSvc,
	}
}

func (s *svc) AddSong(ctx context.Context, payload createSongPayload) (repo.Song, error) {

	urlInfo, err := s.spotifyClient.ExtractURLInfo(payload.SpotifyUrl)
	if err != nil {
		return repo.Song{}, fmt.Errorf("invalid spotify URL: %w", err)
	}

	token, err := s.spotifyClient.GetAccessToken()
	if err != nil {
		return repo.Song{}, fmt.Errorf("failed to get access token: %w", err)
	}

	if urlInfo.Type == "album" {
		return s.addAlbumTracks(ctx, token, urlInfo.ID, payload.Download, payload.SavePath)
	}

	track, err := s.spotifyClient.GetTrack(ctx, token, urlInfo.ID)
	if err != nil {
		return repo.Song{}, fmt.Errorf("failed to get track info: %w", err)
	}

	song, err := s.createSongFromTrack(ctx, track)
	if err != nil {
		return repo.Song{}, fmt.Errorf("failed to create song: %w", err)
	}

	// Download track if requested
	if payload.Download && s.downloadSvc != nil {
		if err := s.downloadSvc.DownloadTrack(ctx, track, payload.SavePath); err != nil {
			slog.Warn("failed to download track", "error", err, "track", track.Name)
			// Don't fail the whole operation if download fails
		}
	}

	return song, nil
}

func (s *svc) addAlbumTracks(ctx context.Context, token, albumID string, download bool, savePath string) (repo.Song, error) {
	tracks, _, err := s.spotifyClient.GetAlbumTracks(ctx, token, albumID)
	if err != nil {
		return repo.Song{}, fmt.Errorf("failed to get album tracks: %w", err)
	}

	if len(tracks) == 0 {
		return repo.Song{}, errors.New("album has no tracks")
	}

	var lastSong repo.Song
	for _, track := range tracks {
		song, err := s.createSongFromTrack(ctx, track)
		if err != nil {
			return repo.Song{}, fmt.Errorf("failed to create song %s: %w", track.Name, err)
		}
		lastSong = song
	}

	// Download tracks if requested
	if download && s.downloadSvc != nil {
		if _, err := s.downloadSvc.DownloadTracks(ctx, tracks, savePath); err != nil {
			slog.Warn("failed to download some album tracks", "error", err)
			// Don't fail the whole operation if download fails
		}
	}

	return lastSong, nil
}

func (s *svc) createSongFromTrack(ctx context.Context, track *spotify.Track) (repo.Song, error) {
	artistNames := make([]string, 0, len(track.Artists))
	for _, artist := range track.Artists {
		artistNames = append(artistNames, artist.Name)
	}

	artistsJSON, err := json.Marshal(artistNames)
	if err != nil {
		return repo.Song{}, fmt.Errorf("failed to marshal artists: %w", err)
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
		Artists:         artistsJSON,
		Album:           albumName,
		DurationSeconds: durationSeconds,
	}

	song, err := s.repo.CreateSong(ctx, createParams)
	if err != nil {
		return repo.Song{}, fmt.Errorf("failed to create song: %w", err)
	}

	return song, nil
}
