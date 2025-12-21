package download

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Danztee/shazam-build/internal/spotify"
)

const DefaultSavePath = "./downloads"

type Service interface {
	DownloadTrack(ctx context.Context, track *spotify.Track, savePath string) error
	DownloadTracks(ctx context.Context, tracks []*spotify.Track, savePath string) (int, error)
}

type svc struct {
	logger *slog.Logger
}

func NewService(logger *slog.Logger) Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &svc{logger: logger}
}

func (s *svc) DownloadTrack(ctx context.Context, track *spotify.Track, savePath string) error {
	if savePath == "" {
		savePath = DefaultSavePath
	}
	if err := os.MkdirAll(savePath, 0755); err != nil {
		return fmt.Errorf("create save directory: %w", err)
	}

	ytID, err := s.getYouTubeID(ctx, track)
	if err != nil {
		return fmt.Errorf("get YouTube ID: %w", err)
	}

	fileName := fmt.Sprintf("%s - %s", sanitize(track.Name), sanitize(s.artist(track)))
	filePath := filepath.Join(savePath, fileName+".mp3")

	if err := s.download(ctx, ytID, filePath); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	s.addTags(ctx, filePath, track)
	s.logger.Info("downloaded", "track", track.Name)
	return nil
}

func (s *svc) DownloadTracks(ctx context.Context, tracks []*spotify.Track, savePath string) (int, error) {
	var count int
	for _, track := range tracks {
		if err := s.DownloadTrack(ctx, track, savePath); err != nil {
			s.logger.Error("download failed", "track", track.Name, "error", err)
			continue
		}
		count++
	}
	return count, nil
}

func (s *svc) getYouTubeID(ctx context.Context, track *spotify.Track) (string, error) {
	query := fmt.Sprintf("%s %s", track.Name, s.artist(track))
	cmd := exec.CommandContext(ctx, "yt-dlp", "--get-id", "--default-search", "ytsearch", query)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	ytID := strings.TrimSpace(string(output))
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]{11}$`).MatchString(ytID) {
		return "", fmt.Errorf("invalid YouTube ID: %s", ytID)
	}
	return ytID, nil
}

func (s *svc) download(ctx context.Context, ytID, outputPath string) error {
	cmd := exec.CommandContext(ctx, "yt-dlp",
		"-x", "--audio-format", "mp3", "--audio-quality", "0",
		"-o", outputPath,
		fmt.Sprintf("https://www.youtube.com/watch?v=%s", ytID))
	return cmd.Run()
}

func (s *svc) addTags(ctx context.Context, filePath string, track *spotify.Track) {
	album := track.Album.Name
	if album == "" {
		album = "Unknown Album"
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", "-i", filePath,
		"-c", "copy",
		"-metadata", fmt.Sprintf("title=%s", track.Name),
		"-metadata", fmt.Sprintf("artist=%s", s.artist(track)),
		"-metadata", fmt.Sprintf("album=%s", album),
		"-y", filePath+"_temp")
	if err := cmd.Run(); err != nil {
		return
	}
	os.Rename(filePath+"_temp", filePath)
}

func (s *svc) artist(track *spotify.Track) string {
	if len(track.Artists) == 0 {
		return "Unknown Artist"
	}
	return track.Artists[0].Name
}

func sanitize(s string) string {
	s = regexp.MustCompile(`[<>:"/\\|?*]`).ReplaceAllString(s, "")
	s = strings.Trim(s, " .")
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}
