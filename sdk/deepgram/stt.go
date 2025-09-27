package deepgram

// Copyright 2023-2024 Deepgram SDK contributors. All Rights Reserved.
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// SPDX-License-Identifier: MIT

// streaming
import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	api "github.com/deepgram/deepgram-go-sdk/pkg/api/listen/v1/websocket/interfaces"
	interfaces "github.com/deepgram/deepgram-go-sdk/pkg/client/interfaces"
	client "github.com/deepgram/deepgram-go-sdk/pkg/client/listen"
	websocketv1 "github.com/deepgram/deepgram-go-sdk/pkg/client/listen/v1/websocket"
	"github.com/gorilla/websocket"
)

const secondToConsiderSpeechEnd = 2

type UserSaid struct {
	text          string
	when          time.Time
	said          bool
	isFinal       bool
	mu            sync.Mutex
	sb            *strings.Builder
	AgentResponse func(bool, string)
	UserSpeaking  func(bool)
}

func (u *UserSaid) Say(text string, isFinal bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.text = text
	u.said = true
	u.isFinal = isFinal
	u.when = time.Now().UTC()
}
func (u *UserSaid) MonitorAndTrigger() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	largeSentence := ""
	// pastSaid := ""
	for {
		select {
		// case ctx.CancelFunc://add this later
		// 	fmt.Println("Closing routine MonitorAndTrigger")
		// 	return
		case <-ticker.C:
			u.mu.Lock()
			if u.isFinal {
				largeSentence += ". " + u.text
				u.isFinal = false
			}
			// if pastSaid = u.said{
			// 	u.said=false
			// }
			if time.Since(u.when) > secondToConsiderSpeechEnd*time.Second && u.said {
				largeSentence += "" + u.text
				u.said = false
				if u.sb != nil {
					u.sb.Reset()
					// fmt.Println("Clearned")
				}
				u.UserSpeaking(false)
				u.AgentResponse(true, largeSentence)
				fmt.Println("Triggggggggger broo mateyyyys", time.Now().UTC())
				fmt.Println("--------------------------------------------", largeSentence)
				largeSentence = ""
			}
			u.mu.Unlock()
		}
	}
}

// Implement your own callback
type DeepgramSTTCallback struct {
	sb       *strings.Builder
	dgClient *websocketv1.WSCallback
	Sid      string
	WsConn   *websocket.Conn
	UserSaid *UserSaid
}

func (c DeepgramSTTCallback) ConnectWS() {
	bConnected := c.dgClient.Connect()
	if !bConnected {
		fmt.Println("Client.Connect failed")
		os.Exit(1)
	}
}
func (c DeepgramSTTCallback) Disconnect() {
	// close DG client
	defer c.dgClient.Stop()
}

func (c DeepgramSTTCallback) Message(mr *api.MessageResponse) error {
	// handle the message
	sentence := strings.TrimSpace(mr.Channel.Alternatives[0].Transcript)

	if len(mr.Channel.Alternatives) == 0 || len(sentence) == 0 {
		return nil
	}

	if mr.IsFinal {
		c.sb.WriteString(sentence)
		c.sb.WriteString(" ")

		if mr.SpeechFinal {
			c.UserSaid.Say(c.sb.String(), true)

			// c.AgentResponse(true, c.WsConn, c.sb.String(), c.Sid)
			fmt.Printf("[------- Is Final]: %s %s\n", time.Now().UTC(), c.sb.String())
			c.sb.Reset()
		}
	} else {
		c.UserSaid.Say(c.sb.String(), false)
		c.UserSaid.UserSpeaking(true)
		fmt.Printf("[Interm Result]: %s %s\n", time.Now().UTC(), sentence)
	}

	return nil
}

func (c DeepgramSTTCallback) Open(ocr *api.OpenResponse) error {
	// handle the open
	fmt.Printf("\n[Open] Received\n")
	return nil
}

