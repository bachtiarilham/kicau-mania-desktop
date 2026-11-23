package model

type Record map[string]any
type Config struct {
	LocalURL      string `json:"local_url"`
	RTSPBase      string `json:"rtsp_base"`
	CloudURL      string `json:"cloud_url"`
	VLC           string `json:"vlc"`
	LiveCacheMS   int    `json:"live_cache_ms"`
	RecordingsDir string `json:"recordings_dir,omitempty"`
}
type Rect struct {
	X            int  `json:"x"`
	Y            int  `json:"y"`
	Width        int  `json:"width"`
	Height       int  `json:"height"`
	ClipX        int  `json:"clip_x"`
	ClipY        int  `json:"clip_y"`
	ClipWidth    int  `json:"clip_width"`
	ClipHeight   int  `json:"clip_height"`
	Visible      bool `json:"visible"`
	DevicePixels bool `json:"device_pixels,omitempty"`
}
type PlayerStatus struct {
	ChannelID string  `json:"channel_id"`
	State     string  `json:"state"`
	Message   string  `json:"message,omitempty"`
	Codec     string  `json:"codec,omitempty"`
	FPS       float64 `json:"fps"`
	Decoded   int32   `json:"decoded"`
	Displayed int32   `json:"displayed"`
	Lost      int32   `json:"lost"`
}
