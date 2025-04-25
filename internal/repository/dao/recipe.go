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
	ID          int             `db:"id" json:"id,omitempty"`
	Name        string          `db:"name" json:"name,omitempty"`
	Desc        string          `db:"description" json:"description,omitempty"`
	Img         string          `db:"image" json:"image,omitempty"`
	CookingTime int             `db:"ready_in_minutes" json:"cookingTime,omitempty"`
	ServingsNum int             `db:"servings" json:"servingsNum,omitempty"`
	Steps       string          `db:"steps" json:"steps,omitempty"`
	DishTypes   json.RawMessage `db:"dish_types" json:"dishTypes,omitempty"`
	Diets       json.RawMessage `db:"diets" json:"diets,omitempty"`
	HealthScore int             `db:"healthscore" json:"healthscore,omitempty"`
	TotalSteps  int             `db:"total_steps" json:"totalSteps,omitempty"`
	Ingredients json.RawMessage `db:"ingredients" json:"ingredients,omitempty"`
	Version     int             `db:"version" json:"version,omitempty"`
	Query       string          `db:"query" json:"query,omitempty"`
}

type CurrentRecipeTable struct {
	ID          int             `db:"recipe_id"`
	Name        string          `db:"name"`
	NumStep     int             `db:"step_num"`
	Step        string          `db:"step"`
	TotalSteps  int             `db:"total_steps"`
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

type IngredientTable struct {
	IngredientID int     `db:"ingredient_id" json:"id"`
	Name         string  `db:"name" json:"name"`
	Image        string  `db:"image" json:"image"`
	Amount       float64 `db:"amount" json:"amount"`
	Unit         string  `db:"unit" json:"unit"`
}

type GeneratedRecipe struct {
	Version        int `json:"version"`
	ID             int
	Name           string          `db:"name" json:"name"`
	Desc           string          `db:"description" json:"description"`
	ServingsNum    int             `db:"servings" json:"servingsNum"`
	TotalSteps     int             `db:"total_steps" json:"totalSteps"`
	ReadyInMinutes int             `db:"ready_in_minutes" json:"readyInMinutes"`
	Ingredients    json.RawMessage `db:"ingredients" json:"ingredients"`
	Steps          json.RawMessage `db:"steps" json:"steps"`
	DishTypes      json.RawMessage `db:"dish_types" json:"dishTypes"`
	Diets          json.RawMessage `db:"diets" json:"diets"`
	Query          string
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
		ID:         cr.ID,
		Name:       cr.Name,
		TotalSteps: cr.TotalSteps,
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
			Steps:       r.Steps,
			HealthScore: r.HealthScore,
			Diets:       string(r.Diets),
			DishTypes:   string(r.DishTypes),
			Version:     r.Version,
			Query:       r.Query,
		})
	}
	return RecipeItems
}

func ConvertGenRecipeToRecipeModel(rt []RecipeTable) []models.RecipeModel {
	RecipeItems := make([]models.RecipeModel, 0, len(rt))
	for _, r := range rt {
		RecipeItems = append(RecipeItems, models.RecipeModel{
			ID:          r.ID,
			Name:        r.Name,
			Desc:        r.Desc,
			Img:         r.Img,
			CookingTime: r.CookingTime,
			ServingsNum: r.ServingsNum,
			Steps:       r.Steps,
			HealthScore: r.HealthScore,
			Diets:       string(r.Diets),
			DishTypes:   string(r.DishTypes),
			Ingredients: r.Ingredients,
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

func ParseGeneratedRecipe(rawData json.RawMessage) (GeneratedRecipe, error) {
	var recipe GeneratedRecipe
	err := json.Unmarshal(rawData, &recipe)
	if err != nil {
		return GeneratedRecipe{}, fmt.Errorf("failed to parse recipe: %w", err)
	}
	return recipe, nil
}

func ConvertGeneratedRecipeToRecipeModels(gr []GeneratedRecipe) []models.RecipeModel {
	RecipeItems := make([]models.RecipeModel, 0, len(gr))
	for _, recipe := range gr {
		RecipeItems = append(RecipeItems, models.RecipeModel{
			ID:          recipe.ID,
			Name:        recipe.Name,
			Desc:        recipe.Desc,
			ServingsNum: recipe.ServingsNum,
			Steps:       string(recipe.Steps),
			DishTypes:   string(recipe.DishTypes),
			Diets:       string(recipe.Diets),
			Ingredients: recipe.Ingredients,
			Version:     recipe.Version,
			Query:       recipe.Query,
		})
	}
	return RecipeItems
}
