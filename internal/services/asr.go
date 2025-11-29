package services

import (
	"context"
	"time"

	"qwen3-compatibility/internal/models"
	"qwen3-compatibility/pkg/client"
)

type ASRService struct {
	client client.ASRProvider
}

func NewASRService(client client.ASRProvider) *ASRService {
	return &ASRService{
		client: client,
	}
}

// TranscribeAudio transcribes audio using DashScope ASR service
func (s *ASRService) TranscribeAudio(ctx context.Context, apiKey, audioURL, model string, language *models.SupportedLanguage, prompt string) (*models.ASRResponse, error) {
	return s.client.CallASR(ctx, apiKey, audioURL, model, language, true, prompt)
}

// ConvertToOpenAIFormat converts ASR response to OpenAI compatible format
func (s *ASRService) ConvertToOpenAIFormat(asrResponse *models.ASRResponse, processingTimeMs int64) *models.TranscriptionResponse {
	if asrResponse == nil || len(asrResponse.Output.Choices) == 0 {
		return &models.TranscriptionResponse{
			Text:           "",
			Task:           "transcribe",
			Language:       "unknown",
			ProcessingTime: processingTimeMs,
		}
	}

	choice := asrResponse.Output.Choices[0]
	var detectedLanguage string
	var duration float64

	// Extract language and emotion from annotations if available
	if len(choice.Message.Content) > 0 && len(choice.Message.Content[0].Text) > 0 {
		// Get detected language from annotations
		if len(choice.Message.Annotations) > 0 {
			detectedLanguage = string(choice.Message.Annotations[0].Language)
		}
	}

	// Get duration from usage
	if asrResponse.Usage.Seconds != nil {
		duration = *asrResponse.Usage.Seconds
	}

	text := ""
	if len(choice.Message.Content) > 0 {
		text = choice.Message.Content[0].Text
	}

	return &models.TranscriptionResponse{
		Text:           text,
		Task:           "transcribe",
		Language:       detectedLanguage,
		Duration:       duration,
		ProcessingTime: processingTimeMs,
	}
}

// CreateVerboseResponse creates a detailed response with metadata
func (s *ASRService) CreateVerboseResponse(asrResponse *models.ASRResponse, processingTimeMs int64, uploadInfo *models.UploadResult) *models.VerboseTranscriptionResponse {
	baseResponse := s.ConvertToOpenAIFormat(asrResponse, processingTimeMs)

	verboseResponse := &models.VerboseTranscriptionResponse{
		TranscriptionResponse: *baseResponse,
		RequestID:             asrResponse.Request,
		Timestamp:             "", // This will be set by the handler
	}

	// Add upload info if available
	if uploadInfo != nil {
		verboseResponse.UploadInfo = &models.UploadInfo{
			OSSURL:     uploadInfo.OSSURL,
			ExpireTime: uploadInfo.ExpireTime.Format(time.RFC3339),
			ModelUsed:  uploadInfo.ModelUsed,
		}
	}

	// Add ASR metadata if available
	if len(asrResponse.Output.Choices) > 0 {
		choice := asrResponse.Output.Choices[0]
		if len(choice.Message.Annotations) > 0 {
			annotation := choice.Message.Annotations[0]
			verboseResponse.ASRMetadata = &models.ASRMetadata{
				DetectedLanguage: string(annotation.Language),
				Emotion:          string(annotation.Emotion),
				FinishReason:     choice.FinishReason,
				Usage:            s.convertUsageInfo(asrResponse.Usage),
			}
		}
	}

	return verboseResponse
}

// convertUsageInfo converts ASR usage to usage info
func (s *ASRService) convertUsageInfo(usage models.ASRUsage) models.UsageInfo {
	inputTokens := 0
	if usage.InputTokensDetails != nil {
		inputTokens = usage.InputTokensDetails.TextTokens
	}

	audioSeconds := 0.0
	if usage.Seconds != nil {
		audioSeconds = *usage.Seconds
	}

	return models.UsageInfo{
		InputTokens:  inputTokens,
		OutputTokens: usage.OutputTokensDetails.TextTokens,
		AudioSeconds: audioSeconds,
	}
}
