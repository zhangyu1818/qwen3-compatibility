package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"qwen3-compatibility/internal/config"
	"qwen3-compatibility/internal/middleware"
	"qwen3-compatibility/internal/models"
	"qwen3-compatibility/internal/services"
)

type TranscriptionHandler struct {
	uploadService services.IUploadService
	asrService    services.IASRService
	config        *config.Config
}

func NewTranscriptionHandler(uploadService services.IUploadService, asrService services.IASRService, cfg *config.Config) *TranscriptionHandler {
	return &TranscriptionHandler{
		uploadService: uploadService,
		asrService:    asrService,
		config:        cfg,
	}
}

// Transcription handles the /v1/audio/transcriptions endpoint
func (h *TranscriptionHandler) Transcription(c *gin.Context) {
	startTime := time.Now()

	// Parse form data
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Failed to get file from form",
		})
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close file: %v", err)
		}
	}()

	// Extract API key from context
	apiKey, exists := c.Get(middleware.APIKeyContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Missing API key in context",
		})
		return
	}
	apiKeyStr, ok := apiKey.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Invalid API key type",
		})
		return
	}

	// Parse other form fields
	model := c.PostForm("model")
	language := c.PostForm("language")
	prompt := c.PostForm("prompt") // Optional: contextual information for transcription

	// Validate model is provided
	if model == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "model parameter is required",
		})
		return
	}

	log.Printf("Transcription request: file=%s, model=%s, language=%s, prompt=%s, size=%d",
		header.Filename, model, language, prompt, header.Size)

	// Validate language if provided
	var languagePtr *models.SupportedLanguage
	if language != "" && !models.IsValidLanguage(language) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Unsupported language",
		})
		return
	}
	if language != "" {
		lang := models.SupportedLanguage(language)
		languagePtr = &lang
	}

	// Upload file
	uploadResult, err := h.uploadService.UploadFile(c.Request.Context(), apiKeyStr, file, header, model, 48) // 48 hours default
	if err != nil {
		log.Printf("File upload failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "File upload failed",
		})
		return
	}

	log.Printf("File uploaded successfully: %s, expires: %s", uploadResult.OSSURL, uploadResult.ExpireTime.Format(time.RFC3339))

	// Call ASR service with prompt
	asrResponse, err := h.asrService.TranscribeAudio(c.Request.Context(), apiKeyStr, uploadResult.OSSURL, model, languagePtr, prompt)
	if err != nil {
		log.Printf("ASR service failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Transcription failed",
		})
		return
	}

	// Calculate processing time
	processingTimeMs := time.Since(startTime).Milliseconds()

	// Always return JSON format
	response := h.asrService.ConvertToOpenAIFormat(asrResponse, processingTimeMs)

	// Add upload info to response
	uploadInfo := &models.UploadInfo{
		OSSURL:     uploadResult.OSSURL,
		ExpireTime: uploadResult.ExpireTime.Format(time.RFC3339),
		ModelUsed:  uploadResult.ModelUsed,
	}

	response.UploadInfo = uploadInfo
	response.ProcessingTime = processingTimeMs

	c.JSON(http.StatusOK, response)
}
