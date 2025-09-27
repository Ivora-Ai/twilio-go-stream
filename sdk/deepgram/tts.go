package deepgram

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	msginterfaces "github.com/deepgram/deepgram-go-sdk/pkg/api/speak/v1/websocket/interfaces"
	interfaces "github.com/deepgram/deepgram-go-sdk/pkg/client/interfaces/v1"
	speak "github.com/deepgram/deepgram-go-sdk/pkg/client/speak"
	websocketv1 "github.com/deepgram/deepgram-go-sdk/pkg/client/speak/v1/websocket"
	"github.com/gorilla/websocket"
)

type MyCallback struct {
	wsConn         *websocket.Conn
	streamId       string
	audioBuffer    Queue
	ChanBuff       chan []byte
	dgClient       *websocketv1.WSCallback
	exit           chan struct{}
	writeMutex     sync.Mutex
	cancelWriter   context.CancelFunc
	stopProcessing bool // New flag to stop sending audio
}

func (c *MyCallback) Disconnect() {
	fmt.Println("[Disconnect] Cleaning up resources...")

	// Stop Deepgram client
	if c.dgClient != nil {
		c.dgClient.Stop()
	}

	if c.wsConn != nil {
		_ = c.wsConn.Close()

	}

	fmt.Println("[Disconnect] Cleanup complete. All resources freed.")
}

func (c MyCallback) Open(or *msginterfaces.OpenResponse) error {
	fmt.Println("[Open] Connection Established")
	return nil
}

func (c MyCallback) Metadata(md *msginterfaces.MetadataResponse) error {
	fmt.Printf("[Metadata] Request ID: %s\n", strings.TrimSpace(md.RequestID))
	return nil
}

func (c MyCallback) Binary(byMsg []byte) error {
	fmt.Println("[Binary] Received audio data")
	c.ChanBuff <- byMsg
	return nil
}

func (c MyCallback) Flush(fl *msginterfaces.FlushedResponse) error {
	fmt.Println("[Flushed] Received")
	return nil
}

func (c MyCallback) Clear(fl *msginterfaces.ClearedResponse) error {
	fmt.Println("[Cleared] Received")
	c.dgClient.Finish()
	return nil
}

func (c MyCallback) Close(cr *msginterfaces.CloseResponse) error {
	fmt.Println("[Close] Connection closed by Deepgram")
	return nil
}

func (c MyCallback) Warning(wr *msginterfaces.WarningResponse) error {
	fmt.Printf("[Warning] Code: %s | Description: %s\n", wr.WarnCode, wr.WarnMsg)
	return nil
}

func (c MyCallback) Error(er *msginterfaces.ErrorResponse) error {
	fmt.Printf("[Error] Code: %s | Description: %s\n", er.ErrCode, er.ErrMsg)
	return nil
}

func (c MyCallback) UnhandledEvent(byData []byte) error {
	fmt.Printf("[UnhandledEvent] %s\n", string(byData))
	return nil
}

func (c *MyCallback) ConnectTTS() {
	if c.dgClient == nil {
		fmt.Println("Deepgram client is not initialized.")
		return
	}

	if !c.dgClient.Connect() {
		fmt.Println("Failed to connect to Deepgram")
		os.Exit(1)
	}
	fmt.Println("Connected to Deepgram TTS service")
}

// Singleton initialization
var (
	initOnce         sync.Once
	callbackInstance *MyCallback
)

func Init() *MyCallback {
	// initOnce.Do(func() {
	cOptions := &interfaces.ClientOptions{}
	ttsOptions := &interfaces.WSSpeakOptions{
		Model:      "aura-asteria-en",
		Encoding:   "mulaw",
		SampleRate: 8000,
	}
	callback := MyCallback{ChanBuff: make(chan []byte, 2000)}

	// Get Deepgram API key from environment
	apiKey := os.Getenv("DEEPGRAM_API_KEY")
	if apiKey == "" {
		fmt.Println("ERROR: DEEPGRAM_API_KEY environment variable not set")
		return &callback
	}

	dgClient, err := speak.NewWSUsingCallback(context.Background(), apiKey, cOptions, ttsOptions, callback)
	if err != nil {
		fmt.Println("ERROR creating TTS connection:", err)
		return &callback
	}

	callback.dgClient = dgClient
	fmt.Println("Deepgram TTS client initialized")
	callbackInstance = &callback
	//	})
	return callbackInstance
}

func (c *MyCallback) StreamTTSDeepGram(ctx context.Context, m string, wsConn *websocket.Conn, sid string) {
	c.wsConn = wsConn
	c.streamId = sid
	fmt.Println("Agent", m)

	// Stop previous writer if active
	if c.cancelWriter != nil {
		fmt.Println("[Pause] Stopping previous process...")
		c.stopProcessing = true            // Prevents audio sending
		time.Sleep(500 * time.Millisecond) // Small pause before restarting
		c.cancelWriter()
	}

	c.exit = make(chan struct{})
	var newCtx context.Context
	newCtx, c.cancelWriter = context.WithCancel(ctx)

	// Clear buffer to prevent duplicate audio
	c.clearChannelBuffer()

	// Reset processing flag and start audio streaming
	c.stopProcessing = false
	go c.PushAudioToWs(newCtx)

	if c.dgClient == nil {
		fmt.Println("It's nill")
	}
	if err := c.dgClient.SpeakWithText(m); err != nil {
		fmt.Printf("Error sending text input: %v\n", err)
		return
	}
	if err := c.dgClient.Flush(); err != nil {
		fmt.Printf("Error sending flush signal: %v\n", err)
		return
	}

	<-c.exit
	c.dgClient.Stop()
	fmt.Println("Streaming process finished.")
}

func (c *MyCallback) PushAudioToWs(ctx context.Context) {
	for {
		select {
		case <-c.exit:
			fmt.Println("[StreamTTSDeepGram] Exit signal received. Stopping processing.")
			return

		case <-ctx.Done():
			fmt.Println("[Stopping] Audio writer goroutine")
			return
		case audioData := <-c.ChanBuff:
			if c.stopProcessing {
				fmt.Println("[Skipped] Ignoring audio chunk due to interruption.")
				return
			}
			fmt.Println("[Playing] Audio data chunk")

			c.writeMutex.Lock()
			muLawAudio := audioData
			chunkSize := 160

			for i := 0; i < len(muLawAudio); i += chunkSize {
				end := i + chunkSize
				if end > len(muLawAudio) {
					end = len(muLawAudio)
				}
				chunk := muLawAudio[i:end]
				payload := base64.StdEncoding.EncodeToString(chunk)

				message := map[string]interface{}{
					"event":     "media",
					"streamSid": c.streamId,
					"media": map[string]string{
						"track":   "audio",
						"payload": payload,
					},
				}

				if err := c.wsConn.WriteJSON(message); err != nil {
					fmt.Println("[Error] Failed to send WebSocket message:", err)
					c.writeMutex.Unlock()
					return
				}

				time.Sleep(18 * time.Millisecond) // Smooth audio streaming

				// Stop inner loop if interruption is detected
				if c.stopProcessing {
					fmt.Println("[Paused] Stopped sending further audio chunks.")
					break
				}
			}
			c.writeMutex.Unlock()
		}
	}
}

// Clears all old audio chunks from the buffer
func (c *MyCallback) clearChannelBuffer() {
	for len(c.ChanBuff) > 0 {
		<-c.ChanBuff
	}
	fmt.Println("[Cleared] Old audio buffer cleaned.")
}
