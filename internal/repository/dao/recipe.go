package dao

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
)

type RecipeTable struct {
	ID          int    `db:"id" json:"id,omitempty"`
	Name        string `db:"name" json:"name,omitempty"`
	Desc        string `db:"description" json:"description,omitempty"`
	Img         string `db:"image" json:"image,omitempty"`
	CookingTime int    `db:"ready_in_minutes" json:"cookingTime,omitempty"`
	ServingsNum int    `db:"servings" json:"servingsNum,omitempty"`
	Steps       string `db:"steps" json:"steps,omitempty"`
}

type CurrentRecipeTable struct {
	ID          int             `db:"recipe_id"`
	Name        string          `db:"name"`
	NumStep     int             `db:"step_num"`
	Step        string          `db:"step"`
	Ingredients json.RawMessage `db:"ingredients"`
	Equipment   json.RawMessage `db:"equipment"`
	Length      json.RawMessage `db:"length"`
}

type CurrentRecipeStepTable struct {
	NumStep     int             `json:"number" db:"step_num"`
	Step        string          `json:"step" db:"step"`
	Ingredients json.RawMessage `json:"ingredients"`
	Equipment   json.RawMessage `json:"equipment" db:"equipment"`
	Length      json.RawMessage `json:"length" db:"length"`
}

type TimerTable struct {
	StepNum     int       `db:"step_num"`
	Description string    `db:"description"`
	EndTime     time.Time `db:"end_time"`
}

type LengthTimer struct {
	Number int    `json:"number"`
	Unit   string `json:"unit"`
}

func ConvertTimerToDAO(tt []TimerTable) ([]models.TimerRecipeModel, error) {
	timers := make([]models.TimerRecipeModel, len(tt))
	for i, timer := range tt {
		diff := int(math.Round(time.Until(timer.EndTime).Seconds()))

		if diff < 0 {
			diff = 0
		}

		length := LengthTimer{
			Unit:   "seconds",
			Number: diff,
		}

		jsonLength, err := json.Marshal(length)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal length: %w", err)
		}

		timers[i] = models.TimerRecipeModel{
			Step:    timer.Description,
			Length:  jsonLength,
			StepNum: timer.StepNum,
		}
	}
	return timers, nil
}

func ConvertDaoToCurrentRecipe(cr CurrentRecipeTable) models.CurrentRecipeModel {
	return models.CurrentRecipeModel{
		ID:   cr.ID,
		Name: cr.Name,
		CurrentStep: models.CurrentStepRecipeModel{
			NumStep:     cr.NumStep,
			Step:        cr.Step,
			Ingredients: cr.Ingredients,
			Equipment:   cr.Equipment,
			Length:      cr.Length},
	}
}

func ConvertStringToSQLNullString(s sql.NullString) string {
	var length string
	if s.Valid {
		length = s.String
	} else {
		length = "null"
	}
	return length
}

func ConvertSQLNullStringToString(s string) sql.NullString {
	if s == "" || s == "NULL" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func ConvertDaoToCurrentStepRecipe(cs CurrentRecipeStepTable) models.CurrentStepRecipeModel {
	return models.CurrentStepRecipeModel{
		NumStep:     cs.NumStep,
		Step:        cs.Step,
		Ingredients: cs.Ingredients,
		Equipment:   cs.Equipment,
		Length:      cs.Length,
	}
}

func ConvertCurrentStepRecipeToDAO(cs models.CurrentStepRecipeModel) CurrentRecipeStepTable {
	return CurrentRecipeStepTable{
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
			ID:          r.ID,
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
			ID:   r.ID,
			Name: r.Name,
			Desc: r.Desc,
			Img:  r.Img,
		})
	}
	return RecipeItems
}
