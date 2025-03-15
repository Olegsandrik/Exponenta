package repository

import (
	"context"
	DB "github.com/Olegsandrik/Exponenta/internal/adapters/postgres"
	"github.com/Olegsandrik/Exponenta/internal/repository/dao"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/logger"
)

type CookingRecipeRepo struct {
	storage *DB.Adapter
}

func NewCookingRecipeRepo(storage *DB.Adapter) *CookingRecipeRepo {
	return &CookingRecipeRepo{
		storage: storage,
	}
}

func (repo *CookingRecipeRepo) GetAllRecipe(ctx context.Context, num int) []models.RecipeModel {
	q := `SELECT name, description, image FROM public.recipes LIMIT $1`
	var recipeRows []dao.RecipeTable
	err := repo.storage.Select(ctx, &recipeRows, q, num)
	if err != nil {
		logger.Error(ctx, "Query error", err)
		return nil
	}
	recipeItems := dao.ConvertDaoToModel(recipeRows)
	return recipeItems
}

func (repo *CookingRecipeRepo) GetRecipeByID(ctx context.Context, id int) []models.RecipeModel {
	// 645348 salad
	q := `SELECT r.name, r.description, r.image, r.steps FROM public.recipes as r WHERE id = $1`
	var recipeRows []dao.RecipeTable
	err := repo.storage.Select(ctx, &recipeRows, q, id)
	if err != nil {
		logger.Error(ctx, "Query error", err)
		return nil
	}
	logger.Info(ctx, "get: ", recipeRows)
	recipeItem := dao.ConvertDaoToModel(recipeRows)
	return recipeItem
}
