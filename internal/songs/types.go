package songs

type createSongPayload struct {
	SpotifyUrl string `json:"spotify_url"`
	Download   bool   `json:"download,omitempty"`
	SavePath   string `json:"save_path,omitempty"`
}
