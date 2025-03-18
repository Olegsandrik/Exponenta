package usecase

import (
	"context"
	"errors"
	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/utils"
)

type CookingRecipeRepo interface {
	GetAllRecipe(ctx context.Context, num int) []models.RecipeModel
	GetRecipeByID(ctx context.Context, id int) []models.RecipeModel
	EndCooking(ctx context.Context, uId uint) error
	StartCooking(ctx context.Context, uId uint, recipeId int) error
	GetCurrentRecipe(ctx context.Context, uId uint) models.CurrentRecipe
}

type CookingRecipeUsecase struct {
	repo CookingRecipeRepo
}

func NewCookingRecipeUsecase(repo CookingRecipeRepo) *CookingRecipeUsecase {
	return &CookingRecipeUsecase{
		repo: repo,
	}
}

func (u *CookingRecipeUsecase) GetAllRecipe(ctx context.Context, num int) ([]dto.RecipeDto, error) {
	recipeModels := u.repo.GetAllRecipe(ctx, num)

	utils.SanitizeRecipeDescription(recipeModels)

	recipeDto := models.ConvertModelToDto(recipeModels)

	return recipeDto, nil
}

func (u *CookingRecipeUsecase) GetRecipeByID(ctx context.Context, id int) ([]dto.RecipeDto, error) {
	recipeModels := u.repo.GetRecipeByID(ctx, id)

	utils.SanitizeRecipeDescription(recipeModels)

	recipeDto := models.ConvertModelToDto(recipeModels)

	return recipeDto, nil
}

func (u *CookingRecipeUsecase) StartCookingRecipe(ctx context.Context, recipeId int) error {
	uId := uint(1)

	err := u.repo.StartCooking(ctx, uId, recipeId)
	if err != nil {
		return err
	}

	return nil
}

func (u *CookingRecipeUsecase) EndCookingRecipe(ctx context.Context) error {
	uId := uint(1)

	err := u.repo.EndCooking(ctx, uId)
	if err != nil {
		return err
	}

	return nil
}

func (u *CookingRecipeUsecase) GetCurrentRecipe(context.Context) (dto.CurrentRecipeDto, error) {
	uId := uint(1)

	currentRecipe := u.repo.GetCurrentRecipe(context.Background(), uId)

	// надо потестить
	if currentRecipe.Id == 0 {
		return dto.CurrentRecipeDto{}, errors.New("recipe not found")
	}

	currentRecipeDto := models.ConvertCurrentRecipeToDTO(currentRecipe)

	return currentRecipeDto, nil
}

func (u *CookingRecipeUsecase) NextStepRecipe(context.Context) (dto.CurrentStepRecipeDto, error) {
	return dto.CurrentStepRecipeDto{}, nil
}

func (u *CookingRecipeUsecase) PreviousStepRecipe(context.Context) (dto.CurrentStepRecipeDto, error) {
	return dto.CurrentStepRecipeDto{}, nil
}
