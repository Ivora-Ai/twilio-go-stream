# Twilio Go Stream - Voice AI System

A voice-based AI system that connects phone calls with AI capabilities, using Twilio for telephony and various speech services for processing.

## Key Features

- Real-time voice conversations between callers and AI
- Speech-to-text and text-to-speech capabilities
- Streaming audio processing
- Support for multiple service providers (Deepgram, Google Cloud)

## GCP TTS
   -  `hi-IN-Chirp3-HD-Fenrirv`  -> 16 USD for 1M char(23 hour)
   -  `hi-IN-Standard-D`         -> 4 USD

## Environment Configuration

The application is configured using environment variables. You can set these directly or use a `.env` file:

### Speech Service Providers

You can configure different providers for Speech-to-Text (STT) and Text-to-Speech (TTS) services:

- **STT_PROVIDER**: Set to "deepgram" or "gcp" to choose the speech recognition service
- **TTS_PROVIDER**: Set to "deepgram" or "gcp" to choose the speech synthesis service

This allows you to mix and match providers based on your needs. For example, you might use Deepgram for STT and Google Cloud for TTS, or vice versa.

```
# Public URL for Twilio to connect to
PUBLIC_URL=your-domain.com

# Speech-to-Text provider ("deepgram" or "gcp")
STT_PROVIDER=deepgram

# Text-to-Speech provider ("deepgram" or "gcp") 
TTS_PROVIDER=deepgram

# Speech finalization delay in milliseconds (default: 1000)
SPEECH_FINAL_DELAY=1000

# Deepgram API key (required if using Deepgram for STT or TTS)
DEEPGRAM_API_KEY=your_deepgram_api_key

# Google Cloud credentials file path (required if using GCP for STT or TTS)
GOOGLE_APPLICATION_CREDENTIALS=sa.json

# Port to run the server on (default: 80)
PORT=80
```

## Running Locally

1. Clone the repository
2. Create a `.env` file with your configuration
3. Run the application:

```bash
go run main.go
```

## Docker Deployment

```bash
# Build the Docker image
docker build -t twilio-go-stream .

# Run the container
docker run -p 80:80 --env-file .env twilio-go-stream
```

## Architecture

The application is organized into several layers:

- **Handler**: HTTP and WebSocket endpoints
- **Core**: Business logic for conversation flow
- **SDK**: Integrations with speech services
- **Domain**: Data models and utilities

## Twilio Integration

To connect this service with Twilio:

1. Create a Twilio account and set up a phone number
2. Configure the webhook URL to point to your deployed instance:
   - Voice Request URL: `https://your-domain.com/incoming-call`
   - HTTP Method: POST

## License

MIT