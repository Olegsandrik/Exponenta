package repository

import (
	"context"
	"fmt"
	"github.com/Olegsandrik/Exponenta/internal/adapters/postgres"
	internalErrors "github.com/Olegsandrik/Exponenta/internal/internalerrors"
	"github.com/Olegsandrik/Exponenta/internal/repository/dao"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/logger"
)

const pageSizeConst = 6

type MainPageRepository struct {
	adapter *postgres.Adapter
}

func NewMainPageRepository(adapter *postgres.Adapter) *MainPageRepository {
	return &MainPageRepository{adapter: adapter}
}

func (r *MainPageRepository) GetRecipesByDishType(
	ctx context.Context, dishType string, page int) ([]models.RecipeModel, int, error) {
	q := `WITH counter AS (
		SELECT count(*) as total_count
		FROM recipes
		WHERE jsonb_exists(dish_types::jsonb, $1)
	) SELECT id, name, description, image, ready_in_minutes, counter.total_count
	FROM recipes, counter WHERE jsonb_exists(dish_types::jsonb, $2) LIMIT 6 OFFSET $3;`

	recipeRows := make([]dao.MainPageRecipeTable, 0, pageSizeConst)

	err := r.adapter.Select(ctx, &recipeRows, q, dishType, dishType, page*pageSizeConst-pageSizeConst)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"error getting recipe rows: %s with page: %d, dishType: %s", err.Error(), page, dishType))
		return nil, 0, internalErrors.ErrFailToGetRecipes
	}

	if len(recipeRows) == 0 {
		logger.Error(ctx, fmt.Sprintf(
			"error getting recipe zero row with num: page: %d, dishType: %s", page, dishType))
		return nil, 0, internalErrors.ErrFailToGetRecipes
	}

	recipeItems := dao.ConvertMainPageRecipeTableToRecipeModel(recipeRows)

	logger.Info(ctx, fmt.Sprintf("select recipes with dishType: %s, page: %d", dishType, page))

	return recipeItems, recipeRows[0].TotalNum, nil
}

func (r *MainPageRepository) GetRecipesByDiet(
	ctx context.Context, diet string, page int) ([]models.RecipeModel, int, error) {
	q := `WITH counter AS (
		SELECT count(*) as total_count
		FROM recipes
		WHERE jsonb_exists(diets::jsonb, $1)
	) SELECT id, name, description, image, ready_in_minutes, counter.total_count
	FROM recipes, counter WHERE jsonb_exists(diets::jsonb, $2) LIMIT 6 OFFSET $3;`

	recipeRows := make([]dao.MainPageRecipeTable, 0, pageSizeConst)

	err := r.adapter.Select(ctx, &recipeRows, q, diet, diet, page*pageSizeConst-pageSizeConst)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"error getting recipe rows: %s with page: %d, diet: %s", err.Error(), page, diet))
		return nil, 0, internalErrors.ErrFailToGetRecipes
	}

	if len(recipeRows) == 0 {
		logger.Error(ctx, fmt.Sprintf(
			"error getting recipe zero row with num: page: %d, diet: %s", page, diet))
		return nil, 0, internalErrors.ErrFailToGetRecipes
	}

	recipeItems := dao.ConvertMainPageRecipeTableToRecipeModel(recipeRows)

	logger.Info(ctx, fmt.Sprintf("select recipes with diet: %s, page: %d", diet, page))

	return recipeItems, recipeRows[0].TotalNum, nil
}

func (r *MainPageRepository) GetCollectionByID(
	ctx context.Context, collectionID int, page int) ([]models.RecipeModel, int, error) {
	return []models.RecipeModel{}, 0, nil
}

func (r *MainPageRepository) GetAllCollections(
	ctx context.Context) ([]models.Collection, error) {
	return []models.Collection{}, nil
}
