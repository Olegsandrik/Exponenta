package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

func BuildRequest(ctx context.Context, userInput string, APIURL string, APIKey string,
	promptChoice string) (*http.Request, error) {
	reqChat := ChatRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "system", Content: promptChoice},
			{Role: "user", Content: userInput},
		},
		Stream: false,
	}

	reqChatBytes, err := json.Marshal(reqChat)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, APIURL, bytes.NewBuffer(reqChatBytes))

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", APIKey))

	return req, nil
}

func GetResponseData(ctx context.Context, q string, APIURL string, APIKey string, prompt string) (string, error) {
	req, err := BuildRequest(ctx, q, APIURL, APIKey, prompt)

	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return "", fmt.Errorf("status code %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	respData, err := dto.ConvertGenerationData(resp.Body)

	if err != nil {
		return "", err
	}

	respData = strings.Replace(respData, "```", "", 2)
	respData = strings.Replace(respData, "\n", "", -1)
	respData = strings.Replace(respData, "\t", "", -1)
	respData = strings.Replace(respData, "json", "", 1)

	return respData, nil
}
