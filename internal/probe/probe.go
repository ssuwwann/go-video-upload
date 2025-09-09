package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"upload/internal/config"
	"upload/internal/exec"
)

type Prober struct {
	config config.Config
	runner exec.Runner
}

type VideoInfo struct {
	Duration   float64
	Width      int
	Height     int
	FPS        float64
	Bitrate    int
	CodecName  string
	AudioCodec string
}

type ffprobeOutput struct {
	Streams []struct {
		CodecType  string `json:"codec_type"`
		CodecName  string `json:"codec_name"`
		Width      int    `json:"width,omitempty"`
		Height     int    `json:"height,omitempty"`
		RFrameRate string `json:"r_frame_rate,omitempty"`
		BitRate    string `json:"bit_rate,omitempty"`
		Duration   string `json:"duration,omitempty"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
		BitRate  string `json:"bit_rate"`
	} `json:"format"`
}

func NewProber(config config.Config, runner exec.Runner) *Prober {
	return &Prober{
		config: config,
		runner: runner,
	}
}

func (prober *Prober) ProbeVideo(context context.Context, videoPath string) (*VideoInfo, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoPath,
	}

	output, err := prober.runner.Run(context, prober.config.FFprobePath, args...)
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	var probeData ffprobeOutput
	if err := json.Unmarshal(output, &probeData); err != nil {
		return nil, fmt.Errorf("parse ffprobe output: %w", err)
	}

	info := &VideoInfo{}

	// Parse duration from format
	if probeData.Format.Duration != "" {
		if duration, err := strconv.ParseFloat(probeData.Format.Duration, 64); err == nil {
			info.Duration = duration
		}
	}

	// Parse bitrate from format
	if probeData.Format.BitRate != "" {
		if bitrate, err := strconv.Atoi(probeData.Format.BitRate); err == nil {
			info.Bitrate = bitrate / 1000 // Convert to kbps
		}
	}

	// Find video and audio streams
	for _, stream := range probeData.Streams {
		if stream.CodecType == "video" && info.Width == 0 {
			info.Width = stream.Width
			info.Height = stream.Height
			info.CodecName = stream.CodecName

			// Parse FPS
			if stream.RFrameRate != "" {
				parts := strings.Split(stream.RFrameRate, "/")
				if len(parts) == 2 {
					num, _ := strconv.ParseFloat(parts[0], 64)
					den, _ := strconv.ParseFloat(parts[1], 64)
					if den > 0 {
						info.FPS = num / den
					}
				}
			}

			// Use stream duration if format duration is missing
			if info.Duration == 0 && stream.Duration != "" {
				if duration, err := strconv.ParseFloat(stream.Duration, 64); err == nil {
					info.Duration = duration
				}
			}
		} else if stream.CodecType == "audio" && info.AudioCodec == "" {
			info.AudioCodec = stream.CodecName
		}
	}

	return info, nil
}
