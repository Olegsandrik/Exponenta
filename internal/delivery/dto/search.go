package dto

type SearchResponseDto struct {
	Recipes []RecipeDto `json:"recipes,omitempty"`
}

type SuggestResponseDto struct {
	Suggestions []string `json:"suggestions,omitempty"`
}
