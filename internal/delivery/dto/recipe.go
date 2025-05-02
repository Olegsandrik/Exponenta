package dto

import (
	"encoding/json"
	"net/http"
)

type RecipeDto struct {
	ID              int             `json:"id,omitempty"`
	Version         int             `json:"version,omitempty"`
	Name            string          `json:"name,omitempty"`
	Desc            string          `json:"description,omitempty"`
	Img             string          `json:"img,omitempty"`
	CookingTime     int             `json:"cookingTimeMinutes,omitempty"`
	ServingsNum     int             `json:"servingsNum,omitempty"`
	DishTypes       json.RawMessage `json:"dishTypes,omitempty"`
	Diets           json.RawMessage `json:"diets,omitempty"`
	HealthScore     int             `json:"healthScore,omitempty"`
	Ingredients     json.RawMessage `json:"ingredients,omitempty"`
	Steps           json.RawMessage `json:"steps,omitempty"`
	Query           string          `json:"query,omitempty"`
	UserIngredients json.RawMessage `json:"userIngredients,omitempty"`
}

type CurrentRecipeDto struct {
	ID          int                  `json:"id,omitempty"`
	Name        string               `json:"name,omitempty"`
	TotalSteps  int                  `json:"totalSteps,omitempty"`
	CurrentStep CurrentStepRecipeDto `json:"currentStep,omitempty"`
}

type CurrentStepRecipeDto struct {
	NumStep     int             `json:"number,omitempty"`
	Step        string          `json:"step,omitempty"`
	Ingredients json.RawMessage `json:"ingredients,omitempty"`
	Equipment   json.RawMessage `json:"equipment,omitempty"`
	Length      json.RawMessage `json:"length,omitempty"`
}

type TimerRecipeDto struct {
	Length  json.RawMessage `json:"length,omitempty"`
	Step    string          `json:"step,omitempty"`
	StepNum int             `json:"stepNum,omitempty"`
}

type TimerRecipeDataDto struct {
	Time    int `json:"length"`
	StepNum int `json:"step"`
}

type GenerationRecipeDto struct {
	Query       string   `json:"query"`
	Ingredients []string `json:"ingredients"`
}

func GetCookingRecipeData(r *http.Request) (RecipeDto, error) {
	var recipe RecipeDto

	err := json.NewDecoder(r.Body).Decode(&recipe)

	if err != nil {
		return RecipeDto{}, err
	}

	return recipe, nil
}

func GetGenerationData(r *http.Request) (GenerationRecipeDto, error) {
	var generateDTO GenerationRecipeDto

	err := json.NewDecoder(r.Body).Decode(&generateDTO)

	if err != nil {
		return GenerationRecipeDto{}, err
	}

	return generateDTO, nil
}

func GetTimerRecipeData(r *http.Request) (TimerRecipeDataDto, error) {
	var timer TimerRecipeDataDto

	err := json.NewDecoder(r.Body).Decode(&timer)

	if err != nil {
		return TimerRecipeDataDto{}, err
	}

	return timer, nil
}
