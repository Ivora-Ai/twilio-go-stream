package core

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"
	audio_translator "twilio-go-stream/audio-translation"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

// StreamGoogleTTS is the main entry point for Google TTS streaming
func (c *Client) StreamGoogleTTS(ctx context.Context, text, streamSid string, wsConn *websocket.Conn) error {
	start := time.Now().UTC()
	fmt.Println("Starting Google TTS streaming at", start)

	// Get the GoogleTTSClient from the interface if possible
	// googleTTS, ok := c.tts.(*gcp.GoogleTTSClient)

	// // If we can use the enhanced streaming API, use it
	// if ok { //tod comment this reduce billing, increse
	// 	fmt.Println("Using enhanced streaming for GCP TTS")
	// 	return c.streamGoogleTTSEnhanced(ctx, text, streamSid, wsConn, googleTTS)
	// }

	// Otherwise fall back to the legacy approach
	fmt.Println("Using legacy non-streaming for GCP TTS")
	return c.streamGoogleTTSLegacy(ctx, text, streamSid, wsConn)
}

// Legacy implementation using non-streaming API
func (c *Client) streamGoogleTTSLegacy(ctx context.Context, text, streamSid string, wsConn *websocket.Conn) error {
	log.Println("Using legacy GCP TTS (non-streaming)")
	start := time.Now().UTC()
	ttsToWs := true

	// Get the entire speech in one go
	ttsResp, err := c.tts.GetSpeech(cleanText(text))
	if err != nil {
		fmt.Println("Error getting speech:", err)
		return err
	}

	audioData := ttsResp

	fmt.Println("Time to speak ==>>>", time.Since(start), time.Now())

	// Convert PCM16 to μ-law (G.711)
	muLawAudio := audio_translator.ConvertPCM16ToMuLaw(audioData)

	c.InterruptAgentSpoke(true)
	// Stream Audio to Twilio WebSocket
	chunkSize := 160 // 20ms of 8kHz μ-law audio = 160 bytes

	log.Println("Streaming μ-law audio to Twilio...")
	c.tts.Speaking(true)

	for i := 0; i < len(muLawAudio); i += chunkSize {
		select {
		case <-ctx.Done():
			fmt.Println("Stopping old goroutine...")
			c.tts.Speaking(false)
			return nil // Exit if context is canceled
		default:
			end := i + chunkSize
			if end > len(muLawAudio) {
				end = len(muLawAudio)
			}
			chunk := muLawAudio[i:end]

			// Encode chunk in Base64
			payload := base64.StdEncoding.EncodeToString(chunk)

			// Prepare JSON message
			message := map[string]interface{}{
				"event":     "media",
				"streamSid": streamSid,
				"media": map[string]string{
					"track":   "audio",
					"payload": payload,
				},
			}

			if ttsToWs {
				fmt.Println("TTS -> WS time in ms ==>>>", time.Since(start), time.Now())
				ttsToWs = false
			}

			// Send message over WebSocket
			err = wsConn.WriteJSON(message)
			if err != nil {
				log.Println("Error sending WebSocket message:", err)
				return err
			}

			// Sleep for 16ms to match real-time streaming
			time.Sleep(16 * time.Millisecond)
		}
	}
	c.InterruptAgentSpoke(false)

	c.tts.Speaking(false)
	log.Println("TTS audio streaming completed.")
	return nil
}

// cleanText ensures the text is valid UTF-8
func cleanText(text string) string {
	if utf8.ValidString(text) {
		return text
	}
	v := make([]rune, 0, len(text))
	for i, r := range text {
		if r == utf8.RuneError {
			_, size := utf8.DecodeRuneInString(text[i:])
			if size == 1 {
				continue
			}
		}
		v = append(v, r)
	}
	return string(v)
}
