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

const (
	PromptVoice = `You are a helpful assistant, you need to recognize main idea of russian text and send me only a number.
	You should send me 1 if main idea of text is next step or switch step.
	You should send me 2 if main idea of text is previous step or switch step to previous.
	You should send me 3 if main idea of text is end cooking.
	You should send me 4 if main idea of text is end timer.
	You should send me 5 if main idea of text is start timer.
	You should send me 6 if main idea of text is get all timers.
	You should send me 0 on other ideas.`

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
	"ingredients": [
		{
			"id": 1, // 1... inf
			"name": "str",
			"image": "str",
			"amount": 0.5,
			"unit": "ч. л."
		},
		{
			"id": 2,
			"name": "chocolate",
			"image": "milk-chocolate.jpg",
			"amount": 8,
			"unit": "унций"
		}
	],
	"totalSteps": int, 
	"readyInMinutes": int,
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

	PromptMod = `You are a professional chef assistant. I will send you my recipe and you should reform
	my recipe with my promise. Send me only json of my new recipe.`
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

type Prompts struct {
	Prompts map[string]string
}

func NewPrompts() *Prompts {
	promptMap := make(map[string]string)

	promptMap["Voice"] = PromptVoice
	promptMap["Gen"] = PromptGen
	promptMap["Mod"] = PromptMod

	return &Prompts{
		Prompts: promptMap,
	}
}

func BuildRequest(ctx context.Context, userInput string, APIURL string, APIKey string,
	promptChoice string) (*http.Request, error) {
	Prompt := NewPrompts()

	reqChat := ChatRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "system", Content: Prompt.Prompts[promptChoice]},
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
