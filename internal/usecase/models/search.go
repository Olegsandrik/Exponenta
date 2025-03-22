package models

import "github.com/Olegsandrik/Exponenta/internal/delivery/dto"

type SearchResponseModel struct {
	Recipes []RecipeModel
}

type SuggestResponseModel struct {
	Suggestions []string
}

func ConvertSearchResponseToDto(searchResponse SearchResponseModel) dto.SearchResponseDto {
	return dto.SearchResponseDto{
		Recipes: ConvertRecipeToDto(searchResponse.Recipes),
	}
}

func ConvertSuggestResponseToDto(suggestResponse SuggestResponseModel) dto.SuggestResponseDto {
	return dto.SuggestResponseDto{
		Suggestions: suggestResponse.Suggestions,
	}
}
