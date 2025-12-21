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
	parsedURL, err := url.Parse(spotifyURL)
	if err == nil && parsedURL.Host == "open.spotify.com" {
		parts := strings.Split(strings.TrimPrefix(parsedURL.Path, "/"), "/")
		if len(parts) == 2 {
			urlType := parts[0]
			id := parts[1]
			if urlType == "track" || urlType == "album" {
				return URLInfo{Type: urlType, ID: id}, nil
			}
		}
	}

	trackRe := regexp.MustCompile(`track/([a-zA-Z0-9]+)`)
	trackMatches := trackRe.FindStringSubmatch(spotifyURL)
	if len(trackMatches) > 1 {
		return URLInfo{Type: "track", ID: trackMatches[1]}, nil
	}

	albumRe := regexp.MustCompile(`album/([a-zA-Z0-9]+)`)
	albumMatches := albumRe.FindStringSubmatch(spotifyURL)
	if len(albumMatches) > 1 {
		return URLInfo{Type: "album", ID: albumMatches[1]}, nil
	}

	return URLInfo{}, errors.New("could not extract track or album ID from URL")
}

func (s *svc) GetAccessToken() (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("spotify API error: %s, body: %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.AccessToken == "" {
		return "", errors.New("access token is empty")
	}

	return result.AccessToken, nil
}

func (s *svc) GetTrack(ctx context.Context, token, trackID string) (*Track, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.spotify.com/v1/tracks/"+trackID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("spotify API error: %s, body: %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var track Track
	if err := json.Unmarshal(body, &track); err != nil {
		return nil, err
	}

	return &track, nil
}

func (s *svc) GetAlbumTracks(ctx context.Context, token, albumID string) ([]*Track, *Album, error) {
	albumInfo, err := s.getAlbumInfo(ctx, token, albumID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get album info: %w", err)
	}

	trackIDs, err := s.getAlbumTrackIDs(ctx, token, albumID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get album track IDs: %w", err)
	}

	if len(trackIDs) == 0 {
		return []*Track{}, albumInfo, nil
	}

	tracks, err := s.getTracks(ctx, token, trackIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get track details: %w", err)
	}

	for _, track := range tracks {
		if albumInfo.Name != "" {
			track.Album.Name = albumInfo.Name
		}
	}

	return tracks, albumInfo, nil
}

func (s *svc) getAlbumTrackIDs(ctx context.Context, token, albumID string) ([]string, error) {
	var allTrackIDs []string
	offset := 0
	limit := 50

	for {
		url := fmt.Sprintf("https://api.spotify.com/v1/albums/%s/tracks?limit=%d&offset=%d", albumID, limit, offset)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("spotify API error: %s, body: %s", resp.Status, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var result struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
			Total int `json:"total"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}

		for _, item := range result.Items {
			allTrackIDs = append(allTrackIDs, item.ID)
		}

		if offset+len(result.Items) >= result.Total {
			break
		}

		offset += limit
	}

	return allTrackIDs, nil
}

func (s *svc) getAlbumInfo(ctx context.Context, token, albumID string) (*Album, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.spotify.com/v1/albums/"+albumID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("spotify API error: %s, body: %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var album struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &album); err != nil {
		return nil, err
	}

	return &Album{Name: album.Name}, nil
}

func (s *svc) getTracks(ctx context.Context, token string, trackIDs []string) ([]*Track, error) {
	if len(trackIDs) == 0 {
		return nil, errors.New("no track IDs provided")
	}

	const maxBatchSize = 50
	var allTracks []*Track

	for i := 0; i < len(trackIDs); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(trackIDs) {
			end = len(trackIDs)
		}

		batch := trackIDs[i:end]
		idsParam := strings.Join(batch, ",")
		url := fmt.Sprintf("https://api.spotify.com/v1/tracks?ids=%s", idsParam)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("spotify API error: %s, body: %s", resp.Status, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var result struct {
			Tracks []Track `json:"tracks"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}

		for i := range result.Tracks {
			allTracks = append(allTracks, &result.Tracks[i])
		}
	}

	return allTracks, nil
}
