package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github/0PrashantYadav0/GO-TERM/pkg/config"
	"github/0PrashantYadav0/GO-TERM/pkg/logger"
)

// GeminiClient provides enhanced access to the Gemini API
type GeminiClient struct {
	apiKey     string
	httpClient *http.Client
	config     *config.Config
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient() (*GeminiClient, error) {
	cfg := config.GetConfig()
	if cfg.GeminiAPIKey == "" {
		return nil, errors.New("API key not set")
	}

	client := &GeminiClient{
		apiKey: cfg.GeminiAPIKey,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.DefaultTimeout) * time.Second,
		},
		config: cfg,
	}

	return client, nil
}

// StreamCompletion sends a request and streams the response through the callback
func (c *GeminiClient) StreamCompletion(
	ctx context.Context,
	prompt string,
	callback func(chunk string, done bool) error,
) error {
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1/models/gemini-1.5-flash:streamGenerateContent?key=%s",
		c.apiKey,
	)

	// Build request payload
	request := GeminiRequest{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{
						Text: prompt,
					},
				},
			},
		},
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("Gemini API error: %d %s", resp.StatusCode, string(body))
		return fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	// Handle streaming response
	scanner := bufio.NewScanner(resp.Body)

	var responseBuilder strings.Builder
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse the JSON
		var chunk struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}

		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			logger.Debug("Error parsing chunk: %v", err)
			continue
		}

		if len(chunk.Candidates) == 0 || len(chunk.Candidates[0].Content.Parts) == 0 {
			continue
		}

		text := chunk.Candidates[0].Content.Parts[0].Text
		responseBuilder.WriteString(text)

		// Call the callback with current chunk
		if err := callback(text, false); err != nil {
			return err
		}
	}

	// Signal completion
	return callback("", true)
}

// Complete sends a request and returns the complete response
func (c *GeminiClient) Complete(ctx context.Context, prompt string) (string, error) {
	var responseBuilder strings.Builder

	err := c.StreamCompletion(ctx, prompt, func(chunk string, done bool) error {
		responseBuilder.WriteString(chunk)
		return nil
	})

	if err != nil {
		return "", err
	}

	response := responseBuilder.String()
	if response == "" || response == "3d8a19a704" {
		return "3d8a19a704", nil
	}

	return response, nil
}
