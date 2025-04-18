package utils

import (
	"regexp"

	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
)

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

func SanitizeRecipeDescription(recipes []models.RecipeModel) {
	for i := range recipes {
		recipes[i].Desc = htmlTagRegex.ReplaceAllString(recipes[i].Desc, "")
	}
}
