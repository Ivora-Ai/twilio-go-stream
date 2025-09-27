package twilio

type Client struct {
	TwilioURL  string
	WebHookURL string
	Secret     string
	Frequency  int
	AudioCodec string
	Greeting   string
}

func Must(TwilioURL, WebHookURL, Secret, AudioCodec, Greeting string, Frequency int) *Client {
	c := &Client{}
	c.TwilioURL = TwilioURL
	c.WebHookURL = WebHookURL
	c.Secret = Secret
	c.Frequency = Frequency
	c.AudioCodec = AudioCodec
	c.Greeting = Greeting
	return c
}
