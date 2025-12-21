package spotify

type Track struct {
	Name       string   `json:"name"`
	Artists    []Artist `json:"artists"`
	Album      Album    `json:"album"`
	DurationMs int      `json:"duration_ms"`
}

type Artist struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Images []Image `json:"images"`
}

type Album struct {
	Name   string  `json:"name"`
	Images []Image `json:"images"`
}

type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type URLInfo struct {
	Type string
	ID   string
}
