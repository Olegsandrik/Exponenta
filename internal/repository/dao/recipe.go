package dao

import (
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
)

type RecipeTable struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
	Desc string `db:"description"`
	Img  string `db:"image"`
	// HealthScore string `db:"healt_score"`
	CookingTime int `db:"ready_in_minutes"`
	ServingsNum int `db:"servings"`
	// Ingredients string `db:"ingredients"`
	Steps string `db:"steps"`
}

func ConvertDaoToModel(rt []RecipeTable) []models.RecipeModel {
	RecipeItems := make([]models.RecipeModel, 0, len(rt))
	for _, r := range rt {
		RecipeItems = append(RecipeItems, models.RecipeModel{
			Id:          r.Id,
			Name:        r.Name,
			Desc:        r.Desc,
			Img:         r.Img,
			CookingTime: r.CookingTime,
			ServingsNum: r.ServingsNum,
			// Ingredients: r.Ingredients,
			Steps: r.Steps,
		})
	}
	return RecipeItems
}

func ConvertOneDaoToOneModel(rt RecipeTable) models.RecipeModel {
	return models.RecipeModel{
		Name:        rt.Name,
		Desc:        rt.Desc,
		Img:         rt.Img,
		CookingTime: rt.CookingTime,
		ServingsNum: rt.ServingsNum,
		// Ingredients: rt.Ingredients,
		Steps: rt.Steps,
	}
}

func ConvertModelToDao(rm []models.RecipeModel) []RecipeTable {
	RecipeItems := make([]RecipeTable, 0, len(rm))
	for _, r := range rm {
		RecipeItems = append(RecipeItems, RecipeTable{
			Id:   r.Id,
			Name: r.Name,
			Desc: r.Desc,
			Img:  r.Img,
		})
	}
	return RecipeItems
}
