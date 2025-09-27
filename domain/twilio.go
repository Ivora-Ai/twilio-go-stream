package domain

// MediaStream represents WebSocket messages
type MediaStream struct {
	Event     string `json:"event"`
	StreamSid string `json:"streamSid"`
	Media     struct {
		Track   string `json:"track"`
		Payload string `json:"payload"`
	} `json:"media"`
}
