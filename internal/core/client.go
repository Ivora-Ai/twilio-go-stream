package core

import (
	"context"
	"fmt"
	"sync"
	"time"
	audio_translator "twilio-go-stream/audio-translation"
	"twilio-go-stream/domain"
	"twilio-go-stream/sdk/dectector"
	"twilio-go-stream/sdk/deepgram"
	"twilio-go-stream/sdk/gcp"

	"github.com/gorilla/websocket"
)

type TTS interface {
	GetSpeech(string) ([]byte, error)
	Speaking(bool)
	// GetSpeechStreaming is optional and may be implemented for optimized streaming
}
type STT interface {
}

type VAD interface {
	Start()
	GetOutputChannel() chan string
	UserSpeakChannel() chan bool
	AgentSpeakChannel() chan bool
	Stop()
}
type LanguageProcessor interface {
	Chat(string) string
}

type Client struct {
	deepgramSTT  *deepgram.DeepgramSTTCallback
	prompt       *domain.Prompt
	deepgram     *deepgram.MyCallback
	timeSTTEND   time.Time
	timeTTSStart time.Time
	timeLLMEND   time.Time
	wsConn       *websocket.Conn
	streamID     string
	STT          *gcp.GoogleSTTClient
	tts          TTS
	// vad         VAD
	UserMessage         []string
	mu                  sync.Mutex
	ctx                 context.Context
	cancel              context.CancelFunc
	packetCount         int // Counter for audio packets
	InterruptAgentSpoke func(bool)
	Interrupt           *dectector.Interrupt
}

func Must(stt *gcp.GoogleSTTClient, tts TTS, deepgram *deepgram.MyCallback, deepgramSTT *deepgram.DeepgramSTTCallback) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		STT:         stt,
		tts:         tts,
		deepgram:    deepgram,
		prompt:      domain.InitPrompt(),
		deepgramSTT: deepgramSTT,
		ctx:         ctx,
		cancel:      cancel,
	}

	interrupt := &dectector.Interrupt{}
	deepgramSTT.UserSaid.UserSpeaking = interrupt.UserSpoke
	c.Interrupt = interrupt
	c.InterruptAgentSpoke = interrupt.AgentSpoke
	interrupt.AgentResponse = c.AgentResponse

	// Configure Google STT with the callback approach - similar to Deepgram
	if stt != nil && deepgramSTT == nil {
		// Set the callback parameters needed for the GoogleSTT client to call AgentResponse directly
		stt.AgentResponse = c.AgentResponse

		fmt.Println("Google STT configured with direct callback")
	}

	return c
}

// HandleTwilioAudio processes audio data from Twilio for Google STT
func (c *Client) HandleTwilioAudio(data []byte, stt *gcp.GoogleSTTClient) {
	// Convert μ-law to PCM16 format that Google STT requires
	pcm16Data := audio_translator.ConvertMuLawToPCM16(data)

	// Send to Google STT
	if stt != nil {
		stt.PushAudioByte(pcm16Data)
	}
}
