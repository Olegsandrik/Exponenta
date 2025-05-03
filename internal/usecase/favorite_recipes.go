package usecase

import (
	"context"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	internalErrors "github.com/Olegsandrik/Exponenta/internal/internalerrors"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
)

type FavoriteRecipesRepo interface {
	AddRecipeToFavorite(ctx context.Context, userID uint, recipeID int) error
	DeleteRecipeFromFavorite(ctx context.Context, userID uint, recipeID int) error
	GetFavoriteRecipes(ctx context.Context, userID uint, page int) ([]models.RecipeModel, error)
}

type FavoriteRecipesUsecase struct {
	RecipeRepo FavoriteRecipesRepo
}

func NewFavoriteRecipesUsecase(repo FavoriteRecipesRepo) *FavoriteRecipesUsecase {
	return &FavoriteRecipesUsecase{
		RecipeRepo: repo,
	}
}

func (u *FavoriteRecipesUsecase) AddRecipeToFavorite(ctx context.Context, userID uint, recipeID int) error {
	return u.RecipeRepo.AddRecipeToFavorite(ctx, userID, recipeID)
}

func (u *FavoriteRecipesUsecase) DeleteRecipeFromFavorite(ctx context.Context, userID uint, recipeID int) error {
	return u.RecipeRepo.DeleteRecipeFromFavorite(ctx, userID, recipeID)
}

func (u *FavoriteRecipesUsecase) GetFavoriteRecipes(
	ctx context.Context, userID uint, page int) ([]dto.RecipeDto, error) {
	recipeDao, err := u.RecipeRepo.GetFavoriteRecipes(ctx, userID, page)
	if err != nil {
		return nil, err
	}

	if len(recipeDao) == 0 {
		return []dto.RecipeDto{}, internalErrors.ErrZeroRowsGet
	}

	recipe := models.ConvertRecipeToDto(recipeDao)
	return recipe, nil
}
