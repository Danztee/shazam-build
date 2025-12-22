package download

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Danztee/shazam-build/internal/spotify"
	"github.com/kkdai/youtube/v2"
)

const DefaultSavePath = "./downloads"

type Service interface {
	DownloadTrack(ctx context.Context, track *spotify.Track, savePath string) (string, error)
	DownloadTracks(ctx context.Context, tracks []*spotify.Track, savePath string) (int, error)
}

type svc struct {
	logger *slog.Logger
	client *youtube.Client
}

func NewService(logger *slog.Logger) Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &svc{logger: logger, client: &youtube.Client{}}
}

func (s *svc) DownloadTrack(ctx context.Context, track *spotify.Track, savePath string) (string, error) {
	if savePath == "" {
		savePath = DefaultSavePath
	}
	os.MkdirAll(savePath, 0755)

	query := fmt.Sprintf("%s %s", track.Name, s.artist(track))
	resp, _ := http.Get(fmt.Sprintf("https://www.youtube.com/results?search_query=%s", url.QueryEscape(query)))
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	re := regexp.MustCompile(`"videoId":"([a-zA-Z0-9_-]{11})"`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return "", fmt.Errorf("no video found")
	}

	video, _ := s.client.GetVideoContext(ctx, matches[1])
	var format *youtube.Format
	for _, f := range video.Formats {
		if f.AudioQuality != "" {
			format = &f
			break
		}
	}
	if format == nil {
		return "", fmt.Errorf("no audio format")
	}

	fileName := fmt.Sprintf("%s - %s.mp3", sanitize(track.Name), sanitize(s.artist(track)))
	filePath := filepath.Join(savePath, fileName)
	tempPath := filePath + ".tmp"

	stream, _, _ := s.client.GetStreamContext(ctx, video, format)
	file, _ := os.Create(tempPath)
	file.ReadFrom(stream)
	stream.Close()
	file.Close()

	if exec.Command("ffmpeg", "-i", tempPath, "-acodec", "libmp3lame", "-q:a", "2", "-y", filePath).Run() == nil {
		os.Remove(tempPath)
	} else {
		os.Rename(tempPath, filePath)
	}

	exec.Command("ffmpeg", "-i", filePath, "-c", "copy",
		"-metadata", fmt.Sprintf("title=%s", track.Name),
		"-metadata", fmt.Sprintf("artist=%s", s.artist(track)),
		"-metadata", fmt.Sprintf("album=%s", track.Album.Name),
		"-y", filePath+"_t").Run()
	os.Rename(filePath+"_t", filePath)

	wavPath := strings.TrimSuffix(filePath, ".mp3") + ".wav"
	if err := exec.Command("ffmpeg",
		"-i", filePath,
		"-ar", "44100", // Sample rate: 44.1kHz - CD quality
		"-ac", "1", // Mono - single channel
		"-sample_fmt", "s16", // 16-bit signed integer
		"-y", // Overwrite existing file
		wavPath,
	).Run(); err != nil {
		s.logger.Warn("failed to convert to WAV", "error", err, "track", track.Name)
	} else {
		s.logger.Info("converted to WAV", "track", track.Name, "path", wavPath)
	}

	s.logger.Info("downloaded", "track", track.Name)
	return wavPath, nil
}

func (s *svc) DownloadTracks(ctx context.Context, tracks []*spotify.Track, savePath string) (int, error) {
	var count int
	for _, t := range tracks {
		if _, err := s.DownloadTrack(ctx, t, savePath); err == nil {
			count++
		}
	}
	return count, nil
}

func (s *svc) artist(track *spotify.Track) string {
	if len(track.Artists) == 0 {
		return "Unknown"
	}
	return track.Artists[0].Name
}

func sanitize(s string) string {
	s = regexp.MustCompile(`[<>:"/\\|?*]`).ReplaceAllString(s, "")
	return strings.Trim(s, " .")
}
