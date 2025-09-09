package processor

import (
	"context"
	"log"
	"path/filepath"
	"time"

	"upload/internal/config"
	"upload/internal/exec"
	"upload/internal/meta"
	"upload/internal/probe"
	"upload/internal/thumbnail"
	"upload/internal/transcoder"
)

type Processor struct {
	cfg    config.Config
	store  meta.Store
	runner exec.Runner
}

func New(cfg config.Config, store meta.Store) *Processor {
	return &Processor{
		cfg:    cfg,
		store:  store,
		runner: exec.NewCommandRunner(),
	}
}

func (p *Processor) ProcessVideo(videoID string) {
	ctx := context.Background()
	prober := probe.NewProber(p.cfg, p.runner)
	thumbGen := thumbnail.NewGenerator(p.cfg, p.runner)
	transcdr := transcoder.NewTranscoder(p.cfg, p.runner, p.store)

	log.Printf("Starting processing for video: %s", videoID)

	// Get metadata
	m, err := p.store.Get(videoID)
	if err != nil {
		log.Printf("Failed to get metadata for %s: %v", videoID, err)
		return
	}

	inputPath := filepath.Join(p.cfg.StorageDir, "originals", videoID, "original"+filepath.Ext(m.OriginalFilename))

	// Extract video info with FFprobe
	videoInfo, err := prober.ProbeVideo(ctx, inputPath)
	if err != nil {
		log.Printf("Failed to probe video %s (FFprobe not available): %v", videoID, err)
		// Skip video analysis if FFprobe is not available, but continue with basic metadata
		m.Status = "processing"
		m.ErrorMessage = "FFprobe not available, skipping video analysis"
		p.store.Update(m)

		// Try to generate thumbnails anyway (will also fail gracefully)
		thumbGen := thumbnail.NewGenerator(p.cfg, p.runner)
		thumbOpts := thumbnail.DefaultOptions()
		_, thumbErr := thumbGen.GenerateThumbnails(ctx, videoID, inputPath, 0, thumbOpts)
		if thumbErr != nil {
			log.Printf("Failed to generate thumbnails for %s: %v", videoID, thumbErr)
		}

		// Mark as failed since we can't process without FFmpeg
		m.Status = "failed"
		m.ErrorMessage = "FFmpeg/FFprobe not installed. Please install FFmpeg to enable video processing."
		p.store.Update(m)
		return
	}

	// Update metadata with video info
	m.DurationSec = videoInfo.Duration
	m.Width = videoInfo.Width
	m.Height = videoInfo.Height
	m.FPS = videoInfo.FPS
	m.UpdatedAt = time.Now()
	p.store.Update(m)

	// Generate thumbnails
	thumbOpts := thumbnail.DefaultOptions()
	thumbnails, err := thumbGen.GenerateThumbnails(ctx, videoID, inputPath, videoInfo.Duration, thumbOpts)
	if err != nil {
		log.Printf("Failed to generate thumbnails for %s: %v", videoID, err)
	} else {
		log.Printf("Generated %d thumbnails for %s", len(thumbnails), videoID)
	}

	// Start transcoding
	if err := transcdr.TranscodeVideo(ctx, videoID); err != nil {
		log.Printf("Failed to transcode video %s: %v", videoID, err)
		return
	}

	log.Printf("Successfully processed video: %s", videoID)
}
