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

type FavoriteRecipeRepository struct {
	adapter *postgres.Adapter
}

func NewFavoriteRecipeRepository(adapter *postgres.Adapter) *FavoriteRecipeRepository {
	return &FavoriteRecipeRepository{
		adapter: adapter,
	}
}

func (r *FavoriteRecipeRepository) AddRecipeToFavorite(ctx context.Context, userID uint, recipeID int) error {
	q := `INSERT INTO favorite_recipes (user_ID, recipe_ID) VALUES ($1, $2);`

	result, err := r.adapter.Exec(ctx, q, userID, recipeID)

	if err != nil {
		if r.adapter.IsDuplicateKeyError(err) {
			logger.Error(ctx, fmt.Sprintf(
				"duplicate key error on add recipe to favorite recipes: %v for userID: %d and recipeID: %d",
				err,
				userID,
				recipeID),
			)
			return internalErrors.ErrDuplicateRow
		} else if r.adapter.IsNotExistForeignKey(err) {
			logger.Error(ctx, fmt.Sprintf(
				"recipe with ID: %d does not exist: %v",
				recipeID,
				err),
			)
			return internalErrors.ErrRecipeWithThisIDDoesNotExist
		}
		logger.Error(ctx, fmt.Sprintf(
			"error on add recipe to favorite recipes: %v for userID: %d and recipeID: %d",
			err,
			userID,
			recipeID),
		)
		return internalErrors.ErrFailedToAddFavoriteRecipe
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"fail to get rows affected on add recipe to favorite recipes: %v for userID: %d and recipeID: %d",
			err,
			userID,
			recipeID),
		)
		return internalErrors.ErrFailedToAddFavoriteRecipe
	}

	if rowsAffected == 0 {
		logger.Error(ctx, fmt.Sprintf(
			"zero rows affected on add recipe to favorite recipes: %v for userID: %d and recipeID: %d",
			err,
			userID,
			recipeID),
		)
		return internalErrors.ErrFailedToAddFavoriteRecipe
	}

	logger.Info(ctx, fmt.Sprintf(
		"successfully add recipe %d to favorite recipes for user %d ", recipeID, userID))
	return nil
}

func (r *FavoriteRecipeRepository) DeleteRecipeFromFavorite(ctx context.Context, userID uint, recipeID int) error {
	q := `DELETE FROM favorite_recipes WHERE user_ID = $1 AND recipe_ID = $2`

	result, err := r.adapter.Exec(ctx, q, userID, recipeID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"error on delete recipe from favorite recipes: %v for userID: %d and recipeID: %d",
			err,
			userID,
			recipeID),
		)
		return internalErrors.ErrFailedToDeleteFavoriteRecipe
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"fail to get rows affected on delete recipe from favorite recipes: %v for userID: %d and recipeID: %d",
			err,
			userID,
			recipeID),
		)
		return internalErrors.ErrFailedToDeleteFavoriteRecipe
	}

	if rowsAffected == 0 {
		logger.Error(ctx, fmt.Sprintf(
			"zero rows affected on delete recipe from favorite recipes: %v for userID: %d and recipeID: %d",
			err,
			userID,
			recipeID),
		)
		return internalErrors.ErrZeroRowsDeleted
	}

	logger.Info(ctx, fmt.Sprintf(
		"successfully delete recipe %d from favorite recipes for user %d ", recipeID, userID))
	return nil
}

func (r *FavoriteRecipeRepository) GetFavoriteRecipes(
	ctx context.Context, userID uint, page int) ([]models.RecipeModel, error) {
	q := `SELECT id, name, description, image, ready_in_minutes FROM public.recipes WHERE id IN (
    SELECT recipe_id FROM public.favorite_recipes WHERE user_id = $1
    ) LIMIT 6 OFFSET $2;`

	recipeRows := make([]dao.RecipeTable, 0, pageSizeConst)

	err := r.adapter.Select(ctx, &recipeRows, q, userID, page*pageSizeConst-pageSizeConst)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"error getting recipe rows: %s, userID: %d with page: %d", err.Error(), userID, page))
		return nil, internalErrors.ErrFailToGetRecipes
	}

	recipeItems := dao.ConvertDaoToRecipe(recipeRows)

	if len(recipeItems) == 0 {
		if page > 1 {
			logger.Error(ctx, fmt.Sprintf(
				"error getting recipe zero row with page: %d for userID: %d", page, userID))
			return nil, internalErrors.ErrGetZeroRowsWithPageGreaterThanOne
		}
		logger.Error(ctx, fmt.Sprintf(
			"error getting recipe zero row with page: %d for userID: %d", page, userID))
		return nil, internalErrors.ErrZeroRowsGet
	}

	logger.Info(ctx, fmt.Sprintf("select %d recipes", len(recipeRows)))

	return recipeItems, nil
}
