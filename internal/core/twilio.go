package core

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"twilio-go-stream/domain"
	language_processor "twilio-go-stream/sdk/language-processor"

	"github.com/gorilla/websocket"
)

func (c *Client) Talk(wsConn *websocket.Conn) {
	// Set Ping/Pong handler
	fmt.Println("Talk")
	// attach vad
	// c.vad.Start()
	fmt.Println("Attached")

	// Configure STT provider with WebSocket and callback
	if c.deepgramSTT != nil {
		// Deepgram STT setup
		c.deepgramSTT.WsConn = wsConn
		c.deepgramSTT.UserSaid.AgentResponse = c.AgentResponse
	} else if c.STT != nil {
		// Google STT setup - use callback approach like Deepgram
		c.STT.WsConn = wsConn
		c.STT.Sid = c.streamID // Will be set once we get it
		// c.STT.AgentResponse is already set in Must()
		fmt.Println("Using Google STT with callback instead of Deepgram STT")
	} else {
		fmt.Println("No STT provider available")
	}
	// fmt.Println("Running Vad")

	// wsConn.SetPingHandler(func(appData string) error {
	// 	log.Println("Received ping from Twilio, sending pong...")
	// 	return wsConn.WriteMessage(websocket.PongMessage, nil)
	// })
	fmt.Println("Pinging")
	// c.Routine(c.STT)
	fmt.Println("Lising audio")

	for {
		// Read message from WebSocket
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		// Parse incoming JSON
		var stream domain.MediaStream
		if err := json.Unmarshal(message, &stream); err != nil {
			log.Println("JSON unmarshal error:", err)
			continue
		}

		// Capture Stream SID and start audio sending
		if stream.Event == "start" && stream.StreamSid != "" {
			log.Printf("Stream SID received: %s\n", stream.StreamSid)
			c.streamID = stream.StreamSid
			c.wsConn = wsConn

			// Set Sid for the active STT provider
			if c.deepgramSTT != nil {
				c.deepgramSTT.Sid = stream.StreamSid
			} else if c.STT != nil {
				c.STT.Sid = stream.StreamSid
			}

			go func() {
				time.Sleep(1 * time.Second)
				c.AgentResponse(false, "Hello, how can I help you today?")
				// c.AgentResponse(false, wsConn, "Hello, how can I help you today?", stream.StreamSid)
				// c.AgentResponse(false, wsConn, "Hello, how can I help you today?", stream.StreamSid)
				// c.AgentResponse(false, wsConn, "Hello, how can I help you today?", stream.StreamSid)
				// todo this is practical scenario and system is not able to produce speech properly,
				// this might be related to 2 words skip issue, this can be solved by remoing context cancelled and using callback or channel instead
				// c.AgentResponse(false, wsConn, "Hello, how can I help you today?", stream.StreamSid)
			}()

		}

		// Handle incoming audio payload
		if stream.Event == "media" {

			decodedAudio, err := base64.StdEncoding.DecodeString(stream.Media.Payload)
			if err != nil {
				log.Printf("Error decoding audio: %v", err)
				continue
			}

			// Handle audio based on which STT provider is being used
			if c.deepgramSTT != nil {
				// Use Deepgram for STT (expects μ-law audio)
				c.deepgramSTT.PushAudioByte(decodedAudio)
			} else if c.STT != nil {
				// Use Google Cloud for STT (requires PCM16 audio)
				c.HandleTwilioAudio(decodedAudio, c.STT)
			}

			// Log occasionally to reduce noise
			if c.packetCount%100 == 0 {
				// log.Printf("Processed %d audio packets", c.packetCount)
			}
			c.packetCount++
		}

		// Handle call stop event
		if stream.Event == "stop" {
			log.Println("Call ended, closing WebSocket.")
			break
		}
	}
}

func (c *Client) AgentResponse(genAi bool, response string) {

	// message -> gen ai -> tts -> ws
	start := time.Now().UTC()
	fmt.Println("User Speech enved at", start)

	if genAi {
		// Start sending audio

		c.prompt.PushMessage("user", response)

		resp := language_processor.GetChatResponseFromGroq(c.prompt)
		msg := domain.ChatCompletion{}
		err := json.Unmarshal([]byte(resp), &msg)
		if err != nil {
			fmt.Println("Error", err)
		}
		if len(msg.Choices) == 0 {
			fmt.Println("No data from llm")
			return
		}
		response = msg.Choices[0].Message.Content
		c.prompt.PushMessage("system", response)
		c.timeLLMEND = time.Now().UTC()
		// fmt.Println("GenAI: ", c.prompt)

		if response == "close()" {
			// Safely disconnect services that are in use
			if c.deepgramSTT != nil {
				c.deepgramSTT.Disconnect()
			}
			if c.deepgram != nil {
				c.deepgram.Disconnect()
			}
			c.wsConn.Close()
		}
	}

	// for _,w := range breakSentenceByPunctuators(response) {

	// Cancel previous goroutine if it exists
	if c.cancel != nil {
		c.cancel()
	}

	// Create a new context for the new goroutine
	c.ctx, c.cancel = context.WithCancel(context.Background())
	// Start the new goroutine
	go func(ctx context.Context) {
		c.timeTTSStart = time.Now().UTC()

		// Use appropriate TTS provider
		if c.deepgram != nil {
			// Use Deepgram for TTS
			c.deepgram.StreamTTSDeepGram(ctx, response, c.wsConn, c.streamID)
		} else if c.tts != nil {
			// Use Google Cloud for TTS
			c.StreamGoogleTTS(ctx, response, c.streamID, c.wsConn)
		}

		fmt.Println("Agent Spoken __ ms after User Stopped", c.timeTTSStart.Sub(c.timeSTTEND))
	}(c.ctx)
	// time.Sleep(2 * time.Second)
	// }
}
