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

type CookingHistoryRepo struct {
	adapter *postgres.Adapter
}

func NewCookingHistoryRepo(adapter *postgres.Adapter) *CookingHistoryRepo {
	return &CookingHistoryRepo{
		adapter: adapter,
	}
}

func (r *CookingHistoryRepo) GetRecipesFromHistory(
	ctx context.Context, uID uint, page int) ([]models.RecipeModel, error) {
	q := `SELECT r.id, r.name, r.description, r.image, r.ready_in_minutes, uch.is_generated, uch.created_at
		  FROM public.recipes r
		  JOIN public.user_cooking_history uch ON r.id = uch.recipe_id
		  WHERE uch.user_id = $1 AND uch.is_generated = false
		  UNION ALL
		  SELECT gr.id, gr.name, gr.description, 'null', gr.ready_in_minutes, uch.is_generated, uch.created_at
		  FROM public.generated_recipes gr
	      JOIN public.user_cooking_history uch ON gr.id = uch.recipe_id
		  WHERE uch.user_id = $1 AND uch.is_generated = true ORDER BY created_at DESC LIMIT $2 OFFSET $3;`

	recipeRows := make([]dao.RecipeTable, 0, pageSizeConst)

	err := r.adapter.Select(ctx, &recipeRows, q, uID, pageSizeConst, page*pageSizeConst-pageSizeConst)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"error getting recipe rows: %s, userID: %d with page: %d", err.Error(), uID, page))
		return nil, internalErrors.ErrFailToGetRecipes
	}

	recipeItems := dao.ConvertDaoToRecipe(recipeRows)

	if len(recipeItems) == 0 {
		if page > 1 {
			logger.Error(ctx, fmt.Sprintf(
				"error getting recipe zero row with page: %d for userID: %d", page, uID))
			return nil, internalErrors.ErrGetZeroRowsWithPageGreaterThanOne
		}
		logger.Error(ctx, fmt.Sprintf(
			"error getting recipe zero row with page: %d for userID: %d", page, uID))
		return nil, internalErrors.ErrZeroRowsGet
	}

	logger.Info(ctx, fmt.Sprintf("select %d recipes", len(recipeRows)))

	return recipeItems, nil
}
