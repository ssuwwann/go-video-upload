package middleware

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"upload/internal/store/config"

	"github.com/labstack/echo/v4"
)

type Validator struct {
	config config.Config
}

func NewValidator(config config.Config) *Validator {
	return &Validator{config: config}
}

func (validator *Validator) ValidateUpload() echo.MiddlewareFunc {

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			// only validate on upload endpoint
			if context.Request().Method != http.MethodPost || context.Path() != "/videos" {
				return next(context)
			}

			// Parse multipart form with size limit
			maxSize := int64(validator.config.MaxUploadMB) * 1024 * 1024
			context.Request().Body = http.MaxBytesReader(context.Response().Writer, context.Request().Body, maxSize)

			if err := context.Request().ParseMultipartForm(maxSize); err != nil {

				return context.JSON(http.StatusRequestEntityTooLarge, map[string]string{
					"error": fmt.Sprintf("file too large, max size is %d MB", validator.config.MaxUploadMB),
				})
			}

			// Get file header
			file, err := context.FormFile("file")
			if err != nil {
				return context.JSON(http.StatusBadRequest, map[string]string{
					"error": "file field is required",
				})
			}

			// Validate file size
			if file.Size > maxSize {
				return context.JSON(http.StatusRequestEntityTooLarge, map[string]string{
					"error": fmt.Sprintf("file size %d bytes exceeds maximum %d bytes", file.Size, maxSize),
				})
			}

			// Validate MIME type
			if !validator.isAllowedMIME(file) {
				return context.JSON(http.StatusBadRequest, map[string]string{
					"error": fmt.Sprintf("unsupported file type, allowed types: %v", validator.config.AllowedMIME),
				})
			}

			// Validate file extention
			if !validator.isAllowedExtension(file.Filename) {
				return context.JSON(http.StatusUnsupportedMediaType, map[string]string{
					"error": "unsupported file extension, allowed: .mp4, .mov, .mkv, .avi, .webm",
				})
			}

			return next(context)
		}
	}
}

func (validator *Validator) isAllowedMIME(file *multipart.FileHeader) bool {
	// Check content-type header
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		return false
	}

	// Remove parameters from MIME type (e.g., "video/mp4; codecs=avc1.42E01E")
	if index := strings.Index(contentType, ";"); index != -1 {
		contentType = strings.TrimSpace(contentType[:index])
	}

	for _, allowed := range validator.config.AllowedMIME {
		if strings.EqualFold(contentType, allowed) {
			return true
		}
	}

	return false
}

func (validator *Validator) isAllowedExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := []string{".mp4", ".mov", ".mkv", ".avi", ".webm", ".m4v", ".mpg", ".mpeg", ".wmv", ".flv"}

	for _, allowed := range allowedExts {
		if ext == allowed {
			return true
		}
	}

	return false
}

func (validator *Validator) RateLimiter(requestsPerMinute int) echo.MiddlewareFunc {
	// Simple in-memory rate limiter (for production, use Redis or mimilar)
	requests := make(map[string][]int64)

	return func(next echo.HandlerFunc) echo.HandlerFunc {

		return func(context echo.Context) error {
			// Only rate limit upload endpoint
			if context.Request().Method != http.MethodPost || context.Path() != "/videos" {
				return next(context)
			}

			clientIP := context.RealIP()
			now := context.Request().Context().Value("timestamp").(int64)
			minute := now / 60

			key := fmt.Sprintf("%s:%d", clientIP, minute)

			if count := len(requests[key]); count >= requestsPerMinute {
				return context.JSON(http.StatusTooManyRequests, map[string]string{
					"error": fmt.Sprintf("rate limit exceeded, max %d uploads per minute", requestsPerMinute),
				})
			}

			requests[key] = append(requests[key], now)

			// Clean old entires
			for k := range requests {
				parts := strings.Split(k, ":")
				if len(parts) == 2 {
					var keyMinute int64
					fmt.Sscanf(parts[1], "%d", &keyMinute)
					if keyMinute < minute-1 {
						delete(requests, k)
					}
				}
			}

			return next(context)
		}
	}
}
