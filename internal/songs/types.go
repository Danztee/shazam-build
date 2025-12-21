package songs

type createSongPayload struct {
	SpotifyUrl string `json:"spotify_url"`
	Download   bool   `json:"download,omitempty"`
}

type matchSongPayload struct {
	Data string `json:"data"`
}
