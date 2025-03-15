package usecase

import (
	"context"
	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/logger"
	"regexp"
	"strconv"
)

type CookingRecipeRepo interface {
	GetAllRecipe(ctx context.Context, num int) []models.RecipeModel
	GetRecipeByID(ctx context.Context, id int) []models.RecipeModel
}

type CookingRecipeUsecase struct {
	repo CookingRecipeRepo
}

func NewCookingRecipeUsecase(repo CookingRecipeRepo) *CookingRecipeUsecase {
	return &CookingRecipeUsecase{
		repo: repo,
	}
}

func (u *CookingRecipeUsecase) GetAllRecipe(ctx context.Context, numStr string) ([]dto.RecipeDto, error) {
	num, err := strconv.Atoi(numStr)
	if err != nil {
		logger.Error(ctx, "Atoi error", err)
		return nil, err
	}

	recipeModels := u.repo.GetAllRecipe(ctx, num)

	re := regexp.MustCompile(`<[^>]*>`)
	for i := range recipeModels {
		recipeModels[i].Desc = re.ReplaceAllString(recipeModels[i].Desc, "")
	}

	recipeDto := models.ConvertModelToDto(recipeModels)

	return recipeDto, nil
}

func (u *CookingRecipeUsecase) GetRecipeByID(ctx context.Context, idStr string) ([]dto.RecipeDto, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error(ctx, "Atoi error", err)
		return nil, err
	}
	recipeModels := u.repo.GetRecipeByID(ctx, id)
	recipeDto := models.ConvertModelToDto(recipeModels)
	return recipeDto, nil
}
