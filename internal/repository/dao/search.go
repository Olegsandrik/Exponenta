package dao

import (
	"encoding/json"

	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
)

type ResponseElasticSuggestIndex struct {
	Hits struct {
		Hits []struct {
			Source Suggest `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type Suggest struct {
	Name string `json:"name"`
}

type ResponseElasticRecipeIndex struct {
	Hits struct {
		Hits []struct {
			Source RecipeTable `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type DishTypesDao struct {
	DishTypes []string
}

type DietsDao struct {
	Diets string
}

type TimeDao struct {
	Min int `db:"min"`
	Max int `db:"max"`
}

func ConvertTimeDaoToModel(tm TimeDao) models.TimeModel {
	return models.TimeModel{
		Min: tm.Min,
		Max: tm.Max,
	}
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

func ConvertResponseElasticSuggestIndexToModel(resp ResponseElasticSuggestIndex) models.SuggestResponseModel {
	var result models.SuggestResponseModel

	for _, item := range resp.Hits.Hits {
		result.Suggestions = append(result.Suggestions, item.Source.Name)
	}

	return result
}

func MakeSet(items []json.RawMessage) (map[string]struct{}, error) {
	hashMapDiets := make(map[string]struct{})
	var currentRow []string

	for _, item := range items {
		err := json.Unmarshal(item, &currentRow)

		if err != nil {
			return nil, err
		}

		for idx := range currentRow {
			hashMapDiets[currentRow[idx]] = struct{}{}
		}
	}

	return hashMapDiets, nil
}
