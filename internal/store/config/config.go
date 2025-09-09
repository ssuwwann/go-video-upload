package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port        string
	StorageDir  string
	FFMPEGPath  string
	FFPROBEPath string
	MaxUploadMB int
	Workers     int
	AllowedMIME []string
	Resolutions []int
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

func FromEnv() Config {
	port := getenv("PORT", "1323")
	storage := getenv("STORAGE_DIR", "storage")
	ffmpeg := getenv("FFMPEG_PATH", "ffmpeg")
	ffprobe := getenv("FFPROBE_PATH", "ffprobe")

	maxUploadMB := 512
	if v := os.Getenv("MAX_UPLOAD_MB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxUploadMB = n
		}
	}

	workers := 1
	if v := os.Getenv("WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			workers = n
		}
	}

	allowed := []string{"video/mp4", "video/quicktime", "video/x-msvideo"}
	if v := os.Getenv("ALLOWED_MIME"); v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}

		if len(out) > 0 {
			allowed = out
		}
	}

	// default ladder 480/720/1080
	resolutions := []int{480, 720, 1080}
	if v := os.Getenv("RESOLUTIONS"); v != "" {
		parts := strings.Split(v, ",")
		res := make([]int, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}

			if p != "" {
				if n, err := strconv.Atoi(p); err == nil {
					res = append(res, n)
				}
			}

			if len(res) > 0 {
				resolutions = res
			}
		}

	}
	return Config{
		Port:        port,
		StorageDir:  storage,
		FFMPEGPath:  ffmpeg,
		FFPROBEPath: ffprobe,
		MaxUploadMB: maxUploadMB,
		Workers:     workers,
		AllowedMIME: allowed,
		Resolutions: resolutions,
	}
}
