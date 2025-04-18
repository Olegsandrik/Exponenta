package models

import "github.com/Olegsandrik/Exponenta/internal/delivery/dto"

type SearchResponseModel struct {
	Recipes []RecipeModel
}

type SuggestResponseModel struct {
	Suggestions []string
}

type FiltersModel struct {
	Diets     []string
	DishTypes []string
}

type TimeModel struct {
	Min int
	Max int
}

func ConvertTimeModelToDto(time TimeModel) dto.TimeDto {
	return dto.TimeDto{
		Min: time.Min,
		Max: time.Max,
	}
}

func ConvertFilterModelToDto(dishTypes []string, diets []string, tm TimeModel) dto.FiltersDto {
	return dto.FiltersDto{
		Diets:     diets,
		DishTypes: dishTypes,
		Time:      ConvertTimeModelToDto(tm),
	}
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
