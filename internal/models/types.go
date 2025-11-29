package models

import "time"

// Supported languages
type SupportedLanguage string

const (
	LanguageZh  SupportedLanguage = "zh"
	LanguageYue SupportedLanguage = "yue"
	LanguageEn  SupportedLanguage = "en"
	LanguageJa  SupportedLanguage = "ja"
	LanguageDe  SupportedLanguage = "de"
	LanguageKo  SupportedLanguage = "ko"
	LanguageRu  SupportedLanguage = "ru"
	LanguageFr  SupportedLanguage = "fr"
	LanguagePt  SupportedLanguage = "pt"
	LanguageAr  SupportedLanguage = "ar"
	LanguageIt  SupportedLanguage = "it"
	LanguageEs  SupportedLanguage = "es"
	LanguageHi  SupportedLanguage = "hi"
	LanguageId  SupportedLanguage = "id"
	LanguageTh  SupportedLanguage = "th"
	LanguageTr  SupportedLanguage = "tr"
	LanguageUk  SupportedLanguage = "uk"
	LanguageVi  SupportedLanguage = "vi"
)

var SupportedLanguages = []SupportedLanguage{
	LanguageZh, LanguageYue, LanguageEn, LanguageJa, LanguageDe, LanguageKo, LanguageRu, LanguageFr,
	LanguagePt, LanguageAr, LanguageIt, LanguageEs, LanguageHi, LanguageId, LanguageTh, LanguageTr, LanguageUk, LanguageVi,
}

func IsValidLanguage(lang string) bool {
	for _, supportedLang := range SupportedLanguages {
		if string(supportedLang) == lang {
			return true
		}
	}
	return false
}

// Emotion types
type EmotionType string

const (
	EmotionSurprised EmotionType = "surprised"
	EmotionNeutral   EmotionType = "neutral"
	EmotionHappy     EmotionType = "happy"
	EmotionSad       EmotionType = "sad"
	EmotionDisgusted EmotionType = "disgusted"
	EmotionAngry     EmotionType = "angry"
	EmotionFearful   EmotionType = "fearful"
)

// DashScope ASR request
type ASRRequest struct {
	Model      string        `json:"model"`
	Input      ASRInput      `json:"input"`
	Parameters ASRParameters `json:"parameters"`
}

type ASRInput struct {
	Messages []ASRMessage `json:"messages"`
}

type ASRMessage struct {
	Role        string          `json:"role"`
	Content     []ASRContent    `json:"content"`
	Annotations []ASRAnnotation `json:"annotations,omitempty"`
}

type ASRContent struct {
	Text  string `json:"text,omitempty"`
	Audio string `json:"audio,omitempty"`
}

type ASRAnnotation struct {
	Language SupportedLanguage `json:"language"`
	Type     string            `json:"type"`
	Emotion  EmotionType       `json:"emotion"`
}

type ASRParameters struct {
	ASROptions ASROptions `json:"asr_options"`
}

type ASROptions struct {
	EnableITN bool              `json:"enable_itn"`
	Language  SupportedLanguage `json:"language,omitempty"`
}

// DashScope ASR response
type ASRResponse struct {
	Output  ASROutput `json:"output"`
	Usage   ASRUsage  `json:"usage"`
	Request string    `json:"request_id"`
}

type ASROutput struct {
	Choices []ASRChoice `json:"choices"`
}

type ASRChoice struct {
	FinishReason string     `json:"finish_reason"`
	Message      ASRMessage `json:"message"`
}

type ASRUsage struct {
	InputTokensDetails  *ASRInputTokensDetails `json:"input_tokens_details,omitempty"`
	OutputTokensDetails ASROutputTokensDetails `json:"output_tokens_details"`
	Seconds             *float64               `json:"seconds,omitempty"`
}

type ASRInputTokensDetails struct {
	TextTokens int `json:"text_tokens"`
}

type ASROutputTokensDetails struct {
	TextTokens int `json:"text_tokens"`
}

// Upload policy response
type UploadPolicyResponse struct {
	Data UploadPolicyData `json:"data"`
}

type UploadPolicyData struct {
	UploadDir           string `json:"upload_dir"`
	UploadHost          string `json:"upload_host"`
	OSSAccessKeyID      string `json:"oss_access_key_id"`
	Signature           string `json:"signature"`
	Policy              string `json:"policy"`
	XOSSObjectACL       string `json:"x_oss_object_acl"`
	XOSSForbidOverwrite string `json:"x_oss_forbid_overwrite"`
}

// Upload result
type UploadResult struct {
	OSSURL     string    `json:"oss_url"`
	ExpireTime time.Time `json:"expire_time"`
	ModelUsed  string    `json:"model_used"`
}

// OpenAI compatible transcription response
type TranscriptionResponse struct {
	Text           string              `json:"text"`
	Task           string              `json:"task"`
	Language       string              `json:"language"`
	Duration       float64             `json:"duration,omitempty"`
	Words          []TranscriptionWord `json:"words,omitempty"`
	UploadInfo     *UploadInfo         `json:"upload_info,omitempty"`
	ProcessingTime int64               `json:"processing_time_ms,omitempty"`
}

type TranscriptionWord struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type UploadInfo struct {
	OSSURL     string `json:"oss_url"`
	ExpireTime string `json:"expire_time"`
	ModelUsed  string `json:"model_used"`
}

// Verbose transcription response (for detailed format)
type VerboseTranscriptionResponse struct {
	TranscriptionResponse
	RequestID   string       `json:"request_id"`
	Timestamp   string       `json:"timestamp"`
	ASRMetadata *ASRMetadata `json:"asr_metadata,omitempty"`
}

type ASRMetadata struct {
	DetectedLanguage string    `json:"detected_language"`
	Emotion          string    `json:"emotion"`
	FinishReason     string    `json:"finish_reason"`
	Usage            UsageInfo `json:"usage"`
}

type UsageInfo struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	AudioSeconds float64 `json:"audio_seconds"`
}

// Error response
type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code,omitempty"`
}
