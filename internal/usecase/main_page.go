package usecase

import (
	"context"
	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
)

type MainPageRepository interface {
	GetRecipesByDishType(ctx context.Context, dishType string, page int) ([]models.RecipeModel, int, error)
	GetRecipesByDiet(ctx context.Context, diet string, page int) ([]models.RecipeModel, int, error)
	GetCollectionByID(ctx context.Context, collectionID int, page int) ([]models.RecipeModel, int, error)
	GetAllCollections(ctx context.Context) ([]models.Collection, error)
}

type MainPageUsecase struct {
	repository MainPageRepository
}

func NewMainPageUsecase(repository MainPageRepository) *MainPageUsecase {
	return &MainPageUsecase{repository: repository}
}

func (u *MainPageUsecase) GetRecipesByDishType(ctx context.Context, dishType string, page int) (dto.RecipePage, error) {
	recipes, lastPageNum, err := u.repository.GetRecipesByDishType(ctx, dishType, page)
	if err != nil {
		return dto.RecipePage{}, err
	}

	return dto.RecipePage{Recipes: models.ConvertRecipeToDto(recipes), LastPageNum: lastPageNum}, nil
}

func (u *MainPageUsecase) GetRecipesByDiet(ctx context.Context, diet string, page int) (dto.RecipePage, error) {
	recipes, lastPageNum, err := u.repository.GetRecipesByDiet(ctx, diet, page)
	if err != nil {
		return dto.RecipePage{}, err
	}

	return dto.RecipePage{Recipes: models.ConvertRecipeToDto(recipes), LastPageNum: lastPageNum}, nil
}
func (u *MainPageUsecase) GetCollectionByID(ctx context.Context, collectionID int, page int) (dto.RecipePage, error) {
	recipes, lastPageNum, err := u.repository.GetCollectionByID(ctx, collectionID, page)
	if err != nil {
		return dto.RecipePage{}, err
	}

	return dto.RecipePage{Recipes: models.ConvertRecipeToDto(recipes), LastPageNum: lastPageNum}, nil
}
func (u *MainPageUsecase) GetAllCollections(ctx context.Context) ([]dto.Collection, error) {
	collections, err := u.repository.GetAllCollections(ctx)
	if err != nil {
		return []dto.Collection{}, err
	}

	return models.ConvertCollectionToDTO(collections), nil
}
