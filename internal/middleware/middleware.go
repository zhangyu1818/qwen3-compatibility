package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"qwen3-compatibility/internal/errors"
	"qwen3-compatibility/internal/models"
)

const APIKeyContextKey = "api_key"

// AuthMiddleware extracts and validates the API key from the Authorization header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Missing Authorization header",
			})
			c.Abort()
			return
		}

		// Expected format: "Bearer <api_key>"
		const bearerPrefix = "Bearer "
		if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid Authorization header format. Expected: Bearer <api_key>",
			})
			c.Abort()
			return
		}

		apiKey := authHeader[len(bearerPrefix):]
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "API key is empty",
			})
			c.Abort()
			return
		}

		// Store API key in context for downstream handlers
		c.Set(APIKeyContextKey, apiKey)
		c.Next()
	}
}

// ErrorHandler provides centralized error handling
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during the request
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			handleError(c, err.Err)
		}
	}
}

// Logger provides request logging
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// CORS provides Cross-Origin Resource Sharing middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Recovery handles panics and returns a 500 error
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Printf("Panic recovered: %s", err)
			handleError(c, fmt.Errorf("internal server error: %s", err))
		} else if err, ok := recovered.(error); ok {
			log.Printf("Panic recovered: %v", err)
			handleError(c, fmt.Errorf("internal server error: %v", err))
		} else {
			log.Printf("Panic recovered with unknown type: %v", recovered)
			handleError(c, fmt.Errorf("internal server error"))
		}
	})
}

// handleError handles different types of errors and returns appropriate responses
func handleError(c *gin.Context, err error) {
	// Check if it's an API error
	if apiErr, ok := errors.IsAPIError(err); ok {
		response := models.ErrorResponse{
			Error: apiErr.Message,
		}
		if apiErr.Details != "" {
			response.Error = fmt.Sprintf("%s: %s", response.Error, apiErr.Details)
		}

		c.JSON(apiErr.HTTPStatus(), response)
		return
	}

	// Handle other errors
	c.JSON(http.StatusInternalServerError, models.ErrorResponse{
		Error: "Internal server error",
	})
}
