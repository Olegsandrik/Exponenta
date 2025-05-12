package usecase

import (
	"context"
	"github.com/Olegsandrik/Exponenta/internal/utils"

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
	repository                MainPageRepository
	favoriteRecipesRepository FavoriteRecipesRepo
}

func NewMainPageUsecase(repository MainPageRepository, favoriteRecipesRepo FavoriteRecipesRepo) *MainPageUsecase {
	return &MainPageUsecase{repository: repository, favoriteRecipesRepository: favoriteRecipesRepo}
}

func (u *MainPageUsecase) GetRecipesByDishType(ctx context.Context, dishType string, page int) (dto.RecipePage, error) {
	recipes, lastPageNum, err := u.repository.GetRecipesByDishType(ctx, dishType, page)
	if err != nil {
		return dto.RecipePage{}, err
	}

	uID, err := utils.GetUserIDFromContext(ctx)
	if uID == 0 {
		return dto.RecipePage{Recipes: models.ConvertRecipeToDto(recipes), LastPageNum: lastPageNum}, nil
	}

	favoriteIDsSet, err := u.favoriteRecipesRepository.GetAllIDFavoriteRecipes(ctx, uID)
	if err != nil {
		return dto.RecipePage{}, err
	}

	recipeDTO := models.ConvertRecipeToDto(recipes)

	for i := 0; i < len(recipeDTO); i++ {
		_, ok := favoriteIDsSet[recipeDTO[i].ID]
		if ok {
			recipeDTO[i].IsFavorite = true
		}
	}

	return dto.RecipePage{Recipes: recipeDTO, LastPageNum: lastPageNum}, nil
}

func (u *MainPageUsecase) GetRecipesByDiet(ctx context.Context, diet string, page int) (dto.RecipePage, error) {
	recipes, lastPageNum, err := u.repository.GetRecipesByDiet(ctx, diet, page)
	if err != nil {
		return dto.RecipePage{}, err
	}

	uID, err := utils.GetUserIDFromContext(ctx)
	if uID == 0 {
		return dto.RecipePage{Recipes: models.ConvertRecipeToDto(recipes), LastPageNum: lastPageNum}, nil
	}

	favoriteIDsSet, err := u.favoriteRecipesRepository.GetAllIDFavoriteRecipes(ctx, uID)
	if err != nil {
		return dto.RecipePage{}, err
	}

	recipeDTO := models.ConvertRecipeToDto(recipes)

	for i := 0; i < len(recipeDTO); i++ {
		_, ok := favoriteIDsSet[recipeDTO[i].ID]
		if ok {
			recipeDTO[i].IsFavorite = true
		}
	}

	return dto.RecipePage{Recipes: recipeDTO, LastPageNum: lastPageNum}, nil
}

func (u *MainPageUsecase) GetCollectionByID(ctx context.Context, collectionID int, page int) (dto.RecipePage, error) {
	recipes, lastPageNum, err := u.repository.GetCollectionByID(ctx, collectionID, page)
	if err != nil {
		return dto.RecipePage{}, err
	}

	uID, err := utils.GetUserIDFromContext(ctx)
	if uID == 0 {
		return dto.RecipePage{Recipes: models.ConvertRecipeToDto(recipes), LastPageNum: lastPageNum}, nil
	}

	favoriteIDsSet, err := u.favoriteRecipesRepository.GetAllIDFavoriteRecipes(ctx, uID)
	if err != nil {
		return dto.RecipePage{}, err
	}

	recipeDTO := models.ConvertRecipeToDto(recipes)

	for i := 0; i < len(recipeDTO); i++ {
		_, ok := favoriteIDsSet[recipeDTO[i].ID]
		if ok {
			recipeDTO[i].IsFavorite = true
		}
	}

	return dto.RecipePage{Recipes: recipeDTO, LastPageNum: lastPageNum}, nil
}

func (u *MainPageUsecase) GetAllCollections(ctx context.Context) ([]dto.Collection, error) {
	collections, err := u.repository.GetAllCollections(ctx)
	if err != nil {
		return []dto.Collection{}, err
	}

	return models.ConvertCollectionToDTO(collections), nil
}
