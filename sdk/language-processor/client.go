package language_processor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"twilio-go-stream/domain"
)

type Client struct {
	URL       string
	UserID    string
	SessionID string
	AgentID   string
}

func Must(URL, UserID, SessionID, AgentID string) *Client {
	c := &Client{}
	c.URL = URL
	c.UserID = UserID
	c.SessionID = SessionID
	c.AgentID = AgentID
	return c
}

//write me function that will go get req to

func GetChatResponse(sessionID, prompt string) (string, error) {
	resp, err := ChatWithHistory(sessionID, prompt)
	return resp, err
}

func GetChatResponseFromGroq(prompt *domain.Prompt) string {
	start := time.Now().UTC()
	url := "https://api.groq.com/openai/v1/chat/completions"
	method := "POST"

	// Convert struct to JSON
	jsonData, err := json.MarshalIndent(prompt, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	reader := strings.NewReader(string(jsonData))
	// fmt.Println(string(jsonData))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, reader)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Add("Content-Type", "application/json")

	req.Header.Add("Authorization", "Bearer "+os.Getenv("GROQ_API_KEY"))

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	fmt.Println("Grok took", time.Now().UTC().Sub(start), "ms")
	return string(body)
}
