package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// OllamaRequest is the JSON body sent to the Ollama /api/generate endpoint.
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse is the JSON response from Ollama.
type OllamaResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// QueryRequest is the frontend-facing input for AI queries.
type QueryRequest struct {
	Role   string `json:"role"`   // "cto" or "zen"
	Prompt string `json:"prompt"` // raw user prompt (will be anonymized)
}

// QueryResponse is the result returned to the frontend.
type QueryResponse struct {
	Role     string `json:"role"`
	Response string `json:"response"`
	Model    string `json:"model"`
}

// Nudge represents an inter-context reminder.
type Nudge struct {
	Message string `json:"message"`
	Type    string `json:"type"` // "inter_context"
	TaskID  string `json:"task_id,omitempty"`
}

// ollamaBaseURL can be overridden by OLLAMA_HOST env variable.
func ollamaBaseURL() string {
	if host := os.Getenv("OLLAMA_HOST"); host != "" {
		return host
	}
	return "http://localhost:11434"
}

// getSystemPrompt returns the AI role-specific system instruction.
func getSystemPrompt(role string) string {
	switch role {
	case "cto":
		return `You are a senior CTO and engineering mentor. 
Give concise, practical advice on work tasks, technical decisions, and priorities.
Be direct, efficient, and action-oriented. Keep responses under 200 words.`
	case "zen":
		return `You are a Zen master and mindfulness coach.
Give calm, thoughtful guidance on personal goals, habits, and wellbeing.
Use gentle language, short sentences, occasional metaphors from nature.
Keep responses under 150 words.`
	}
	return "You are a helpful assistant. Be concise."
}

// aiModel returns the configured model name, defaulting to gemma2:2b.
func aiModel() string {
	if m := os.Getenv("OLLAMA_MODEL"); m != "" {
		return m
	}
	return "gemma2:2b"
}

// Query sends an anonymized prompt to Ollama and returns the AI response.
// The raw prompt is NEVER logged or stored.
func Query(role string, rawPrompt string) (*QueryResponse, error) {
	systemPrompt := getSystemPrompt(role)
	anonymized := AnonymizeText(rawPrompt)

	fullPrompt := fmt.Sprintf("[INST] %s\n\n%s [/INST]", systemPrompt, anonymized)

	payload := OllamaRequest{
		Model:  aiModel(),
		Prompt: fullPrompt,
		Stream: false,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Post(ollamaBaseURL()+"/api/generate", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read ollama response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned %d: %s", resp.StatusCode, string(respBody))
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return nil, fmt.Errorf("parse ollama response: %w", err)
	}

	log.Printf("🤖 AI query [%s]: %d chars → %d chars response", role, len(anonymized), len(ollamaResp.Response))

	return &QueryResponse{
		Role:     role,
		Response: ollamaResp.Response,
		Model:    ollamaResp.Model,
	}, nil
}

// Ping checks if Ollama is reachable.
func Ping() error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(ollamaBaseURL() + "/api/tags")
	if err != nil {
		return fmt.Errorf("ollama not reachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}
	return nil
}
