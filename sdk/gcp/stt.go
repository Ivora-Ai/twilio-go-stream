package gcp

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	speech "cloud.google.com/go/speech/apiv1"
	"github.com/gorilla/websocket"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

const (
	sampleRate   = 8000 // Twilio sends μ-law audio at 8kHz
	chunkSize    = 320  // Twilio audio chunk size (~20ms audio per chunk)
	channelCount = 1
)

// GoogleSTTClient handles the STT streaming with a callback approach similar to Deepgram
type GoogleSTTClient struct {
	client         *speech.Client
	stream         speechpb.Speech_StreamingRecognizeClient
	ctx            context.Context
	errChan        chan error
	DataChan       chan []byte
	WsConn         *websocket.Conn
	Sid            string
	AgentResponse  func(bool, string)
	lastTranscript string
}

func (c *GoogleSTTClient) Close() {
	c.client.Close()
}

// NewGoogleSTTClient initializes Google Cloud STT
func NewGoogleSTTClient() (*GoogleSTTClient, error) {
	// Open Google Cloud Speech Client
	ctx := context.Background()
	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Printf("Failed to create Google Speech client: %v", err)
		return nil, err
	}

	return &GoogleSTTClient{
		client:   client,
		ctx:      ctx,
		errChan:  make(chan error, 10),   // Buffer for up to 10 errors
		DataChan: make(chan []byte, 100), // Buffer for up to 100 audio chunks
	}, nil
}

func (c *GoogleSTTClient) StartStreamingAndAttach() {
	// Configure improved streaming request
	streamingConfig := &speechpb.StreamingRecognitionConfig{
		Config: &speechpb.RecognitionConfig{
			Encoding:                   speechpb.RecognitionConfig_LINEAR16,
			SampleRateHertz:            sampleRate,
			LanguageCode:               "en-IN",
			MaxAlternatives:            1,
			EnableAutomaticPunctuation: true,
			Model:                      "default", // Use 'phone_call' for telephony audio or 'default'
			UseEnhanced:                true,      // Higher quality for audio
			EnableWordTimeOffsets:      false,     // We don't need word timing
			EnableWordConfidence:       false,     // We don't need word confidence
			ProfanityFilter:            false,     // No profanity filter
			AlternativeLanguageCodes:   []string{"hi-IN"},
		},
		InterimResults:  true,  // Get partial results
		SingleUtterance: false, // Don't stop after first utterance
	}

	// Start streaming
	stream, err := c.client.StreamingRecognize(c.ctx)
	if err != nil {
		log.Printf("Failed to start streaming: %v", err)
		return
	}

	// Send initial configuration
	if err := stream.Send(&speechpb.StreamingRecognizeRequest{
		StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
			StreamingConfig: streamingConfig,
		},
	}); err != nil {
		log.Printf("Failed to send config: %v", err)
		return
	}

	c.stream = stream
	log.Println("Google STT stream initialized successfully")
}

// Transcribe processes Google STT responses with direct callback
func (c *GoogleSTTClient) Transcribe() {
	fmt.Println("Transcribing called")

	go func() {
		for {
			select {
			case <-c.ctx.Done():
				fmt.Println("Context cancelled, stopping transcription")
				return

			default:
				resp, err := c.stream.Recv()
				if err == io.EOF {
					fmt.Println("Stream closed by server (EOF)")
					return
				}
				if err != nil {
					log.Printf("Error receiving from stream: %v", err)
					return
				}

				// Process results and call agent response directly
				c.processResults(resp)
			}
		}
	}()
}

// processResults handles the STT response and calls the agent directly
func (c *GoogleSTTClient) processResults(resp *speechpb.StreamingRecognizeResponse) {
	if len(resp.Results) == 0 || len(resp.Results[0].Alternatives) == 0 {
		return // Skip empty results
	}

	for _, result := range resp.Results {
		if len(result.Alternatives) == 0 {
			continue
		}

		transcript := result.Alternatives[0].Transcript
		isFinal := result.IsFinal

		if isFinal {
			// Only trigger agent response on final results
			fmt.Printf("FINAL: %s %s\n", transcript, time.Now().UTC())
			c.lastTranscript = transcript

			// Only call agent if we have a websocket connection
			if c.WsConn != nil && c.Sid != "" && c.AgentResponse != nil {
				// Use the callback function directly
				c.AgentResponse(true, transcript)
			} else {
				log.Println("Cannot trigger agent response - missing websocket or callback")
			}
		} else {
			// Log interim results without triggering agent response
			fmt.Printf("Interim: %s %s\n", transcript, time.Now().UTC())
		}
	}
}

// PushAudioByte provides a Deepgram-like interface for pushing audio
func (c *GoogleSTTClient) PushAudioByte(data []byte) {
	select {
	case c.DataChan <- data:
		// Audio data sent to channel
	default:
		// Channel full - log occasionally to avoid spam
		log.Println("Audio buffer full, dropping packet")
	}
}

func (c *GoogleSTTClient) SendAudioInRealTime() {
	fmt.Println("SendAudioInRealTime called")

	// Create a buffer to accumulate audio data
	audioBuffer := make([]byte, 0, 8192) // 8KB initial capacity

	go func() {
		for {
			select {
			case <-c.ctx.Done():
				fmt.Println("Context cancelled, stopping audio streaming")
				return

			case receivedData, ok := <-c.DataChan:
				if !ok {
					// Channel closed
					fmt.Println("Data channel closed, stopping audio streaming")
					return
				}

				// Add received data to buffer
				audioBuffer = append(audioBuffer, receivedData...)

				// If we've accumulated enough data, send it
				// Google recommends 100ms-400ms chunks for better recognition
				if len(audioBuffer) >= 3200 { // ~200ms of audio at 16kHz
					err := c.stream.Send(&speechpb.StreamingRecognizeRequest{
						StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
							AudioContent: audioBuffer,
						},
					})

					if err != nil {
						log.Printf("Error sending audio to Google: %v", err)
						if err.Error() == "rpc error: code = Canceled desc = context canceled" {
							return
						}
					}

					// Clear buffer after sending
					audioBuffer = audioBuffer[:0]
				}

			// Check for streaming errors with a non-blocking select
			case err := <-c.errChan:
				log.Printf("Streaming error: %v", err)
				return
			}
		}
	}()
}
