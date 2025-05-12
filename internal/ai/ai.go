package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Config struct {
	GeminiAPIKey string `json:"gemini_apiKey"`
}

type CommandLog struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Command   struct {
		Raw        string   `json:"raw"`
		Executable string   `json:"executable"`
		Arguments  []string `json:"arguments"`
		CWD        string   `json:"cwd"`
	} `json:"command"`
	Output struct {
		Stderr   string `json:"stderr"`
		ExitCode int    `json:"exitCode"`
		Error    string `json:"error,omitempty"`
	} `json:"output"`
	Metadata struct {
		User     string `json:"user"`
		Platform string `json:"platform"`
		Shell    string `json:"shell"`
	} `json:"metadata"`
}

type GeminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

const (
	instructionForHm = `
- As an intelligent assistant, interpret the user's intent accurately. Provide precise shell commands in response, based on your analysis of the user's input and any errors they encountered.
- Your goal is to assist the user by giving them only the correct command they need to execute, formatted without explanations or additional details. Assume the user has a minimal shell environment installed and respond with the exact command they should run.
- Be concise and efficient, responding with only the command.
- platform %s
- Be very smart
- Do not hallucinate
- **Note:** If you're unsure of the correct response, or prefer not to answer for any reason, reply only with the UUID: 3d8a19a704.
`

	instructionForHp = `
- You are a command-line assistant, helping users run commands in a shell environment. Analyze the user's input and determine the exact shell command they need to execute, assuming they have a basic installation.
- Respond solely with the unformatted command line instruction, omitting any explanations or extraneous text.
- Focus on providing precise commands, interpreting user input efficiently and accurately to meet their needs.
- platform %s
- Be very smart
- Do not hallucinate
- **Note:** If you're unsure of the correct response, or prefer not to answer for any reason, reply only with the UUID: 3d8a19a704.
`

	instructionForExplain = `
You are a smart command-line assistant. The question the user has asked is -> %s
Explain it to the user properly, focusing on command-line concepts. If you cannot explain something just respond with 3d8a19a704 and nothing else. The output will be passed to a terminal so keep it clean and use clear formatting.
`

	instructionForChat = `
You are a helpful assistant answering a user's question. Provide a concise, informative answer in 3-4 lines maximum.
Be accurate, to the point, and helpful.

The question is: %s

Remember to keep your answer to 3-4 lines maximum.
`
)

func getApiKey() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(homeDir, ".goterm.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", errors.New("API key not found")
	}

	fileData, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	var config Config
	if err := json.Unmarshal(fileData, &config); err != nil {
		return "", err
	}

	if config.GeminiAPIKey == "" {
		return "", errors.New("API key is empty")
	}

	return config.GeminiAPIKey, nil
}

func saveApiKey(apiKey string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".goterm.json")
	config := Config{GeminiAPIKey: apiKey}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func getLastCommandLog() (*CommandLog, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	errorFile := filepath.Join(homeDir, ".goterm_error")
	if _, err := os.Stat(errorFile); os.IsNotExist(err) {
		return nil, errors.New("no error log found")
	}

	data, err := os.ReadFile(errorFile)
	if err != nil {
		return nil, err
	}

	var logs []CommandLog
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil, err
	}

	if len(logs) == 0 {
		return nil, errors.New("error log is empty")
	}

	return &logs[len(logs)-1], nil
}

func GenerateCommandForHm(ctx context.Context) (string, error) {
	apiKey, err := getApiKey()
	if err != nil {
		return "", err
	}

	lastLog, err := getLastCommandLog()
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(instructionForHm, runtime.GOOS)
	lastLogJSON, err := json.Marshal(lastLog)
	if err != nil {
		return "", err
	}

	fullPrompt := prompt + "\n" + string(lastLogJSON)

	return callGeminiAPI(ctx, apiKey, fullPrompt)
}

func GenerateCommandForHp(ctx context.Context, query string) (string, error) {
	apiKey, err := getApiKey()
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(instructionForHp, runtime.GOOS)
	fullPrompt := prompt + "\n" + query

	return callGeminiAPI(ctx, apiKey, fullPrompt)
}

func ExplainCommand(ctx context.Context, query string) (string, error) {
	apiKey, err := getApiKey()
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(instructionForExplain, query)

	return callGeminiAPI(ctx, apiKey, prompt)
}

func ChatWithAI(ctx context.Context, question string) (string, error) {
	apiKey, err := getApiKey()
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(instructionForChat, question)

	return callGeminiAPI(ctx, apiKey, prompt)
}

func callGeminiAPI(ctx context.Context, apiKey string, prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1/models/gemini-1.5-flash:generateContent?key=%s", apiKey)

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
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response GeminiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "3d8a19a704", nil
	}

	responseText := response.Candidates[0].Content.Parts[0].Text
	command := strings.TrimSpace(strings.Split(responseText, "\n")[0])

	if command == "" || command == "3d8a19a704" {
		return "3d8a19a704", nil
	}

	return command, nil
}

func CheckAndSetupApiKey() (bool, error) {
	_, err := getApiKey()
	if err == nil {
		return true, nil
	}

	fmt.Println("⟨ ×︵× ⟩ API key not found.")
	fmt.Print("⟨ ◠︰◠ ⟩ Please enter your Gemini API key: ")

	var apiKey string
	fmt.Scanln(&apiKey)

	if apiKey == "" {
		return false, errors.New("API key cannot be empty")
	}

	err = saveApiKey(apiKey)
	if err != nil {
		return false, err
	}

	fmt.Println("⟨ ◠︶◠ ⟩ API key saved successfully!")
	return true, nil
}
