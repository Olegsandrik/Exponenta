package usecase

import (
	"context"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
)

type CookingHistoryRepo interface {
	GetRecipesFromHistory(ctx context.Context, uID uint, page int) ([]models.RecipeModel, error)
}

type CookingHistoryUsecase struct {
	cookingHistoryRepo  CookingHistoryRepo
	favoriteRecipesRepo FavoriteRecipesRepo
}

func NewCookingHistoryUsecase(cookingHistoryRepo CookingHistoryRepo,
	favoriteRecipesRepo FavoriteRecipesRepo) *CookingHistoryUsecase {
	return &CookingHistoryUsecase{
		cookingHistoryRepo:  cookingHistoryRepo,
		favoriteRecipesRepo: favoriteRecipesRepo,
	}
}

func (u *CookingHistoryUsecase) GetRecipesFromHistory(
	ctx context.Context, uID uint, page int) ([]dto.RecipeDto, error) {
	recipeModels, err := u.cookingHistoryRepo.GetRecipesFromHistory(ctx, uID, page)
	if err != nil {
		return nil, err
	}

	recipeDTO := models.ConvertRecipeToDto(recipeModels)

	favoriteIDsSet, err := u.favoriteRecipesRepo.GetAllIDFavoriteRecipes(ctx, uID)
	if err != nil {
		return []dto.RecipeDto{}, err
	}

	for i := 0; i < len(recipeDTO); i++ {
		if recipeDTO[i].IsGenerated {
			continue
		}

		_, ok := favoriteIDsSet[recipeDTO[i].ID]
		if ok {
			recipeDTO[i].IsFavorite = true
		}
	}

	return recipeDTO, nil
}
