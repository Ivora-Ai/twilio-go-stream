package domain

// Struct representing the JSON structure
type ChatCompletion struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
	SystemFingerprint string   `json:"system_fingerprint"`
	XGroq             XGroq    `json:"x_groq"`
}

type Choice struct {
	Index        int         `json:"index"`
	Message      Message     `json:"message"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Usage struct {
	QueueTime        float64 `json:"queue_time"`
	PromptTokens     int     `json:"prompt_tokens"`
	PromptTime       float64 `json:"prompt_time"`
	CompletionTokens int     `json:"completion_tokens"`
	CompletionTime   float64 `json:"completion_time"`
	TotalTokens      int     `json:"total_tokens"`
	TotalTime        float64 `json:"total_time"`
} // Message represents an individual message in the conversation.
// Request represents the overall request structure.
type Prompt struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type XGroq struct {
	ID string `json:"id"`
}

func InitPrompt() *Prompt {
	return &Prompt{
		Model: "llama-3.3-70b-versatile",
		Messages: []Message{
			{
				Role:    "system",
				Content: `You are Ivora, AI based speaking agent, you should talk like real human so assume everything needed to talk like a human and also your response size as you are speaking on a call, you should response in limited words. all of your responses should be in same language and it should sound well, if language of response is hindi then use '.' fullstop to break sentences . answeres should not exceed 500 character, and answers should me more than 200 chars. If the user wants to disconnect, respond with 'close()'`,
			},
		},
	}
}

func (p *Prompt) PushMessage(role, message string) {
	p.Messages = append(p.Messages, Message{Role: role, Content: message})
}
