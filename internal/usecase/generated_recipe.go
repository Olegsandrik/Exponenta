package usecase

import (
	"context"
	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
)

type GenerateRepository interface {
	GetAllRecipes(ctx context.Context, num int, userID uint) ([]models.RecipeModel, error)
	GetRecipeByID(ctx context.Context, recipeID int, userID uint) ([]models.RecipeModel, error)
	CreateRecipe(ctx context.Context, products []string, query string, userID uint) ([]models.RecipeModel, error)
	UpdateRecipe(ctx context.Context, query string, recipeID int, versionID int,
		userID uint) ([]models.RecipeModel, error)
	GetHistoryByID(ctx context.Context, recipeID int, userID uint) ([]models.RecipeModel, error)
	SetNewMainVersion(ctx context.Context, recipeID int, versionID int, userID uint) error
	StartCookingByRecipeID(ctx context.Context, recipeID int, userID uint) error
}

type GenerateUsecase struct {
	Repository GenerateRepository
}

func NewGenerateUsecase(generateRepository GenerateRepository) *GenerateUsecase {
	return &GenerateUsecase{Repository: generateRepository}
}

func (a *GenerateUsecase) GetAllRecipes(ctx context.Context, num int) ([]dto.RecipeDto, error) {
	uID := uint(1)

	recipesModels, err := a.Repository.GetAllRecipes(ctx, num, uID)
	if err != nil {
		return nil, err
	}

	recipesDto := models.ConvertRecipeToDto(recipesModels)

	return recipesDto, nil
}

func (a *GenerateUsecase) GetRecipeByID(ctx context.Context, recipeID int) (dto.RecipeDto, error) {
	uID := uint(1)

	recipeModel, err := a.Repository.GetRecipeByID(ctx, recipeID, uID)

	if err != nil {
		return dto.RecipeDto{}, err
	}

	recipeDTO := models.ConvertRecipeToDto(recipeModel)

	return recipeDTO[0], nil
}

func (a *GenerateUsecase) CreateRecipe(ctx context.Context, products []string, query string) (dto.RecipeDto, error) {
	uID := uint(1)

	recipeModel, err := a.Repository.CreateRecipe(ctx, products, query, uID)
	if err != nil {
		return dto.RecipeDto{}, err
	}

	recipeDTO := models.ConvertRecipeToDto(recipeModel)

	return recipeDTO[0], nil
}

func (a *GenerateUsecase) UpdateRecipe(ctx context.Context, query string, recipeID int, versionID int) (dto.RecipeDto, error) {
	uID := uint(1)

	recipeModel, err := a.Repository.UpdateRecipe(ctx, query, recipeID, versionID, uID)

	if err != nil {
		return dto.RecipeDto{}, err
	}

	recipeDTO := models.ConvertRecipeToDto(recipeModel)

	return recipeDTO[0], nil
}

func (a *GenerateUsecase) GetHistoryByID(ctx context.Context, recipeID int) ([]dto.RecipeDto, error) {
	uID := uint(1)

	recipeModel, err := a.Repository.GetHistoryByID(ctx, recipeID, uID)

	if err != nil {
		return nil, err
	}

	recipeDTO := models.ConvertRecipeToDto(recipeModel)

	return recipeDTO, nil
}

func (a *GenerateUsecase) SetNewMainVersion(ctx context.Context, recipeID int, versionID int) error {
	uID := uint(1)

	return a.Repository.SetNewMainVersion(ctx, recipeID, versionID, uID)
}

func (a *GenerateUsecase) StartCookingByRecipeID(ctx context.Context, recipeID int) error {
	uID := uint(1)
	return a.Repository.StartCookingByRecipeID(ctx, recipeID, uID)
}
