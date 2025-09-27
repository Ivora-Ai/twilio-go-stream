package gcp

type Client struct {
	SAPath     string
	Frequency  int
	AudioCodec string
}

func Must(SAPath, AudioCodec string, Frequency int) *Client {
	c := &Client{}
	c.SAPath = SAPath
	c.Frequency = Frequency
	c.AudioCodec = AudioCodec
	return c
}
