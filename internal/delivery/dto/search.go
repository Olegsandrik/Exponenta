package dto

type SearchResponseDto struct {
	Recipes []RecipeDto `json:"recipes,omitempty"`
}
