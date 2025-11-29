package errors

import (
	"fmt"
	"net/http"
)

// Custom error types
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error %d: %s", e.Code, e.Message)
}

func (e *APIError) HTTPStatus() int {
	return e.Code
}

// Error constructors
func NewInternalServerError(message string) *APIError {
	return &APIError{
		Code:    http.StatusInternalServerError,
		Message: message,
	}
}
func NewFileSizeError(maxSize int64) *APIError {
	return &APIError{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("File too large. Maximum size is %d bytes", maxSize),
	}
}

func NewFileTypeError(allowedTypes []string) *APIError {
	return &APIError{
		Code:    http.StatusBadRequest,
		Message: "Unsupported file type",
		Details: fmt.Sprintf("Allowed types: %v", allowedTypes),
	}
}
func NewExternalServiceError(service string, details string) *APIError {
	return &APIError{
		Code:    http.StatusBadGateway,
		Message: fmt.Sprintf("External service error: %s", service),
		Details: details,
	}
}

func NewUploadError(details string) *APIError {
	return &APIError{
		Code:    http.StatusInternalServerError,
		Message: "File upload failed",
		Details: details,
	}
}

// Check if error is APIError
func IsAPIError(err error) (*APIError, bool) {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr, true
	}
	return nil, false
}
