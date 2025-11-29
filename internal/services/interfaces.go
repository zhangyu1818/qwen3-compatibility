package services

import (
	"context"
	"mime/multipart"

	"qwen3-compatibility/internal/models"
)

// IUploadService defines the interface for upload service
type IUploadService interface {
	UploadFile(ctx context.Context, apiKey string, file multipart.File, header *multipart.FileHeader, modelName string, validityHours int) (*models.UploadResult, error)
}

// IASRService defines the interface for ASR service
type IASRService interface {
	TranscribeAudio(ctx context.Context, apiKey, audioURL, model string, language *models.SupportedLanguage) (*models.ASRResponse, error)
	ConvertToOpenAIFormat(asrResponse *models.ASRResponse, processingTimeMs int64) *models.TranscriptionResponse
	CreateVerboseResponse(asrResponse *models.ASRResponse, processingTimeMs int64, uploadInfo *models.UploadResult) *models.VerboseTranscriptionResponse
}
