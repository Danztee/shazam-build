package songs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/Danztee/shazam-build/internal/audio"
	repo "github.com/Danztee/shazam-build/internal/database/queries"
	"github.com/Danztee/shazam-build/internal/download"
	"github.com/Danztee/shazam-build/internal/spotify"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	AddSong(ctx context.Context, song createSongPayload) (repo.Song, error)
}

type svc struct {
	repo          repo.Querier
	db            *pgx.Conn
	spotifyClient spotify.Service
	downloadSvc   download.Service
	audioSvc      audio.Service
}

func NewService(repo repo.Querier, db *pgx.Conn, spotifyClient spotify.Service, downloadSvc download.Service, audioSvc audio.Service) Service {
	return &svc{
		repo:          repo,
		db:            db,
		spotifyClient: spotifyClient,
		downloadSvc:   downloadSvc,
		audioSvc:      audioSvc,
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
		return s.addAlbumTracks(ctx, token, urlInfo.ID, payload.Download)
	}

	track, err := s.spotifyClient.GetTrack(ctx, token, urlInfo.ID)
	if err != nil {
		return repo.Song{}, fmt.Errorf("failed to get track info: %w", err)
	}

	song, isNew, err := s.findOrCreateSong(ctx, track)
	if err != nil {
		return repo.Song{}, fmt.Errorf("failed to find or create song: %w", err)
	}

	if !isNew {
		slog.Info("song already exists, skipping download and fingerprint processing", "song_id", song.ID, "title", song.Title)
		return song, nil
	}

	if payload.Download && s.downloadSvc != nil {
		go func() {
			wavPath, err := s.downloadSvc.DownloadTrack(context.Background(), track, "")
			if err != nil {
				slog.Warn("download failed", "error", err, "track", track.Name)
				return
			}

			if wavPath != "" && s.audioSvc != nil {
				if err := s.processAndSaveFingerprints(context.Background(), song.ID, wavPath); err != nil {
					slog.Warn("failed to process fingerprints", "error", err, "track", track.Name)
				}
			}
		}()
	}

	return song, nil
}

func (s *svc) addAlbumTracks(ctx context.Context, token, albumID string, download bool) (repo.Song, error) {
	tracks, _, err := s.spotifyClient.GetAlbumTracks(ctx, token, albumID)
	if err != nil {
		return repo.Song{}, fmt.Errorf("failed to get album tracks: %w", err)
	}

	if len(tracks) == 0 {
		return repo.Song{}, errors.New("album has no tracks")
	}

	// Store created songs to pass to goroutine
	createdSongs := make([]repo.Song, len(tracks))
	var lastSong repo.Song

	for i, track := range tracks {
		song, _, err := s.findOrCreateSong(ctx, track)
		if err != nil {
			return repo.Song{}, fmt.Errorf("failed to find or create song %s: %w", track.Name, err)
		}
		createdSongs[i] = song
		lastSong = song
	}

	if download && s.downloadSvc != nil {
		go func() {
			for i, track := range tracks {
				song := createdSongs[i]

				// Proceed to download and process since we are in the "add album" context
				// The songs were either just created or retrieved, and the user requested download.
				// We assume if the user requests download for an album, they want all tracks processed
				// regardless of whether they existed before this request.
				wavPath, err := s.downloadSvc.DownloadTrack(context.Background(), track, "")
				if err != nil {
					slog.Warn("download failed", "error", err, "track", track.Name)
					continue
				}

				if wavPath != "" && s.audioSvc != nil {
					if err := s.processAndSaveFingerprints(context.Background(), song.ID, wavPath); err != nil {
						slog.Warn("failed to process fingerprints", "error", err, "track", track.Name)
					}
				}
			}
		}()
	}

	return lastSong, nil
}

func (s *svc) findOrCreateSong(ctx context.Context, track *spotify.Track) (repo.Song, bool, error) {
	artistNames := make([]string, 0, len(track.Artists))
	for _, artist := range track.Artists {
		artistNames = append(artistNames, artist.Name)
	}

	artistsJSON, err := json.Marshal(artistNames)
	if err != nil {
		return repo.Song{}, false, fmt.Errorf("failed to marshal artists: %w", err)
	}

	existingSong, err := s.repo.GetSongByTitleAndArtists(ctx, repo.GetSongByTitleAndArtistsParams{
		Title:   track.Name,
		Column2: artistsJSON,
	})
	if err == nil {
		slog.Info("song already exists in database", "song_id", existingSong.ID, "title", existingSong.Title)
		return existingSong, false, nil
	}

	albumName := pgtype.Text{Valid: false}
	if track.Album.Name != "" {
		albumName = pgtype.Text{String: track.Album.Name, Valid: true}
	}

	durationMs := pgtype.Int4{Valid: false}
	if track.DurationMs > 0 {
		durationMs = pgtype.Int4{Int32: int32(track.DurationMs), Valid: true}
	}

	createParams := repo.CreateSongParams{
		Title:      track.Name,
		Artists:    artistsJSON,
		Album:      albumName,
		DurationMs: durationMs,
	}

	song, err := s.repo.CreateSong(ctx, createParams)
	if err != nil {
		return repo.Song{}, false, fmt.Errorf("failed to create song: %w", err)
	}

	slog.Info("created new song", "song_id", song.ID, "title", song.Title)
	return song, true, nil
}

func (s *svc) processAndSaveFingerprints(ctx context.Context, songID int32, wavPath string) error {
	defer func() {
		os.Remove(wavPath)
		mp3Path := strings.TrimSuffix(wavPath, ".wav") + ".mp3"
		os.Remove(mp3Path)
		slog.Info("cleaned up audio files", "wav_path", wavPath)
	}()

	if _, err := os.Stat(wavPath); os.IsNotExist(err) {
		return fmt.Errorf("WAV file not found: %s", wavPath)
	}

	fingerprints, err := s.audioSvc.ProcessAudio(ctx, wavPath)
	if err != nil {
		return fmt.Errorf("failed to process audio: %w", err)
	}

	if len(fingerprints) == 0 {
		slog.Warn("no fingerprints generated", "song_id", songID)
		return nil
	}

	type fingerprintKey struct {
		Hash         int64
		TimeOffsetMs int32
		SongID       int32
	}
	uniqueFingerprints := make(map[fingerprintKey]bool)
	deduplicated := make([]audio.Fingerprint, 0, len(fingerprints))

	for _, fp := range fingerprints {
		key := fingerprintKey{
			Hash:         fp.Hash,
			TimeOffsetMs: int32(fp.TimeOffsetMs),
			SongID:       songID,
		}
		if !uniqueFingerprints[key] {
			uniqueFingerprints[key] = true
			deduplicated = append(deduplicated, fp)
		}
	}

	if len(deduplicated) == 0 {
		slog.Warn("no unique fingerprints after deduplication", "song_id", songID)
		return nil
	}

	slog.Info("saving fingerprints", "song_id", songID, "count", len(deduplicated), "original_count", len(fingerprints))

	rowsCopied, err := s.db.CopyFrom(
		ctx,
		pgx.Identifier{"fingerprints"},
		[]string{"hash", "song_id", "time_offset_ms"},
		pgx.CopyFromSlice(len(deduplicated), func(i int) ([]any, error) {
			return []any{
				deduplicated[i].Hash,
				songID,
				int32(deduplicated[i].TimeOffsetMs),
			}, nil
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to batch insert fingerprints: %w", err)
	}

	slog.Info("successfully saved all fingerprints", "song_id", songID, "count", rowsCopied, "total_fingerprints", len(deduplicated))
	return nil
}
