package models

import (
	"encoding/json"
	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
)

type RecipeModel struct {
	Name        string
	Desc        string
	Img         string
	CookingTime int
	ServingsNum int
	Ingredients string
	Steps       string
}

type Step struct {
	Number      int          `json:"number"`
	Step        string       `json:"step"`
	Ingredients []Ingredient `json:"ingredients"`
	Equipment   []Equipment  `json:"equipment"`
}

type Ingredient struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	LocalizedName string `json:"localizedName"`
	Image         string `json:"image"`
}

type Equipment struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	LocalizedName string `json:"localizedName"`
	Image         string `json:"image"`
}

func ConvertDtoToModel(rt []dto.RecipeDto) []RecipeModel {
	RecipeItems := make([]RecipeModel, 0, len(rt))
	for _, r := range rt {
		RecipeItems = append(RecipeItems, RecipeModel{
			Name:        r.Name,
			Desc:        r.Desc,
			Img:         r.Img,
			CookingTime: r.CookingTime,
			ServingsNum: r.ServingsNum,
			Ingredients: r.Ingredients,
			Steps:       string(r.Steps),
		})
	}
	return RecipeItems
}

func ConvertModelToDto(rm []RecipeModel) []dto.RecipeDto {
	RecipeItems := make([]dto.RecipeDto, 0, len(rm))
	for _, r := range rm {
		RecipeItems = append(RecipeItems, dto.RecipeDto{
			Img:         r.Img,
			Desc:        r.Desc,
			Name:        r.Name,
			CookingTime: r.CookingTime,
			ServingsNum: r.ServingsNum,
			Ingredients: r.Ingredients,
			Steps:       json.RawMessage(r.Steps),
		})
	}
	return RecipeItems
}

func ConvertOneModelToOneDto(rm RecipeModel) dto.RecipeDto {
	return dto.RecipeDto{
		Img:  rm.Img,
		Name: rm.Name,
		Desc: rm.Desc,
	}
}
