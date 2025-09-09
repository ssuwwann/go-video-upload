package config

import (
	"os"
	"strconv"
	"strings"
)

// GetEnv returns env var or default when empty.
func GetEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

type Config struct {
	Port        string
	StorageDir  string
	FFmpegPath  string
	FFprobePath string
	MaxUploadMB int
	Workers     int
	AllowedMIME []string
}

// Load reads configuration from environment with sensible defaults.
func Load() Config {
	cfg := Config{
		Port:        GetEnv("PORT", "1323"),
		StorageDir:  GetEnv("STORAGE_DIR", "storage"),
		FFmpegPath:  GetEnv("FFMPEG_PATH", "E:\\suwan\\file\\ffmpeg-2025-09-08-git-45db6945e9-full_build\\bin\\ffmpeg.exe"),
		FFprobePath: GetEnv("FFPROBE_PATH", "E:\\suwan\\file\\ffmpeg-2025-09-08-git-45db6945e9-full_build\\bin\\ffprobe.exe"),
		MaxUploadMB: 512,
		Workers:     1,
		AllowedMIME: []string{"video/mp4", "video/quicktime", "video/x-matroska"},
	}
	if v := os.Getenv("MAX_UPLOAD_MB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxUploadMB = n
		}
	}
	if v := os.Getenv("WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Workers = n
		}
	}
	if v := os.Getenv("ALLOWED_MIME"); v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			s := strings.TrimSpace(p)
			if s != "" {
				out = append(out, s)
			}
		}
		if len(out) > 0 {
			cfg.AllowedMIME = out
		}
	}
	return cfg
}
