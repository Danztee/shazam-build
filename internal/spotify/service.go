package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Danztee/shazam-build/internal/env"
)

var (
	clientID     = env.GetString("SPOTIFY_CLIENT_ID", "")
	clientSecret = env.GetString("SPOTIFY_CLIENT_SECRET", "")
)

type Service interface {
	ExtractURLInfo(spotifyURL string) (URLInfo, error)
	GetAccessToken() (string, error)
	GetTrack(ctx context.Context, token, trackID string) (*Track, error)
	GetAlbumTracks(ctx context.Context, token, albumID string) ([]*Track, *Album, error)
}

type svc struct{}

func NewService() Service {
	return &svc{}
}

func (s *svc) ExtractURLInfo(spotifyURL string) (URLInfo, error) {
	if parsedURL, err := url.Parse(spotifyURL); err == nil && parsedURL.Host == "open.spotify.com" {
		parts := strings.Split(strings.TrimPrefix(parsedURL.Path, "/"), "/")
		if len(parts) == 2 && (parts[0] == "track" || parts[0] == "album") {
			return URLInfo{Type: parts[0], ID: parts[1]}, nil
		}
	}

	if m := regexp.MustCompile(`track/([a-zA-Z0-9]+)`).FindStringSubmatch(spotifyURL); len(m) > 1 {
		return URLInfo{Type: "track", ID: m[1]}, nil
	}
	if m := regexp.MustCompile(`album/([a-zA-Z0-9]+)`).FindStringSubmatch(spotifyURL); len(m) > 1 {
		return URLInfo{Type: "album", ID: m[1]}, nil
	}
	return URLInfo{}, errors.New("could not extract track or album ID from URL")
}

func (s *svc) GetAccessToken() (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	req, _ := http.NewRequest("POST", "https://accounts.spotify.com/api/token", bytes.NewBufferString(data.Encode()))
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("spotify API error: %s, body: %s", resp.Status, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.AccessToken == "" {
		return "", errors.New("access token is empty")
	}
	return result.AccessToken, nil
}

func (s *svc) GetTrack(ctx context.Context, token, trackID string) (*Track, error) {
	var track Track
	if err := s.get(ctx, token, fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", trackID), &track); err != nil {
		return nil, err
	}
	return &track, nil
}

func (s *svc) GetAlbumTracks(ctx context.Context, token, albumID string) ([]*Track, *Album, error) {
	var album struct {
		Name string `json:"name"`
	}
	if err := s.get(ctx, token, fmt.Sprintf("https://api.spotify.com/v1/albums/%s", albumID), &album); err != nil {
		return nil, nil, err
	}

	var allTrackIDs []string
	offset := 0
	for {
		var result struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
			Total int `json:"total"`
		}
		if err := s.get(ctx, token, fmt.Sprintf("https://api.spotify.com/v1/albums/%s/tracks?limit=50&offset=%d", albumID, offset), &result); err != nil {
			return nil, nil, err
		}
		for _, item := range result.Items {
			allTrackIDs = append(allTrackIDs, item.ID)
		}
		if offset+len(result.Items) >= result.Total {
			break
		}
		offset += 50
	}

	if len(allTrackIDs) == 0 {
		return []*Track{}, &Album{Name: album.Name}, nil
	}

	tracks, err := s.getTracks(ctx, token, allTrackIDs)
	if err != nil {
		return nil, nil, err
	}

	for _, track := range tracks {
		track.Album.Name = album.Name
	}
	return tracks, &Album{Name: album.Name}, nil
}

func (s *svc) getTracks(ctx context.Context, token string, trackIDs []string) ([]*Track, error) {
	const maxBatch = 50
	var allTracks []*Track
	for i := 0; i < len(trackIDs); i += maxBatch {
		end := i + maxBatch
		if end > len(trackIDs) {
			end = len(trackIDs)
		}
		var result struct {
			Tracks []Track `json:"tracks"`
		}
		url := fmt.Sprintf("https://api.spotify.com/v1/tracks?ids=%s", strings.Join(trackIDs[i:end], ","))
		if err := s.get(ctx, token, url, &result); err != nil {
			return nil, err
		}
		for i := range result.Tracks {
			allTracks = append(allTracks, &result.Tracks[i])
		}
	}
	return allTracks, nil
}

func (s *svc) get(ctx context.Context, token, url string, v interface{}) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("spotify API error: %s, body: %s", resp.Status, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(v)
}
