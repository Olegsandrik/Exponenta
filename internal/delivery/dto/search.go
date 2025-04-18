package dto

type SearchResponseDto struct {
	Recipes []RecipeDto `json:"recipes,omitempty"`
}

type SuggestResponseDto struct {
	Suggestions []string `json:"suggestions,omitempty"`
}

type FiltersDto struct {
	Diets     []string `json:"diets"`
	DishTypes []string `json:"dishTypes"`
	Time      TimeDto  `json:"time"`
}

type TimeDto struct {
	Min int `json:"min"`
	Max int `json:"max"`
}
