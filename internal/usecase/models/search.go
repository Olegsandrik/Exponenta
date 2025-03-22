package models

import "github.com/Olegsandrik/Exponenta/internal/delivery/dto"

type SearchResponseModel struct {
	Recipes []RecipeModel
}

func ConvertSearchResponseToDto(searchResponse SearchResponseModel) dto.SearchResponseDto {
	return dto.SearchResponseDto{
		Recipes: ConvertRecipeToDto(searchResponse.Recipes),
	}
}
