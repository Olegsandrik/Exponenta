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

type CurrentRecipe struct {
	Id          int
	Name        string
	CurrentStep CurrentStepRecipe
}

type CurrentStepRecipe struct {
	NumStep     int
	Step        string
	Ingredients json.RawMessage
	Equipment   json.RawMessage
	Length      json.RawMessage
}

func ConvertCurrentRecipeToDTO(recipe CurrentRecipe) dto.CurrentRecipeDto {
	return dto.CurrentRecipeDto{
		Id:   recipe.Id,
		Name: recipe.Name,
		CurrentStep: dto.CurrentStepRecipeDto{
			NumStep:     recipe.CurrentStep.NumStep,
			Step:        recipe.CurrentStep.Step,
			Ingredients: recipe.CurrentStep.Ingredients,
			Equipment:   recipe.CurrentStep.Equipment,
			Length:      recipe.CurrentStep.Length,
		},
	}
}

func ConvertDTOToCurrentRecipe(recipe dto.CurrentRecipeDto) CurrentRecipe {
	return CurrentRecipe{
		Id:   recipe.Id,
		Name: recipe.Name,
		CurrentStep: CurrentStepRecipe{
			NumStep:     recipe.CurrentStep.NumStep,
			Step:        recipe.CurrentStep.Step,
			Ingredients: recipe.CurrentStep.Ingredients,
			Equipment:   recipe.CurrentStep.Equipment,
			Length:      recipe.CurrentStep.Length,
		},
	}
}

func ConvertDtoToModel(rt []dto.RecipeDto) []RecipeModel {
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

func ConvertModelToDto(rm []RecipeModel) []dto.RecipeDto {
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
