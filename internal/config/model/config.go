package model

type Config struct {
	LocalURL      string `json:"local_url"`
	RTSPBase      string `json:"rtsp_base"`
	CloudURL      string `json:"cloud_url"`
	VLC           string `json:"vlc"`
	LiveCacheMS   int    `json:"live_cache_ms"`
	RecordingsDir string `json:"recordings_dir,omitempty"`
}
