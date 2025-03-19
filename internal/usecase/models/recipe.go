package models

import (
	"encoding/json"
	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
)

type RecipeModel struct {
	Id          int
	Name        string
	Desc        string
	Img         string
	CookingTime int
	ServingsNum int
	Ingredients string
	Steps       string
}

type CurrentRecipeModel struct {
	Id          int
	Name        string
	CurrentStep CurrentStepRecipeModel
}

type CurrentStepRecipeModel struct {
	NumStep     int
	Step        string
	Ingredients json.RawMessage
	Equipment   json.RawMessage
	Length      json.RawMessage
}

func ConvertCurrentRecipeToDTO(recipe CurrentRecipeModel) dto.CurrentRecipeDto {
	return dto.CurrentRecipeDto{
		Id:          recipe.Id,
		Name:        recipe.Name,
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
		Id:          recipe.Id,
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
			Id:          r.Id,
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
			Id:          r.Id,
			Img:         r.Img,
			Desc:        r.Desc,
			Name:        r.Name,
			CookingTime: r.CookingTime,
			ServingsNum: r.ServingsNum,
			Steps:       json.RawMessage(r.Steps),
		})
	}
	return RecipeItems
}
