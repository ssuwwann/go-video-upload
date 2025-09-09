package thumbnail

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"upload/internal/config"
	"upload/internal/exec"
)

type Generator struct {
	config config.Config
	runner exec.Runner
}

type Options struct {
	Count    int     // Number of thumbnails to generate
	Interval float64 // Interval between thumbnails in seconds (0 = auto)
	Width    int     // Thumbnail width (height auto-calculated)
	Quality  int     // JPEG quality (1-31, lower is better)
}

func NewGenerator(config config.Config, runner exec.Runner) *Generator {
	return &Generator{
		config: config,
		runner: runner,
	}
}

func DefaultOptions() Options {
	return Options{
		Count:    5,
		Interval: 0,
		Width:    320,
		Quality:  2,
	}
}

func (generator *Generator) GenerateThumbnails(context context.Context, videoID string, inputPath string, duration float64, options Options) ([]string, error) {
	thumbDir := filepath.Join(generator.config.StorageDir, "thumbnails", videoID)
	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		return nil, fmt.Errorf("create thumbnail dir: %w", err)
	}

	if options.Count <= 0 {
		options.Count = 5
	}

	if options.Width <= 0 {
		options.Width = 320
	}

	if options.Quality <= 0 || options.Quality > 31 {
		options.Quality = 2
	}

	// Calculate interval if not specified
	interval := options.Interval
	if interval <= 0 && duration > 0 {
		interval = duration / float64(options.Count+1)
	}

	var thumbnails []string

	// Generate thumbnails at intervals
	for i := 0; i < options.Count; i++ {
		timestamp := interval * float64(i+1)
		if timestamp >= duration {
			timestamp = duration * (float64(i) / float64(options.Count))
		}

		outputPath := filepath.Join(thumbDir, fmt.Sprintf("thumb_%03d.jpg", i+1))

		args := []string{
			"-y",
			"-ss", fmt.Sprintf("%.2f", timestamp),
			"-i", inputPath,
			"-vframes", "1",
			"-vf", fmt.Sprintf("scale=%d:-1", options.Width),
			"-q:v", fmt.Sprintf("%d", options.Quality),
			outputPath,
		}

		if _, err := generator.runner.Run(context, generator.config.FFmpegPath, args...); err != nil {
			return thumbnails, fmt.Errorf("generate thumbnail %d at %.2fs: %w", i+1, timestamp, err)
		}

		thumbnails = append(thumbnails, fmt.Sprintf("thumbnails/%s/thumb_%03d.jpg", videoID, i+1))
	}

	// Also generate a poster image from the first interesting frame
	posterPath := filepath.Join(thumbDir, "poster.jpg")
	posterArgs := []string{
		"-y",
		"-i", inputPath,
		"-vf", fmt.Sprintf("select='gt(scene,0.4)',scale=%d:-1", options.Width*2),
		"-frames:v", "1",
		"-q:v", fmt.Sprintf("%d", options.Quality),
		posterPath,
	}

	if _, err := generator.runner.Run(context, generator.config.FFmpegPath, posterArgs...); err != nil {
		thumbnails = append(thumbnails, fmt.Sprintf("thumbnails/%s/poster.jpg", videoID))
	}

	return thumbnails, nil
}

func (generator *Generator) GenerateSingleThumbnail(context context.Context, videoID string, inputPath string, timestamp float64) (string, error) {
	thumbDir := filepath.Join(generator.config.StorageDir, "thumbnails", videoID)
	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		return "", fmt.Errorf("create thumbnail dir: %w", err)
	}

	outputPath := filepath.Join(thumbDir, "preview.jpg")

	args := []string{
		"-y",
		"-ss", fmt.Sprintf("%.2f", timestamp),
		"-i", inputPath,
		"-vframes", "1",
		"-vf", "scale=640:-1",
		"-q:v", "2",
		outputPath,
	}

	if _, err := generator.runner.Run(context, generator.config.FFmpegPath, args...); err != nil {
		return "", fmt.Errorf("generate thumbnail at %.2fs: %w", timestamp, err)
	}

	return fmt.Sprintf("thumbnails/%s/preview.jpg", videoID), nil
}
