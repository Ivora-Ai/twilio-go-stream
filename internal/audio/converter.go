// Package audio provides audio processing functionality
package audio

import (
	"encoding/base64"
	"twilio-go-stream/audio-translation"
	"twilio-go-stream/internal/interfaces"
)

// Converter implements the AudioConverter interface
type Converter struct{}

// NewConverter creates a new audio converter
func NewConverter() *Converter {
	return &Converter{}
}

// ConvertToMuLaw converts PCM16 audio data to μ-law format
func (c *Converter) ConvertToMuLaw(data []byte) []byte {
	return audio_translator.ConvertPCM16ToMuLaw(data)
}

// ConvertFromMuLaw converts μ-law audio data to PCM16 format
func (c *Converter) ConvertFromMuLaw(data []byte) []byte {
	return audio_translator.ConvertMuLawToPCM16(data)
}

// DecodeBase64Audio decodes base64-encoded audio data
func DecodeBase64Audio(encodedData string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encodedData)
}

// EncodeBase64Audio encodes audio data to base64
func EncodeBase64Audio(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}