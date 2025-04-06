package dto

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

type VoiceDto struct {
	Text string `json:"text"`
}

type DeepSeekAPIResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func GetVoiceData(r *http.Request) (VoiceDto, error) {
	var voice VoiceDto

	err := json.NewDecoder(r.Body).Decode(&voice)

	if err != nil {
		return VoiceDto{}, err
	}

	return voice, nil
}

func ConvertVoiceData(ctx context.Context, closer io.ReadCloser) (int, error) {
	var response DeepSeekAPIResp
	err := json.NewDecoder(closer).Decode(&response)
	if err != nil {
		return 0, err
	}

	id, err := strconv.Atoi(response.Choices[0].Message.Content)

	if err != nil {
		return 0, err
	}

	return id, nil
}
