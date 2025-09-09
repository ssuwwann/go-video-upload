package store

import "time"

type Status string

// go에선 공식적인 enum이 없고 아래와 같이 유사하게 지원한다고 한다.
const (
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusReady      Status = "ready"
	StatusFailed     Status = "failed"
)

type VariantMeta struct {
	Format      string    `json:"format"`
	Height      int       `json:"height"`
	BitrateKbps int       `json:"bitrate_kbps"`
	Path        string    `json:"path"`
	SizeBytes   int64     `json:"size_bytes"`
	ReadyAt     time.Time `json:"ready_at"`
}

type VideoMeta struct {
	ID               string        `json:"id"`
	OriginalFilename string        `json:"original_filename"`
	MIME             string        `json:"mime"`
	SizeBytes        int64         `json:"size_bytes"`
	ChecksumSHA256   string        `json:"checksum_sha256"`
	Status           Status        `json:"status"`
	ErrorMessage     string        `json:"error_message,omitempty"`
	DurationSec      float64       `json:"duration_sec,omitempty"`
	Width            int           `json:"width,omitempty"`
	Height           int           `json:"height,omitempty"`
	FPS              float64       `json:"fps,omitempty"`
	StorageBase      string        `json:"storage_base"`
	Variants         []VariantMeta `json:"variants"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}
