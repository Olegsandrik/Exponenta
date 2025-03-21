package dto

import "encoding/json"

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
