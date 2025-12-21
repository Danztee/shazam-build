package spotify

type Track struct {
	Name       string   `json:"name"`
	Artists    []Artist `json:"artists"`
	Album      Album    `json:"album"`
	DurationMs int      `json:"duration_ms"`
}

type Artist struct {
	Name string `json:"name"`
}

type Album struct {
	Name string `json:"name"`
}

type URLInfo struct {
	Type string
	ID   string
}
