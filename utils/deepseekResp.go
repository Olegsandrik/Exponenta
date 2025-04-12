package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	Prompt = `You are a helpful assistant, you need to recognize main idea of russian text and send me only a number.
	You should send me 1 if main idea of text is next step or switch step.
	You should send me 2 if main idea of text is previous step or switch step to previous.
	You should send me 3 if main idea of text is end cooking.
	You should send me 4 if main idea of text is end timer.
	You should send me 5 if main idea of text is start timer.
	You should send me 6 if main idea of text is get all timers.
	You should send me 0 on other ideas.
	`
	PromptGen = `You are a professional chef assistant. Provide detailed cooking recipes in Russian with 
	next json format. Send me only json od recipe.
	"name": "str",
	"description": "str",
	"servingsNum": int,
	"dishTypes": [
		"str",
		"str",
	],
	"diets": [
		"str",
		"str"
	],
	"steps": [
		{
			"number": int,
			"step": "description movement step",
			"ingredients": [
				{
					"name": "молотый эспрессо",
					"localizedName": "молотый эспрессо"
				},
				{
					"name": "взбитые сливки",
					"localizedName": "взбитые сливки"
					
				}
			],
			"equipment": [
				{
					"name": "пергамент для выпечки",
					"localizedName": "пергамент для выпечки",
				},
				{
					"name": "водяная баня",
					"localizedName": "водяная баня",
				}
			],
			"length": {
				"number": 5,
				"unit": "минут"
			}
		},
		{
			"number": int,
			"step": "description movement step",
			"ingredients": [
				{
					"name": "молотый эспрессо",
					"localizedName": "молотый эспрессо"
				},
				{
					"name": "взбитые сливки",
					"localizedName": "взбитые сливки"
				}
			],
			"equipment": [
				{
					"name": "пергамент для выпечки",
					"localizedName": "пергамент для выпечки",
				},
				{
					"name": "водяная баня",
					"localizedName": "водяная баня",
				}
			],
		}
	]
	You must use products, that i will send you`
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream"`
	MaxTokens int       `json:"max_tokens"`
}

func BuildRequest(ctx context.Context, userInput string, APIURL string, APIKey string) (*http.Request, error) {
	reqChat := ChatRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "system", Content: Prompt},
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

func BuildRequestGeneration(ctx context.Context, userInput string, APIURL string, APIKey string) (*http.Request, error) {
	reqChat := ChatRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "system", Content: PromptGen},
			{Role: "user", Content: userInput},
		},
		Stream:    false,
		MaxTokens: 1000,
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
