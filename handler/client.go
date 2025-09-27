package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"twilio-go-stream/internal/core"
	"twilio-go-stream/sdk/deepgram"
	"twilio-go-stream/sdk/gcp"

	"github.com/gorilla/websocket"
)

type Corer interface {
	Talk(*websocket.Conn)
}
type Client struct {
	core        Corer
	PublicURL   string
	sttProvider string
	ttsProvider string
}

var wsConn *websocket.Conn

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// New creates a new Client with the specified providers
func New(publicUrl string, sttProvider string, ttsProvider string) *Client {
	return &Client{
		PublicURL:   publicUrl,
		sttProvider: sttProvider,
		ttsProvider: ttsProvider,
	}
}

// Must creates a client with default settings (deprecated, use New instead)
func Must(publicUrl string) *Client {
	return New(publicUrl, "deepgram", "deepgram")
}

func (c *Client) SetRoutes() {
	http.HandleFunc("/incoming-call", c.handleIncomingCall)
	http.HandleFunc("/media-stream", c.handleMediaStream)

}

// Handles incoming calls and returns TwiML response
func (c *Client) handleIncomingCall(w http.ResponseWriter, r *http.Request) {
	log.Println("Incoming call received!")
	Host := c.PublicURL
	// fmt.Println("Host:", Host)
	twiml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <Connect>
        <Stream track="inbound_track" url="wss://%s/media-stream">
            <Parameter name="agent_id" value="FNWYO5RqakGnPMcYXmua" />
        </Stream>
    </Connect>
</Response>`, Host)

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(twiml))
}

// WebSocket handler for Twilio's MediaStream
func (c *Client) handleMediaStream(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Variables for providers
	var gcpSTT *gcp.GoogleSTTClient
	var gcpTTS *gcp.GoogleTTSClient
	var deepgramTTS *deepgram.MyCallback
	var deepgramSTT *deepgram.DeepgramSTTCallback

	// Initialize STT based on provider setting
	switch c.sttProvider {
	case "gcp":
		log.Println("Setting up Google STT")
		var err error
		gcpSTT, err = gcp.NewGoogleSTTClient()
		if err != nil {
			log.Printf("Error initializing Google STT: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Start Google STT stream - only initialize, don't start Transcribe yet
		gcpSTT.StartStreamingAndAttach()
		// We'll start Transcribe after setting up the AgentResponse callback
		go gcpSTT.SendAudioInRealTime()

		defer gcpSTT.Close()
	case "deepgram":
		log.Println("Setting up Deepgram STT")
		deepgramSTT = deepgram.InitSTT()
		if deepgramSTT == nil {
			log.Println("Error initializing Deepgram STT")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		deepgramSTT.ConnectWS()
		defer deepgramSTT.Disconnect()
	default:
		log.Printf("Unknown STT provider: %s", c.sttProvider)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Initialize TTS based on provider setting
	switch c.ttsProvider {
	case "gcp":
		log.Println("Setting up Google TTS")
		var err error
		gcpTTS, err = gcp.NewGoogleTTSClient(ctx)
		if err != nil {
			log.Printf("Error initializing Google TTS: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer gcpTTS.Close()
	case "deepgram":
		log.Println("Setting up Deepgram TTS")
		deepgramTTS = deepgram.Init()
		if deepgramTTS == nil {
			log.Println("Error initializing Deepgram TTS")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		deepgramTTS.ConnectTTS()
		defer deepgramTTS.Disconnect()
	default:
		log.Printf("Unknown TTS provider: %s", c.ttsProvider)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create core client with the initialized providers
	coreClient := core.Must(gcpSTT, gcpTTS, deepgramTTS, deepgramSTT)
	c.core = coreClient
	stopChan := make(chan struct{})
	coreClient.Interrupt.Manager(stopChan)
	defer close(stopChan)

	// Now that we have the core client with AgentResponse, start Google STT transcription if needed
	if gcpSTT != nil {
		gcpSTT.Transcribe()
	}

	log.Println("New WebSocket connection established.")

	// Upgrade to WebSocket
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}

	//pass ws to core
	c.core.Talk(wsConn)
	defer wsConn.Close()
}
