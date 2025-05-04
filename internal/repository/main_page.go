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
	FROM recipes, counter WHERE jsonb_exists(dish_types::jsonb, $2) LIMIT $3 OFFSET $4;`

	recipeRows := make([]dao.MainPageRecipeTable, 0, pageSizeConst)

	err := r.adapter.Select(ctx, &recipeRows, q,
		dishType, dishType, pageSizeConst, page*pageSizeConst-pageSizeConst)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"error getting recipe rows: %s with page: %d, dishType: %s", err.Error(), page, dishType))
		return nil, 0, internalErrors.ErrFailToGetRecipes
	}

	if len(recipeRows) == 0 {
		logger.Error(ctx, fmt.Sprintf(
			"error getting recipe zero row with num: page: %d, dishType: %s", page, dishType))
		if page > 1 {
			return nil, 0, internalErrors.ErrGetZeroRowsWithPageGreaterThanOne
		}
		return nil, 0, internalErrors.ErrZeroRowsGet
	}

	recipeItems := dao.ConvertMainPageRecipeTableToRecipeModel(recipeRows)

	logger.Info(ctx, fmt.Sprintf("success select recipes with dishType: %s, page: %d", dishType, page))

	return recipeItems, (recipeRows[0].TotalNum + pageSizeConst - 1) / pageSizeConst, nil
}

func (r *MainPageRepository) GetRecipesByDiet(
	ctx context.Context, diet string, page int) ([]models.RecipeModel, int, error) {
	q := `WITH counter AS (
		SELECT count(*) as total_count
		FROM recipes
		WHERE jsonb_exists(diets::jsonb, $1)
	) SELECT id, name, description, image, ready_in_minutes, counter.total_count
	FROM recipes, counter WHERE jsonb_exists(diets::jsonb, $2) LIMIT $3 OFFSET $4;`

	recipeRows := make([]dao.MainPageRecipeTable, 0, pageSizeConst)

	err := r.adapter.Select(ctx, &recipeRows, q,
		diet, diet, pageSizeConst, page*pageSizeConst-pageSizeConst)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"error getting recipe rows: %s with page: %d, diet: %s", err.Error(), page, diet))
		return nil, 0, internalErrors.ErrFailToGetRecipes
	}

	if len(recipeRows) == 0 {
		logger.Error(ctx, fmt.Sprintf(
			"error getting recipe zero row with num: page: %d, diet: %s", page, diet))
		if page > 1 {
			return nil, 0, internalErrors.ErrGetZeroRowsWithPageGreaterThanOne
		}
		return nil, 0, internalErrors.ErrZeroRowsGet
	}

	recipeItems := dao.ConvertMainPageRecipeTableToRecipeModel(recipeRows)

	logger.Info(ctx, fmt.Sprintf("success select recipes with diet: %s, page: %d", diet, page))

	return recipeItems, (recipeRows[0].TotalNum + pageSizeConst - 1) / pageSizeConst, nil
}

func (r *MainPageRepository) GetCollectionByID(
	ctx context.Context, collectionID int, page int) ([]models.RecipeModel, int, error) {
	q := `WITH counter AS (
		SELECT count(*) as total_count
		FROM recipes_collection_recipes
		WHERE collection_id = $1
	)
	SELECT rc.recipe_id as id, r.name, r.description, r.ready_in_minutes, r.image, 
	(SELECT total_count FROM counter) as total_count
	FROM recipes_collection_recipes as rc
	LEFT JOIN recipes as r ON rc.recipe_id = r.id
	WHERE collection_id = $2 LIMIT $3 OFFSET $4;`

	recipeRows := make([]dao.MainPageRecipeTable, 0, pageSizeConst)

	err := r.adapter.Select(ctx, &recipeRows, q,
		collectionID, collectionID, pageSizeConst, page*pageSizeConst-pageSizeConst)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"error getting recipe rows: %v, collectionID: %d, page: %d", err, collectionID, page))
		return nil, 0, internalErrors.ErrFailToGetRecipes
	}

	if len(recipeRows) == 0 {
		logger.Error(ctx, fmt.Sprintf(
			"getting zero recipe rows collectionID: %d, page: %d", collectionID, page))
		if page > 1 {
			return nil, 0, internalErrors.ErrGetZeroRowsWithPageGreaterThanOne
		}
		return nil, 0, internalErrors.ErrZeroRowsGet
	}

	recipeItems := dao.ConvertMainPageRecipeTableToRecipeModel(recipeRows)

	logger.Info(ctx, fmt.Sprintf(
		"success select recipes with collectionID: %d, page: %d", collectionID, page))

	return recipeItems, (recipeRows[0].TotalNum + pageSizeConst - 1) / pageSizeConst, nil
}

func (r *MainPageRepository) GetAllCollections(
	ctx context.Context) ([]models.Collection, error) {
	q := `SELECT id, name FROM recipes_collection;`

	collectionRows := make([]dao.CollectionTable, 0)

	err := r.adapter.Select(ctx, &collectionRows, q)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("err with getting collections %v", err))
		return []models.Collection{}, internalErrors.ErrGetCollections
	}

	if len(collectionRows) == 0 {
		logger.Error(ctx, "zero rows found")
		return []models.Collection{}, internalErrors.ErrGetCollections
	}

	logger.Info(ctx, "success get all collections")
	collections := dao.ConvertCollectionTableToModel(collectionRows)

	return collections, nil
}
