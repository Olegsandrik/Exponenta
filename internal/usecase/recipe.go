package usecase

import (
	"context"
	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/utils"
)

type CookingRecipeRepo interface {
	GetAllRecipe(ctx context.Context, num int) ([]models.RecipeModel, error)
	GetRecipeByID(ctx context.Context, id int) ([]models.RecipeModel, error)
	EndCooking(ctx context.Context, uId uint) error
	StartCooking(ctx context.Context, uId uint, recipeId int) error
	GetCurrentRecipe(ctx context.Context, uId uint) (models.CurrentRecipeModel, error)
	GetNextRecipeStep(ctx context.Context, uId uint) (models.CurrentStepRecipeModel, error)
	GetPrevRecipeStep(ctx context.Context, uId uint) (models.CurrentStepRecipeModel, error)
	GetCurrentStep(ctx context.Context, uId uint) (models.CurrentStepRecipeModel, error)
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
	recipeModels, err := u.repo.GetAllRecipe(ctx, num)

	if err != nil {
		return nil, err
	}

	utils.SanitizeRecipeDescription(recipeModels)

	recipeDto := models.ConvertRecipeToDto(recipeModels)

	return recipeDto, nil
}

func (u *CookingRecipeUsecase) GetRecipeByID(ctx context.Context, id int) (dto.RecipeDto, error) {
	recipeModels, err := u.repo.GetRecipeByID(ctx, id)

	if err != nil {
		return dto.RecipeDto{}, err
	}

	utils.SanitizeRecipeDescription(recipeModels)

	recipeDto := models.ConvertRecipeToDto(recipeModels)

	return recipeDto[0], nil
}

func (u *CookingRecipeUsecase) StartCookingRecipe(ctx context.Context, recipeId int) (dto.CurrentStepRecipeDto, error) {
	uId := uint(1)

	err := u.repo.StartCooking(ctx, uId, recipeId)
	if err != nil {
		return dto.CurrentStepRecipeDto{}, err
	}

	currentRecipeStepModel, err := u.repo.GetCurrentStep(ctx, uId)
	if err != nil {
		return dto.CurrentStepRecipeDto{}, err
	}

	currentRecipeStep := models.ConvertCurrentStepToDTO(currentRecipeStepModel)

	return currentRecipeStep, nil
}

func (u *CookingRecipeUsecase) EndCookingRecipe(ctx context.Context) error {
	uId := uint(1)

	err := u.repo.EndCooking(ctx, uId)
	if err != nil {
		return err
	}

	return nil
}

func (u *CookingRecipeUsecase) GetCurrentRecipe(ctx context.Context) (dto.CurrentRecipeDto, error) {
	uId := uint(1)

	currentRecipe, err := u.repo.GetCurrentRecipe(ctx, uId)
	if err != nil {
		return dto.CurrentRecipeDto{}, err
	}

	currentRecipeDto := models.ConvertCurrentRecipeToDTO(currentRecipe)

	return currentRecipeDto, nil
}

func (u *CookingRecipeUsecase) NextStepRecipe(ctx context.Context) (dto.CurrentStepRecipeDto, error) {
	uId := uint(1)

	nextStep, err := u.repo.GetNextRecipeStep(ctx, uId)
	if err != nil {
		return dto.CurrentStepRecipeDto{}, err
	}

	nextStepDto := models.ConvertCurrentStepToDTO(nextStep)

	return nextStepDto, nil
}

func (u *CookingRecipeUsecase) PreviousStepRecipe(ctx context.Context) (dto.CurrentStepRecipeDto, error) {
	uId := uint(1)

	prevStep, err := u.repo.GetPrevRecipeStep(ctx, uId)
	if err != nil {
		return dto.CurrentStepRecipeDto{}, err
	}

	prevStepDto := models.ConvertCurrentStepToDTO(prevStep)

	return prevStepDto, nil
}
