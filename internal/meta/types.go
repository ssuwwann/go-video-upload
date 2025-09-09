package meta

import "time"

type Variant struct {
	Format      string `json:"format"` // e.g., hls, mp4
	Height      int    `json:"height"` // 480, 720, 1080
	BitrateKbps int    `json:"bitrate_kbps"`
	PathOrPl    string `json:"path_or_playlist"`
	SizeBytes   int64  `json:"size_bytes"`
	ReadyAtUnix int64  `json:"ready_at,omitempty"`
}

type Metadata struct {
	ID               string    `json:"id"`
	OriginalFilename string    `json:"original_filename"`
	MIME             string    `json:"mime"`
	SizeBytes        int64     `json:"size_bytes"`
	ChecksumSHA256   string    `json:"checksum_sha256"`
	Status           string    `json:"status"` // queued, processing, ready, failed
	ErrorMessage     string    `json:"error_message,omitempty"`
	DurationSec      float64   `json:"duration_sec,omitempty"`
	Width            int       `json:"width,omitempty"`
	Height           int       `json:"height,omitempty"`
	FPS              float64   `json:"fps,omitempty"`
	StorageBase      string    `json:"storage_base"`
	Variants         []Variant `json:"variants"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
