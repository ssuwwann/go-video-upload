package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"upload/internal/config"
	"upload/internal/fsutil"
	"upload/internal/id"
	"upload/internal/meta"
	"upload/internal/processor"
)

type uploadResponse struct {
	ID string `json:"id"`
}

func main() {
	cfg := config.Load()

	// Ensure base storage dirs exist
	os.MkdirAll(fsutil.MetadataDir(cfg.StorageDir), 0o755)
	os.MkdirAll(filepath.Join(cfg.StorageDir, "originals"), 0o755)
	os.MkdirAll(filepath.Join(cfg.StorageDir, "outputs"), 0o755)
	os.MkdirAll(filepath.Join(cfg.StorageDir, "thumbnails"), 0o755)

	store := meta.NewJSONStore(fsutil.MetadataDir(cfg.StorageDir))
	proc := processor.New(cfg, store)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	// Health
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Get video list
	e.GET("/videos", func(c echo.Context) error {
		videos, err := store.List()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list videos"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"videos": videos})
	})

	// Upload: accept a multipart file and create metadata
	e.POST("/videos", func(c echo.Context) error {
		fh, err := c.FormFile("file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "file is required"})
		}
		src, err := fh.Open()
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "cannot open file"})
		}
		defer src.Close()

		vid := id.New()
		origDir := filepath.Join(cfg.StorageDir, "originals", vid)
		if err := os.MkdirAll(origDir, 0o755); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cannot create dir"})
		}
		ext := filepath.Ext(fh.Filename)
		if ext == "" {
			ext = ".mp4"
		}
		dstPath := filepath.Join(origDir, "original"+ext)
		dst, err := os.Create(dstPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cannot save file"})
		}
		n, cErr := io.Copy(dst, src)
		dErr := dst.Close()
		if cErr != nil || dErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cannot write file"})
		}

		// Get MIME type from header first
		mimeType := fh.Header.Get("Content-Type")

		// If header doesn't have MIME type, detect from extension
		if mimeType == "" || mimeType == "application/octet-stream" {
			switch strings.ToLower(ext) {
			case ".mp4":
				mimeType = "video/mp4"
			case ".mov":
				mimeType = "video/quicktime"
			case ".avi":
				mimeType = "video/x-msvideo"
			case ".mkv":
				mimeType = "video/x-matroska"
			case ".webm":
				mimeType = "video/webm"
			case ".wmv":
				mimeType = "video/x-ms-wmv"
			case ".flv":
				mimeType = "video/x-flv"
			default:
				mimeType = "video/mp4" // default fallback
			}
		}

		m := meta.Metadata{
			ID:               vid,
			OriginalFilename: fh.Filename,
			MIME:             mimeType,
			SizeBytes:        n,
			ChecksumSHA256:   "",
			Status:           "queued",
			StorageBase:      cfg.StorageDir,
			Variants:         []meta.Variant{},
		}
		if err := store.Create(m); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cannot write metadata"})
		}

		// Start background processing
		go proc.ProcessVideo(vid)

		return c.JSON(http.StatusOK, uploadResponse{ID: vid})
	})

	e.GET("/videos/:id", func(c echo.Context) error {
		vid := c.Param("id")
		m, err := store.Get(vid)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}

		return c.JSON(http.StatusOK, m)
	})

	e.GET("/videos/:id/master.m3u8", func(c echo.Context) error {
		vid := c.Param("id")
		p := filepath.Join(cfg.StorageDir, "outputs", vid, "master.m3u8")
		if _, err := os.Stat(p); err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}

		return c.File(p)
	})

	// Serve thumbnails
	e.Static("/thumbnails", filepath.Join(cfg.StorageDir, "thumbnails"))

	// Serve static HLS segments under /streams/:id/
	e.Static("/streams", filepath.Join(cfg.StorageDir, "outputs"))

	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
