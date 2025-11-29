# Qwen3 Compatibility Server

A Go implementation providing OpenAI-compatible APIs for Qwen3 services.

## Supported Interfaces

| Interface | Endpoint | Description | Status |
|-----------|----------|-------------|--------|
| **Audio Transcription** | `/v1/audio/transcriptions` | Convert audio/video to text (ASR) | ✅ Supported |

## Features

- **OpenAI Compatibility**: Drop-in replacement for OpenAI client libraries
- **Stateless Architecture**: No local state, fully scalable
- **Multi-Modal Support**: Handles audio, video, and text (future)
- **Production Ready**: Graceful shutdown, structured logging, error handling

## Installation

### Homebrew (macOS/Linux)

```bash
# Install via Homebrew Tap
brew tap zhangyu1818/tap
brew install qwen3-compatibility

# Start as a background service
brew services start zhangyu1818/tap/qwen3-compatibility
```

## Quick Start

### Prerequisites

- Go 1.21+
- DashScope API key (passed via Authorization header)

### Development

**Note**: API Key must be provided via the `Authorization` header in requests, not via environment variables.

```bash
# Build and run
go run ./cmd/server

# Or with custom port
go run ./cmd/server --port=9001
```

### Build

```bash
go build ./cmd/server
```

### Run

```bash
# Run server (no API key needed in environment)
./server

# Or on custom port
./server --port=9001
```

### Usage Examples

```bash
# Start server on custom port
./server --port=9001

# Or use the 'server' subcommand (both work the same)
./server server
```

## Configuration

The server supports configuration via:

1. **CLI flags** (highest priority)
2. **Environment variables** (with `QWEN_COMPAT_` prefix)
3. **Configuration file** (`./configs/config.yaml`)

**Important**: API Key is **NOT** configured via environment variables or config files. It must be provided in the `Authorization` header of each request.

**Important**: Model parameter is **REQUIRED** in each request. There is no default model.

### Environment Variables

- `QWEN_COMPAT_SERVER_PORT` - Server port (default: 9000)
- `QWEN_COMPAT_SERVER_HOST` - Server host (default: 0.0.0.0)

### Configuration File

See `configs/config.yaml` for default configuration.

## API Reference

### 1. Audio Transcription

Convert audio or video files to text. Compatible with OpenAI's `audio/transcriptions` endpoint.

**Endpoint**: `POST /v1/audio/transcriptions`

**Authentication**:
- Header: `Authorization: Bearer <your_dashscope_api_key>`

**Parameters**:

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `file` | File | **Yes** | The audio/video file to transcribe. | `@audio.mp3` |
| `model` | String | **Yes** | ID of the model to use. | `qwen3-asr-flash` |
| `language` | String | No | Language code (ISO-639-1). | `zh`, `en` |
| `prompt` | String | No | Optional text to guide the model's style. | `Keywords: AI, ML` |
| `response_format` | String | No | Format of the response. | `json` (default), `text` |

**Request Example**:
```bash
curl -X POST http://localhost:9000/v1/audio/transcriptions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -F "file=@meeting.mp3" \
  -F "model=qwen3-asr-flash" \
  -F "language=zh"
```

**Response Example**:
```json
{
  "text": "Hello, this is a transcription test.",
  "task": "transcribe",
  "language": "en",
  "duration": 12.5,
  "processing_time_ms": 500
}
```

## Supported Languages

- `zh` - Chinese
- `yue` - Cantonese
- `en` - English
- `ja` - Japanese
- `de` - German
- `ko` - Korean
- `ru` - Russian
- `fr` - French
- `pt` - Portuguese
- `ar` - Arabic
- `it` - Italian
- `es` - Spanish
- `hi` - Hindi
- `id` - Indonesian
- `th` - Thai
- `tr` - Turkish
- `uk` - Ukrainian
- `vi` - Vietnamese

## Supported File Types

### Audio
- AAC, AMR, FLAC, MP3, MPEG, M4A, OGG, Opus, WAV, WebM, WMA

### Video
- AVI, FLV, MKV, MOV, MP4, MPEG, WebM, WMV

### File Size Limit
- Maximum: 100MB (fixed)

## Project Structure

```
qwen3-compatibility/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── handlers/        # HTTP handlers
│   ├── services/        # Business logic services
│   ├── models/          # Data models
│   ├── middleware/      # HTTP middleware
│   └── errors/          # Error handling
├── pkg/client/          # External API client
├── configs/             # Configuration files
└── README.md
```

## Dependencies

- [gin-gonic/gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [spf13/viper](https://github.com/spf13/viper) - Configuration management

## License