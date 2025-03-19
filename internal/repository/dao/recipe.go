package dao

import (
	"database/sql"
	"encoding/json"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
)

type RecipeTable struct {
	Id          int    `db:"id"`
	Name        string `db:"name"`
	Desc        string `db:"description"`
	Img         string `db:"image"`
	CookingTime int    `db:"ready_in_minutes"`
	ServingsNum int    `db:"servings"`
	Steps       string `db:"steps"`
}

type CurrentRecipeTable struct {
	Id          int             `db:"recipe_id"`
	Name        string          `db:"name"`
	NumStep     int             `db:"step_num"`
	Step        string          `db:"step"`
	Ingredients json.RawMessage `db:"ingredients"`
	Equipment   json.RawMessage `db:"equipment"`
	Length      json.RawMessage `db:"length"`
}

type CurrentStepRecipeTable struct {
	NumStep     int             `json:"number" db:"step_num"`
	Step        string          `json:"step" db:"step"`
	Ingredients json.RawMessage `json:"ingredients"`
	Equipment   json.RawMessage `json:"equipment" db:"equipment"`
	Length      json.RawMessage `json:"length" db:"length"`
}

func ConvertDaoToCurrentRecipe(cr CurrentRecipeTable) models.CurrentRecipeModel {
	return models.CurrentRecipeModel{
		Id:   cr.Id,
		Name: cr.Name,
		CurrentStep: models.CurrentStepRecipeModel{
			NumStep:     cr.NumStep,
			Step:        cr.Step,
			Ingredients: cr.Ingredients,
			Equipment:   cr.Equipment,
			Length:      cr.Length},
	}
}

func ConvertStringToSqlNullString(s sql.NullString) string {
	var length string
	if s.Valid {
		length = s.String
	} else {
		length = "null"
	}
	return length
}

func ConvertSqlNullStringToString(s string) sql.NullString {
	if s == "" || s == "NULL" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func ConvertDaoToCurrentStepRecipe(cs CurrentStepRecipeTable) models.CurrentStepRecipeModel {
	return models.CurrentStepRecipeModel{
		NumStep:     cs.NumStep,
		Step:        cs.Step,
		Ingredients: cs.Ingredients,
		Equipment:   cs.Equipment,
		Length:      cs.Length,
	}
}

func ConvertCurrentStepRecipeToDAO(cs models.CurrentStepRecipeModel) CurrentStepRecipeTable {
	return CurrentStepRecipeTable{
		NumStep:     cs.NumStep,
		Step:        cs.Step,
		Ingredients: cs.Ingredients,
		Equipment:   cs.Equipment,
		Length:      cs.Length,
	}
}

func ConvertDaoToRecipe(rt []RecipeTable) []models.RecipeModel {
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
