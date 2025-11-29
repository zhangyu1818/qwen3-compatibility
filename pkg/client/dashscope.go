package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"qwen3-compatibility/internal/errors"
	"qwen3-compatibility/internal/models"
)

type DashScopeClient struct {
	baseURL    string
	httpClient *http.Client
}

const (
	ASREndpoint   = "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation"
	UploadBaseURL = "https://dashscope.aliyuncs.com/api/v1/uploads"
)

func NewDashScopeClient(timeout int) *DashScopeClient {
	return &DashScopeClient{
		baseURL: ASREndpoint,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// GetUploadPolicy gets upload policy from DashScope
func (c *DashScopeClient) GetUploadPolicy(ctx context.Context, apiKey, modelName string) (*models.UploadPolicyData, error) {
	url := fmt.Sprintf("%s?action=getPolicy&model=%s", UploadBaseURL, modelName)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("Failed to create request: %v", err))
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewExternalServiceError("DashScope", err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.NewExternalServiceError("DashScope", fmt.Sprintf("Status: %d, Body: %s", resp.StatusCode, string(body)))
	}

	var uploadResp models.UploadPolicyResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("Failed to decode response: %v", err))
	}

	return &uploadResp.Data, nil
}

// UploadToOSS uploads file to OSS using the provided policy
func (c *DashScopeClient) UploadToOSS(ctx context.Context, policy *models.UploadPolicyData, file io.Reader, fileName string) (string, error) {
	// Create multipart form
	var requestBody bytes.Buffer
	contentType, err := createMultipartForm(&requestBody, policy, file, fileName)
	if err != nil {
		return "", errors.NewUploadError(err.Error())
	}

	req, err := http.NewRequestWithContext(ctx, "POST", policy.UploadHost, &requestBody)
	if err != nil {
		return "", errors.NewUploadError(fmt.Sprintf("Failed to create upload request: %v", err))
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.NewUploadError(fmt.Sprintf("Upload failed: %v", err))
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", errors.NewUploadError(fmt.Sprintf("Upload failed with status %d: %s", resp.StatusCode, string(body)))
	}

	// Return OSS URL
	key := fmt.Sprintf("%s/%s", policy.UploadDir, fileName)
	return fmt.Sprintf("oss://%s", key), nil
}

// CallASR calls the ASR service for transcription
func (c *DashScopeClient) CallASR(ctx context.Context, apiKey, audioURL, model string, language *models.SupportedLanguage, enableITN bool, prompt string) (*models.ASRResponse, error) {
	// Use prompt if provided, otherwise use space to maintain current behavior
	systemText := prompt
	if systemText == "" {
		systemText = " "
	}

	asrRequest := models.ASRRequest{
		Model: model,
		Input: models.ASRInput{
			Messages: []models.ASRMessage{
				{
					Role: "system",
					Content: []models.ASRContent{
						{Text: systemText},
					},
				},
				{
					Role: "user",
					Content: []models.ASRContent{
						{Audio: audioURL}, // Only audio, no text field
					},
				},
			},
		},
		Parameters: models.ASRParameters{
			ASROptions: models.ASROptions{
				EnableITN: enableITN,
			},
		},
	}

	// Set language if specified
	if language != nil {
		asrRequest.Parameters.ASROptions.Language = *language
	}

	jsonData, err := json.Marshal(asrRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("Failed to marshal ASR request: %v", err))
	}

	// Log request for debugging
	// Log request for debugging
	log.Printf("[DEBUG] ASR Request URL: %s\n", c.baseURL)
	log.Printf("[DEBUG] ASR Request Body: %s\n", string(jsonData))

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("Failed to create ASR request: %v", err))
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-DashScope-OssResourceResolve", "enable")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewExternalServiceError("DashScope ASR", err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[ERROR] ASR service error - Status: %d, Response: %s\n", resp.StatusCode, string(body))
		return nil, errors.NewExternalServiceError("DashScope ASR", fmt.Sprintf("Status: %d, Body: %s", resp.StatusCode, string(body)))
	}

	var asrResponse models.ASRResponse
	if err := json.NewDecoder(resp.Body).Decode(&asrResponse); err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("Failed to decode ASR response: %v", err))
	}

	return &asrResponse, nil
}

// createMultipartForm creates multipart form data for OSS upload using standard library
func createMultipartForm(writer io.Writer, policy *models.UploadPolicyData, file io.Reader, fileName string) (string, error) {
	key := fmt.Sprintf("%s/%s", policy.UploadDir, fileName)

	// Create multipart writer
	multipartWriter := multipart.NewWriter(writer)

	// Ensure multipart writer is closed
	// Note: errors from Close() are generally not critical for writing
	defer func() {
		_ = multipartWriter.Close()
	}()

	// Add form fields in the required order
	fields := []struct {
		name  string
		value string
	}{
		{"OSSAccessKeyId", policy.OSSAccessKeyID},
		{"Signature", policy.Signature},
		{"policy", policy.Policy},
		{"x-oss-object-acl", policy.XOSSObjectACL},
		{"x-oss-forbid-overwrite", policy.XOSSForbidOverwrite},
		{"key", key},
		{"success_action_status", "200"},
	}

	// Write all form fields
	for _, field := range fields {
		if err := multipartWriter.WriteField(field.name, field.value); err != nil {
			return "", fmt.Errorf("failed to write field %s: %w", field.name, err)
		}
	}

	// Add file
	part, err := multipartWriter.CreateFormFile("file", fileName)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	// Return content type
	return multipartWriter.FormDataContentType(), nil
}
