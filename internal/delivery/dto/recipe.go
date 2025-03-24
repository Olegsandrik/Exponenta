package dto

import (
	"encoding/json"
	"net/http"
)

type RecipeDto struct {
	ID          int             `json:"id,omitempty"`
	Name        string          `json:"name,omitempty"`
	Desc        string          `json:"description,omitempty"`
	Img         string          `json:"img,omitempty"`
	CookingTime int             `json:"cookingTime,omitempty"`
	ServingsNum int             `json:"servingsNum,omitempty"`
	Ingredients json.RawMessage `json:"ingredients,omitempty"`
	Steps       json.RawMessage `json:"steps,omitempty"`
}

type CurrentRecipeDto struct {
	ID          int                  `json:"id,omitempty"`
	Name        string               `json:"name,omitempty"`
	CurrentStep CurrentStepRecipeDto `json:"currentStep,omitempty"`
}

type CurrentStepRecipeDto struct {
	NumStep     int             `json:"number,omitempty"`
	Step        string          `json:"description,omitempty"`
	Ingredients json.RawMessage `json:"ingredients,omitempty"`
	Equipment   json.RawMessage `json:"equipment,omitempty"`
	Length      json.RawMessage `json:"time,omitempty"`
}

type TimerRecipeDto struct {
	Length  json.RawMessage `json:"time,omitempty"`
	Step    string          `json:"description,omitempty"`
	StepNum int             `json:"stepNum,omitempty"`
}

type TimerRecipeDataDto struct {
	Time    int `json:"time"`
	StepNum int `json:"step"`
}

func GetCookingRecipeData(r *http.Request) (RecipeDto, error) {
	var recipe RecipeDto

	err := json.NewDecoder(r.Body).Decode(&recipe)

	if err != nil {
		return RecipeDto{}, err
	}

	return recipe, nil
}

func GetTimerRecipeData(r *http.Request) (TimerRecipeDataDto, error) {
	var timer TimerRecipeDataDto

	err := json.NewDecoder(r.Body).Decode(&timer)

	if err != nil {
		return TimerRecipeDataDto{}, err
	}

	return timer, nil
}
