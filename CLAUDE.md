# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go server that provides OpenAI-compatible APIs for Qwen3 services, specifically focused on audio transcription (ASR) capabilities. The server acts as a compatibility layer between OpenAI's API format and Alibaba's DashScope Qwen3 services.

## Architecture

The application follows a clean architecture pattern with clear separation of concerns:

- **cmd/server/**: Application entry point using Cobra CLI framework
- **internal/**: Core application logic (not importable by other projects)
  - **config/**: Configuration management using Viper with support for files, environment variables, and CLI flags
  - **handlers/**: HTTP request handlers for OpenAI-compatible endpoints
  - **services/**: Business logic layer with interfaces for dependency injection
  - **models/**: Data models and type definitions for requests/responses
  - **middleware/**: HTTP middleware (logging, recovery, CORS, authentication)
  - **errors/**: Custom error types and handling
- **pkg/client/**: External API client for DashScope services

### Key Design Patterns

- **Interface-based design**: Services use interfaces (`IUploadService`, `IASRService`) to enable testing and flexibility
- **Dependency injection**: Services are injected into handlers, making the system loosely coupled
- **Configuration hierarchy**: CLI flags > environment variables > config file > defaults
- **OpenAI compatibility**: Response formats match OpenAI's API structure for drop-in replacement

## Development Commands

### Build and Run
```bash
# Build the application
go build ./cmd/server

# Run the server (default port 9000)
./server

# Run with custom port
./server --port=9001

# Or run directly with go run
go run ./cmd/server
go run ./cmd/server --port=9001
```

### Dependencies
```bash
# Download dependencies
go mod download

# Tidy up go.mod
go mod tidy

# Vendor dependencies (if needed)
go mod vendor
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests in a specific package
go test ./internal/services
```

### Linting and Formatting
```bash
# Format code
go fmt ./...

# Run go vet
go vet ./...

# Run golint (if installed)
golint ./...
```

## Configuration

The server uses a three-tier configuration system:

1. **CLI Flags** (highest priority): `--host`, `--port`
2. **Environment Variables** (prefix `QWEN_COMPAT_`): `QWEN_COMPAT_SERVER_PORT`, `QWEN_COMPAT_SERVER_HOST`
3. **Config File**: `./configs/config.yaml` (optional)

**Important**: API keys are NOT configured via environment variables or config files. They must be provided in the `Authorization: Bearer <api_key>` header of each request.

## Key Implementation Details

### Authentication
- Uses middleware.AuthMiddleware() to extract API keys from `Authorization: Bearer <key>` headers
- API keys are passed directly to DashScope services without storage

### File Upload Flow
1. Client uploads file to server endpoint
2. Server requests upload policy from DashScope
3. File is uploaded to OSS using the policy
4. OSS URL is used for ASR processing
5. Response is converted to OpenAI format

### Supported Models
- `qwen3-asr-flash`: Primary ASR model (required parameter in requests)

### OpenAI Compatibility
- Endpoint: `POST /v1/audio/transcriptions`
- Response format matches OpenAI's transcription API
- Supports standard parameters: `file`, `model`, `language`, `prompt`, `response_format`

## Error Handling

The application uses a centralized error handling system:
- Custom error types in `internal/errors/`
- Middleware.ErrorHandler() provides consistent error responses
- Errors are returned with appropriate HTTP status codes and messages in OpenAI format