package language_processor

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/sashabaranov/go-openai"
)

// ChatSession stores history for a user session
type ChatSession struct {
	messages []openai.ChatCompletionMessage
	mu       sync.Mutex
}

// Store user sessions in memory
var sessions = make(map[string]*ChatSession)
var sessionsMu sync.Mutex

// chatWithHistory processes a chat message with history
func ChatWithHistory(sessionID, prompt string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}

	// Retrieve or create session
	session := getSession(sessionID)

	// Append user message
	session.mu.Lock()
	session.messages = append(session.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})
	session.mu.Unlock()

	// Call OpenAI API
	client := openai.NewClient(apiKey)
	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    "gpt-4o",
		Messages: session.messages,
	})
	if err != nil {
		return "", err
	}

	// Get response and append to session
	aiResponse := resp.Choices[0].Message.Content
	session.mu.Lock()
	session.messages = append(session.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: aiResponse,
	})
	session.mu.Unlock()

	return aiResponse, nil
}

// getSession retrieves or initializes a chat session
func getSession(sessionID string) *ChatSession {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	if session, exists := sessions[sessionID]; exists {
		return session
	}
	session := &ChatSession{
		messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "You are a helpful assistant."},
		},
	}
	sessions[sessionID] = session
	return session
}
