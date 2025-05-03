package usecase

import (
	"context"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/internal/utils"
)

type GenerateRepository interface {
	GetAllRecipes(ctx context.Context, num int, userID uint) ([]models.RecipeModel, error)
	GetRecipeByID(ctx context.Context, recipeID int, userID uint) ([]models.RecipeModel, error)
	CreateRecipe(ctx context.Context, products []string, query string, userID uint) ([]models.RecipeModel, error)
	UpdateRecipe(ctx context.Context, query string, recipeID int, versionID int,
		userID uint) ([]models.RecipeModel, error)
	GetHistoryByID(ctx context.Context, recipeID int, userID uint) ([]models.RecipeModel, error)
	SetNewMainVersion(ctx context.Context, recipeID int, versionID int, userID uint) error
}

type GenerateUsecase struct {
	GenRepository    GenerateRepository
	RecipeRepository CookingRecipeRepo
}

func NewGenerateUsecase(generateRepository GenerateRepository, recipeRepo CookingRecipeRepo) *GenerateUsecase {
	return &GenerateUsecase{GenRepository: generateRepository, RecipeRepository: recipeRepo}
}

func (a *GenerateUsecase) GetAllRecipes(ctx context.Context, num int) ([]dto.RecipeDto, error) {
	uID, err := utils.GetUserIDFromContext(ctx)

	if err != nil {
		return nil, err
	}

	recipesModels, err := a.GenRepository.GetAllRecipes(ctx, num, uID)
	if err != nil {
		return nil, err
	}

	recipesDto := models.ConvertRecipeToDto(recipesModels)

	return recipesDto, nil
}

func (a *GenerateUsecase) GetRecipeByID(ctx context.Context, recipeID int) (dto.RecipeDto, error) {
	uID, err := utils.GetUserIDFromContext(ctx)

	if err != nil {
		return dto.RecipeDto{}, err
	}

	recipeModel, err := a.GenRepository.GetRecipeByID(ctx, recipeID, uID)

	if err != nil {
		return dto.RecipeDto{}, err
	}

	recipeDTO := models.ConvertRecipeToDto(recipeModel)

	return recipeDTO[0], nil
}

func (a *GenerateUsecase) CreateRecipe(ctx context.Context, products []string, query string) (dto.RecipeDto, error) {
	uID, err := utils.GetUserIDFromContext(ctx)

	if err != nil {
		return dto.RecipeDto{}, err
	}

	recipeModel, err := a.GenRepository.CreateRecipe(ctx, products, query, uID)
	if err != nil {
		return dto.RecipeDto{}, err
	}

	recipeDTO := models.ConvertRecipeToDto(recipeModel)

	return recipeDTO[0], nil
}

func (a *GenerateUsecase) UpdateRecipe(ctx context.Context, query string, recipeID int,
	versionID int) (dto.RecipeDto, error) {
	uID, err := utils.GetUserIDFromContext(ctx)

	if err != nil {
		return dto.RecipeDto{}, err
	}

	recipeModel, err := a.GenRepository.UpdateRecipe(ctx, query, recipeID, versionID, uID)

	if err != nil {
		return dto.RecipeDto{}, err
	}

	recipeDTO := models.ConvertRecipeToDto(recipeModel)

	return recipeDTO[0], nil
}

func (a *GenerateUsecase) GetHistoryByID(ctx context.Context, recipeID int) ([]dto.RecipeDto, error) {
	uID, err := utils.GetUserIDFromContext(ctx)

	if err != nil {
		return nil, err
	}

	recipeModel, err := a.GenRepository.GetHistoryByID(ctx, recipeID, uID)

	if err != nil {
		return nil, err
	}

	recipeDTO := models.ConvertRecipeToDto(recipeModel)

	return recipeDTO, nil
}

func (a *GenerateUsecase) SetNewMainVersion(ctx context.Context, recipeID int, versionID int) error {
	uID, err := utils.GetUserIDFromContext(ctx)

	if err != nil {
		return err
	}

	return a.GenRepository.SetNewMainVersion(ctx, recipeID, versionID, uID)
}

func (a *GenerateUsecase) StartCookingByRecipeID(ctx context.Context, recipeID int) (dto.CurrentStepRecipeDto, error) {
	uID, err := utils.GetUserIDFromContext(ctx)

	if err != nil {
		return dto.CurrentStepRecipeDto{}, err
	}

	err = a.RecipeRepository.StartCooking(ctx, uID, recipeID, "public.generated_recipes")
	if err != nil {
		return dto.CurrentStepRecipeDto{}, err
	}

	currentRecipeStepModel, err := a.RecipeRepository.GetCurrentStep(ctx, uID)
	if err != nil {
		return dto.CurrentStepRecipeDto{}, err
	}

	currentRecipeStep := models.ConvertCurrentStepToDTO(currentRecipeStepModel)

	return currentRecipeStep, nil
}
