package transcoder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"upload/internal/store"

	"upload/internal/config"
	"upload/internal/exec"
	"upload/internal/meta"
)

type Transcoder struct {
	config config.Config
	runner exec.Runner
	store  meta.Store
}

func NewTranscoder(config config.Config, runner exec.Runner, store meta.Store) *Transcoder {
	return &Transcoder{
		config: config,
		runner: runner,
		store:  store,
	}
}

type Resolution struct {
	Height       int
	VideoBitrate string
	AudioBitrate string
	MaxRate      string
	BufSize      string
}

var resolutions = []Resolution{
	{Height: 480, VideoBitrate: "600k", AudioBitrate: "96k", MaxRate: "900k", BufSize: "1200k"},
	{Height: 720, VideoBitrate: "1000k", AudioBitrate: "128k", MaxRate: "1500k", BufSize: "2000k"},
	{Height: 1080, VideoBitrate: "1800k", AudioBitrate: "128k", MaxRate: "2700k", BufSize: "3600k"},
}

func (transcoder *Transcoder) TranscodeVideo(context context.Context, videoID string) error {
	metadata, err := transcoder.store.Get(videoID)
	if err != nil {
		return fmt.Errorf("get metadata: %w", err)
	}

	metadata.Status = string(store.StatusProcessing)
	if err := transcoder.store.Update(metadata); err != nil {
		return fmt.Errorf("update metadata: %w", err)
	}

	inputPath := filepath.Join(transcoder.config.StorageDir, "originals", videoID, "original"+filepath.Ext(metadata.OriginalFilename))
	outputDir := filepath.Join(transcoder.config.StorageDir, "outputs", videoID)

	// 원본 해상도보다 낮은 해상도만 선택
	targetResolutions := selectResolutions(metadata.Height)

	for _, res := range targetResolutions {
		resDir := filepath.Join(outputDir, fmt.Sprintf("%d", res.Height))
		if err := os.MkdirAll(resDir, 0755); err != nil {
			return fmt.Errorf("create resolution dir: %w", err)
		}

		playlistPath := filepath.Join(resDir, "index.m3u8")
		segmentPath := filepath.Join(resDir, "%05d.ts")

		args := []string{
			"-y",
			"-i", inputPath,
			"-vf", fmt.Sprintf("scale=-2:%d", res.Height),
			"-c:v", "libx264",
			"-preset", "ultrafast", // 더 빠른 인코딩
			"-profile:v", "main",
			"-b:v", res.VideoBitrate,
			"-maxrate", res.MaxRate,
			"-bufsize", res.BufSize,
			"-g", "48",
			"-keyint_min", "48",
			"-sc_threshold", "0",
			"-c:a", "aac",
			"-b:a", res.AudioBitrate,
			"-threads", "2", // CPU 스레드 제한
			"-f", "hls",
			"-hls_time", "4",
			"-hls_playlist_type", "vod",
			"-hls_segment_filename", segmentPath,
			playlistPath,
		}

		if _, err := transcoder.runner.Run(context, transcoder.config.FFmpegPath, args...); err != nil {
			metadata.Status = string(store.StatusFailed)
			metadata.ErrorMessage = fmt.Sprintf("transcode %dp failed: %v", res.Height, err)
			transcoder.store.Update(metadata)

			return fmt.Errorf("transcode %dp: %w", res.Height, err)
		}

		varient := meta.Variant{
			Format:      "hls",
			Height:      res.Height,
			BitrateKbps: parseBitrate(res.VideoBitrate),
			PathOrPl:    fmt.Sprintf("%d/index.m3u8", res.Height),
		}
		metadata.Variants = append(metadata.Variants, varient)
	}

	if err := transcoder.generateMasterPlaylist(outputDir, targetResolutions); err != nil {
		return fmt.Errorf("generate master playlist: %w", err)
	}

	metadata.Status = string(store.StatusReady)
	if err := transcoder.store.Update(metadata); err != nil {
		return fmt.Errorf("update final status: %w", err)
	}

	return nil
}

func (transcoder *Transcoder) generateMasterPlaylist(outputDir string, targetResolutions []Resolution) error {
	masterContent := "#EXTM3U\n#EXT-X-VERSION:3\n\n"

	for _, res := range targetResolutions {
		bandwidth := parseBitrate(res.VideoBitrate)*1000 + parseBitrate(res.AudioBitrate)*1000
		masterContent += fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d\n", bandwidth, res.Height*16/9, res.Height)
		masterContent += fmt.Sprintf("%d/index.m3u8\n", res.Height)
	}
	masterPath := filepath.Join(outputDir, "master.m3u8")

	return os.WriteFile(masterPath, []byte(masterContent), 0644)
}

func parseBitrate(inputString string) int {
	var value int
	fmt.Sscanf(inputString, "%dk", &value)

	return value
}

// selectResolutions 원본 해상도보다 낮은 해상도들만 선택
func selectResolutions(originalHeight int) []Resolution {
	var selected []Resolution

	for _, res := range resolutions {
		if res.Height < originalHeight {
			selected = append(selected, res)
		}
	}

	// 원본이 너무 작으면 (360p 이하) 원본 그대로 사용
	if len(selected) == 0 {
		// 가장 낮은 해상도 하나만 사용
		selected = append(selected, resolutions[0]) // 480p
	}

	return selected
}
