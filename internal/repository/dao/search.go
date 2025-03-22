package dao

import (
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
)

type ResponseElasticRecipeIndex struct {
	Hits struct {
		Hits []struct {
			Source RecipeTable `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func ConvertResponseElasticRecipeIndexToModel(resp ResponseElasticRecipeIndex) []models.RecipeModel {
	result := make([]models.RecipeModel, 0, len(resp.Hits.Hits))

	for _, item := range resp.Hits.Hits {
		result = append(result, models.RecipeModel{
			ID:   item.Source.ID,
			Name: item.Source.Name,
			Img:  item.Source.Img,
			Desc: item.Source.Desc,
		})
	}
	return result
}