func (c DeepgramSTTCallback) Metadata(md *api.MetadataResponse) error {
	// handle the metadata
	fmt.Printf("\n[Metadata] Received\n")
	fmt.Printf("Metadata.RequestID: %s\n", strings.TrimSpace(md.RequestID))
	fmt.Printf("Metadata.Channels: %d\n", md.Channels)
	fmt.Printf("Metadata.Created: %s\n\n", strings.TrimSpace(md.Created))
	return nil
}

func (c DeepgramSTTCallback) SpeechStarted(ssr *api.SpeechStartedResponse) error {
	// fmt.Printf("\n[SpeechStarted] Received %s\n", time.Now().UTC())
	return nil
}

func (c DeepgramSTTCallback) UtteranceEnd(ur *api.UtteranceEndResponse) error {
	utterance := strings.TrimSpace(c.sb.String())
	if len(utterance) > 0 {
		c.UserSaid.Say(c.sb.String(), true)
		// c.AgentResponse(true, c.WsConn, c.sb.String(), c.Sid)
		fmt.Printf("[------- UtteranceEnd]: %s\n %s", time.Now().UTC(), utterance)
		c.sb.Reset()
	} else {
		fmt.Printf("\n[UtteranceEnd] Received %s\n", time.Now().UTC())
	}

	return nil
}

func (c DeepgramSTTCallback) Close(ocr *api.CloseResponse) error {
	// handle the close
	fmt.Printf("\n[Close] Received\n")
	return nil
}

func (c DeepgramSTTCallback) Error(er *api.ErrorResponse) error {
	// handle the error
	fmt.Printf("\n[Error] Received\n")
	fmt.Printf("Error.Type: %s\n", er.Type)
	fmt.Printf("Error.ErrCode: %s\n", er.ErrCode)
	fmt.Printf("Error.Description: %s\n\n", er.Description)
	return nil
}

func (c DeepgramSTTCallback) UnhandledEvent(byData []byte) error {
	// handle the unhandled event
	fmt.Printf("\n[UnhandledEvent] Received\n")
	fmt.Printf("UnhandledEvent: %s\n\n", string(byData))
	return nil
}

func (c DeepgramSTTCallback) PushAudioByte(audio []byte) {
	c.dgClient.Write(audio)
}

func InitSTT() *DeepgramSTTCallback {

	/*
		DG Streaming API
	*/
	// init library
	// client.Init(client.InitLib{
	// 	LogLevel: client.LogLevelDefault, // LogLevelDefault, LogLevelFull, LogLevelDebug, LogLevelTrace
	// })

	// Go context
	ctx := context.Background()

	// client options
	cOptions := &interfaces.ClientOptions{
		EnableKeepAlive: true,
	}

	// set the Transcription options
	tOptions := &interfaces.LiveTranscriptionOptions{
		Model: "nova-2", //nova-3
		// Keyterm:     []string{"deepgram"},
		Language: "hi",

		Punctuate:   true,
		Encoding:    "mulaw",
		Channels:    1,
		SampleRate:  8000,
		SmartFormat: true,
		VadEvents:   true,
		// Endpointing: "500",
		FillerWords: true,
		Diarize:     true,
		Dictation:   true,

		// To get UtteranceEnd, the following must be set:
		InterimResults: true,
		UtteranceEndMs: "1000",
	}

	// example on how to send a custom parameter
	// params := make(map[string][]string, 0)
	// params["dictation"] = []string{"true"}
	// ctx = interfaces.WithCustomParameters(ctx, params)
	dc := &DeepgramSTTCallback{
		sb: &strings.Builder{},
	}
	// Get Deepgram API key from environment
	apiKey := os.Getenv("DEEPGRAM_API_KEY")
	if apiKey == "" {
		fmt.Println("ERROR: DEEPGRAM_API_KEY environment variable not set")
		return nil
	}

	// Create a Deepgram client
	dgClient, err := client.NewWSUsingCallback(ctx, apiKey, cOptions, tOptions, dc)
	if err != nil {
		fmt.Println("ERROR creating LiveTranscription connection:", err)
		return nil
	}
	dc.dgClient = dgClient
	dc.UserSaid = &UserSaid{}
	go dc.UserSaid.MonitorAndTrigger()
	// implement your own callback
	return dc
}
