package services

import (
	"context"
	"fmt"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"qwen3-compatibility/internal/config"
	"qwen3-compatibility/internal/errors"
	"qwen3-compatibility/internal/models"
	"qwen3-compatibility/pkg/client"
)

type UploadService struct {
	client client.Uploader
	config *config.UploadConfig
}

func NewUploadService(client client.Uploader, uploadConfig *config.UploadConfig) *UploadService {
	return &UploadService{
		client: client,
		config: uploadConfig,
	}
}

// ValidateFile validates uploaded file
func (s *UploadService) ValidateFile(header *multipart.FileHeader) error {
	// Check file size
	if header.Size > s.config.MaxFileSize {
		return errors.NewFileSizeError(s.config.MaxFileSize)
	}

	// Detect content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// Fallback to detection by file extension
		ext := strings.ToLower(filepath.Ext(header.Filename))
		contentType = mime.TypeByExtension(ext)
	}

	if contentType == "" {
		return errors.NewFileTypeError(s.config.AllowedTypes)
	}

	// Check if content type is allowed
	if !s.isContentTypeAllowed(contentType) {
		return errors.NewFileTypeError(s.config.AllowedTypes)
	}

	return nil
}

// UploadFile uploads a file and returns the upload result
func (s *UploadService) UploadFile(ctx context.Context, apiKey string, file multipart.File, header *multipart.FileHeader, modelName string, validityHours int) (*models.UploadResult, error) {
	// Validate file first
	if err := s.ValidateFile(header); err != nil {
		return nil, err
	}

	// Get upload policy
	policy, err := s.client.GetUploadPolicy(ctx, apiKey, modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get upload policy: %w", err)
	}

	// Generate file name if empty
	fileName := header.Filename
	if fileName == "" {
		fileName = fmt.Sprintf("upload_%d", time.Now().Unix())
	}

	// Upload file to OSS
	ossURL, err := s.client.UploadToOSS(ctx, policy, file, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to OSS: %w", err)
	}

	// Calculate expire time
	expireTime := time.Now().Add(time.Duration(validityHours) * time.Hour)

	return &models.UploadResult{
		OSSURL:     ossURL,
		ExpireTime: expireTime,
		ModelUsed:  modelName,
	}, nil
}

// isContentTypeAllowed checks if the content type is in the allowed list
func (s *UploadService) isContentTypeAllowed(contentType string) bool {
	// Handle multipart content types
	if strings.HasPrefix(contentType, "multipart/") {
		return false
	}

	for _, allowedType := range s.config.AllowedTypes {
		if strings.EqualFold(contentType, allowedType) {
			return true
		}
	}

	return false
}
