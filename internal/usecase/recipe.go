package usecase

import (
	"context"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/internal/utils"
)

type CookingRecipeRepo interface {
	GetAllRecipe(ctx context.Context, num int) ([]models.RecipeModel, error)
	GetRecipeByID(ctx context.Context, id int) ([]models.RecipeModel, error)
	EndCooking(ctx context.Context, uID uint) (int, bool, error)
	StartCooking(ctx context.Context, uID uint, recipeID int, isGenerated bool) error
	GetCurrentRecipe(ctx context.Context, uID uint) (models.CurrentRecipeModel, error)
	GetNextRecipeStep(ctx context.Context, uID uint) (models.CurrentStepRecipeModel, error)
	GetPrevRecipeStep(ctx context.Context, uID uint) (models.CurrentStepRecipeModel, error)
	GetCurrentStep(ctx context.Context, uID uint) (models.CurrentStepRecipeModel, error)
	AddTimerToRecipe(ctx context.Context, uID uint, StepNum int, timeSec int, description string) error
	DeleteTimerFromRecipe(ctx context.Context, uID uint, StepNum int) error
	GetTimersRecipe(ctx context.Context, uID uint) ([]models.TimerRecipeModel, error)
	GetCurrentRecipeStepByNum(ctx context.Context, uID uint, stepNum int) (
		models.CurrentStepRecipeModel, error,
	)
	AddRecipeToHistory(ctx context.Context, userID uint, recipeID int, isGenerated bool) error
}

type CookingRecipeUsecase struct {
	repo                CookingRecipeRepo
	favoriteRecipesRepo FavoriteRecipesRepo
}

func NewCookingRecipeUsecase(repo CookingRecipeRepo, favoriteRecipesRepo FavoriteRecipesRepo) *CookingRecipeUsecase {
	return &CookingRecipeUsecase{
		repo:                repo,
		favoriteRecipesRepo: favoriteRecipesRepo,
	}
}

func (u *CookingRecipeUsecase) GetAllRecipe(ctx context.Context, num int) ([]dto.RecipeDto, error) {
	recipeModels, err := u.repo.GetAllRecipe(ctx, num)

	if err != nil {
		return nil, err
	}

	utils.SanitizeRecipeDescription(recipeModels)

	recipeDTO := models.ConvertRecipeToDto(recipeModels)

	uID, _ := utils.GetUserIDFromContext(ctx)
	if uID == 0 {
		return recipeDTO, nil
	}

	favoriteIDsSet, err := u.favoriteRecipesRepo.GetAllIDFavoriteRecipes(ctx, uID)
	if err != nil {
		return []dto.RecipeDto{}, err
	}

	for i := 0; i < len(recipeDTO); i++ {
		_, ok := favoriteIDsSet[recipeDTO[i].ID]
		if ok {
			recipeDTO[i].IsFavorite = true
		}
	}

	return recipeDTO, nil
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

func (u *CookingRecipeUsecase) StartCookingRecipe(ctx context.Context, recipeID int) (dto.CurrentStepRecipeDto, error) {
	uID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return dto.CurrentStepRecipeDto{}, err
	}

	err = u.repo.StartCooking(ctx, uID, recipeID, false)
	if err != nil {
		return dto.CurrentStepRecipeDto{}, err
	}

	currentRecipeStepModel, err := u.repo.GetCurrentStep(ctx, uID)
	if err != nil {
		return dto.CurrentStepRecipeDto{}, err
	}

	currentRecipeStep := models.ConvertCurrentStepToDTO(currentRecipeStepModel)

	return currentRecipeStep, nil
}

func (u *CookingRecipeUsecase) EndCookingRecipe(ctx context.Context) error {
	uID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	recipeID, IsGenerated, err := u.repo.EndCooking(ctx, uID)
	if err != nil {
		return err
	}

	return u.repo.AddRecipeToHistory(ctx, uID, recipeID, IsGenerated)
}

func (u *CookingRecipeUsecase) GetCurrentRecipe(ctx context.Context) (dto.CurrentRecipeDto, error) {
	uID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return dto.CurrentRecipeDto{}, err
	}

	currentRecipe, err := u.repo.GetCurrentRecipe(ctx, uID)
	if err != nil {
		return dto.CurrentRecipeDto{}, err
	}

	currentRecipeDto := models.ConvertCurrentRecipeToDTO(currentRecipe)

	return currentRecipeDto, nil
}

func (u *CookingRecipeUsecase) NextStepRecipe(ctx context.Context) (dto.CurrentStepRecipeDto, error) {
	uID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return dto.CurrentStepRecipeDto{}, err
	}

	nextStep, err := u.repo.GetNextRecipeStep(ctx, uID)
	if err != nil {
		return dto.CurrentStepRecipeDto{}, err
	}

	nextStepDto := models.ConvertCurrentStepToDTO(nextStep)

	return nextStepDto, nil
}

func (u *CookingRecipeUsecase) PreviousStepRecipe(ctx context.Context) (dto.CurrentStepRecipeDto, error) {
	uID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return dto.CurrentStepRecipeDto{}, err
	}

	prevStep, err := u.repo.GetPrevRecipeStep(ctx, uID)
	if err != nil {
		return dto.CurrentStepRecipeDto{}, err
	}

	prevStepDto := models.ConvertCurrentStepToDTO(prevStep)

	return prevStepDto, nil
}

func (u *CookingRecipeUsecase) AddTimerRecipe(ctx context.Context, stepNum int, timeSec int) error {
	uID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	recipeStep, err := u.repo.GetCurrentRecipeStepByNum(ctx, uID, stepNum)
	if err != nil {
		return err
	}

	err = u.repo.AddTimerToRecipe(ctx, uID, stepNum, timeSec, recipeStep.Step)
	if err != nil {
		return err
	}

	return nil
}

func (u *CookingRecipeUsecase) DeleteTimerRecipe(ctx context.Context, stepNum int) error {
	uID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	err = u.repo.DeleteTimerFromRecipe(ctx, uID, stepNum)
	if err != nil {
		return err
	}

	return nil
}

func (u *CookingRecipeUsecase) GetTimersRecipe(ctx context.Context) ([]dto.TimerRecipeDto, error) {
	uID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	timersModels, err := u.repo.GetTimersRecipe(ctx, uID)
	if err != nil {
		return nil, err
	}

	timersDto := models.ConvertTimersToDTO(timersModels)

	return timersDto, nil
}
