package models

import (
	"encoding/json"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
)

type RecipeModel struct {
	ID          int
	Name        string
	Desc        string
	Img         string
	CookingTime int
	ServingsNum int
	DishTypes   string
	Diets       string
	Ingredients []byte
	HealthScore int
	Steps       string
}

type CurrentRecipeModel struct {
	ID          int
	Name        string
	TotalSteps  int
	CurrentStep CurrentStepRecipeModel
}

type CurrentStepRecipeModel struct {
	NumStep     int
	Step        string
	Ingredients json.RawMessage
	Equipment   json.RawMessage
	Length      json.RawMessage
}

type TimerRecipeModel struct {
	Length  json.RawMessage
	Step    string
	StepNum int
}

func ConvertTimersToDTO(steps []TimerRecipeModel) []dto.TimerRecipeDto {
	StepItems := make([]dto.TimerRecipeDto, len(steps))
	for i, step := range steps {
		StepItems[i] = dto.TimerRecipeDto{
			Length:  step.Length,
			Step:    step.Step,
			StepNum: step.StepNum,
		}
	}
	return StepItems
}

func ConvertCurrentRecipeToDTO(recipe CurrentRecipeModel) dto.CurrentRecipeDto {
	return dto.CurrentRecipeDto{
		ID:          recipe.ID,
		Name:        recipe.Name,
		TotalSteps:  recipe.TotalSteps,
		CurrentStep: ConvertCurrentStepToDTO(recipe.CurrentStep),
	}
}

func ConvertCurrentStepToDTO(step CurrentStepRecipeModel) dto.CurrentStepRecipeDto {
	return dto.CurrentStepRecipeDto{
		NumStep:     step.NumStep,
		Step:        step.Step,
		Ingredients: step.Ingredients,
		Equipment:   step.Equipment,
		Length:      step.Length,
	}
}

func ConvertDTOToCurrentRecipe(recipe dto.CurrentRecipeDto) CurrentRecipeModel {
	return CurrentRecipeModel{
		ID:          recipe.ID,
		Name:        recipe.Name,
		CurrentStep: ConvertDtoToCurrentStep(recipe.CurrentStep),
	}
}

func ConvertDtoToCurrentStep(step dto.CurrentStepRecipeDto) CurrentStepRecipeModel {
	return CurrentStepRecipeModel{
		NumStep:     step.NumStep,
		Step:        step.Step,
		Ingredients: step.Ingredients,
		Equipment:   step.Equipment,
		Length:      step.Length,
	}
}

func ConvertDtoToRecipe(rt []dto.RecipeDto) []RecipeModel {
	RecipeItems := make([]RecipeModel, 0, len(rt))
	for _, r := range rt {
		RecipeItems = append(RecipeItems, RecipeModel{
			ID:          r.ID,
			Name:        r.Name,
			Desc:        r.Desc,
			Img:         r.Img,
			CookingTime: r.CookingTime,
			ServingsNum: r.ServingsNum,
			Steps:       string(r.Steps),
		})
	}
	return RecipeItems
}

func ConvertRecipeToDto(rm []RecipeModel) []dto.RecipeDto {
	RecipeItems := make([]dto.RecipeDto, 0, len(rm))
	for _, r := range rm {
		RecipeItems = append(RecipeItems, dto.RecipeDto{
			ID:          r.ID,
			Img:         r.Img,
			Desc:        r.Desc,
			Name:        r.Name,
			CookingTime: r.CookingTime,
			ServingsNum: r.ServingsNum,
			Steps:       json.RawMessage(r.Steps),
			Diets:       json.RawMessage(r.Diets),
			DishTypes:   json.RawMessage(r.DishTypes),
			HealthScore: r.HealthScore,
			Ingredients: r.Ingredients,
		})
	}
	return RecipeItems
}
