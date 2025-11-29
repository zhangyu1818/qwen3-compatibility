package client

import (
	"context"
	"io"

	"qwen3-compatibility/internal/models"
)

// Uploader defines the interface for file upload operations
type Uploader interface {
	GetUploadPolicy(ctx context.Context, apiKey, modelName string) (*models.UploadPolicyData, error)
	UploadToOSS(ctx context.Context, policy *models.UploadPolicyData, file io.Reader, fileName string) (string, error)
}

// ASRProvider defines the interface for ASR operations
type ASRProvider interface {
	CallASR(ctx context.Context, apiKey, audioURL, model string, language *models.SupportedLanguage, enableITN bool, prompt string) (*models.ASRResponse, error)
}
