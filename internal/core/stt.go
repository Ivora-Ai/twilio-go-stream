// THIS FILE IS DEPRECATED - Using callback approach instead of channels for Google STT
// (Twilio) --> μ-law (8kHz) --> [Decode Base64] --> [Convert μ-law to PCM] --> [Send PCM to Google STT] --> [Receive Text]
package core

import (
	"fmt"
	"twilio-go-stream/sdk/gcp"
	// "time" - removed unused import
)

// DEPRECATED: Routine - No longer used, replaced by callback approach in handleMediaStream
// We now set callbacks directly on the STT client instead of using channels
func (c *Client) Routine(sttClient *gcp.GoogleSTTClient) {
	fmt.Println("DEPRECATED: This function is no longer used. Using callback approach instead.")
	
	/* Original implementation kept for reference
	sttClient.StartStreamingAndAttach()
	fmt.Println("Listening... Speak now!")

	// Ensure transcription starts before sending audio
	go sttClient.Transcribe()
	time.Sleep(100 * time.Millisecond) // Allow STT setup time

	go sttClient.SendAudioInRealTime()

	go func() { //inform vad, adn streo stt data
		for {
			// fmt.Println("Running")
			sttResp := <-sttClient.Result
			fmt.Println("this be")
			for _, result := range sttResp.Results {
				if len(result.Alternatives) == 0 {
					fmt.Println("Len 0")

				}
				if result.IsFinal {
					c.timeSTTEND = time.Now().UTC()
					fmt.Println("This is final", result.Alternatives[0].Transcript)
					//todo callback
					dat := result.Alternatives[0].Transcript
					c.AgentResponse(true, c.wsConn, dat, c.streamID)
				}
				for _, alt := range result.Alternatives {
					fmt.Printf("\rTranscription: %s\n", alt.Transcript)
					c.UserMessage = append(c.UserMessage, alt.Transcript)
					// c.vad.UserSpeakChannel() <- true
					// c.vad.IsUserSpeaking(true)
				}
			}
			// this does not work, because each time this is called
			// fmt.Println("User is not speaking")
			// c.vad.UserSpeakChannel() <- false
		}
	}()
	*/
}