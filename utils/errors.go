package utils

import (
	"errors"
	"fmt"
)

var (
	ErrNoFound                        = errors.New("results not found")
	ErrFailToSearch                   = errors.New("failed to search")
	ErrFailToGetSuggest               = errors.New("failed to get suggest")
	ErrFailToGetRecipes               = fmt.Errorf("failed to get recipes")
	ErrFailToGetRecipeByID            = fmt.Errorf("failed to get recipe by id")
	ErrFailToGetIngredientsRecipeByID = fmt.Errorf("failed to get ingredients recipe by id")
	ErrFailToEndCooking               = fmt.Errorf("failed to end cooking")
	ErrNoCurrentRecipe                = fmt.Errorf("no current recipe was found")
	ErrFailToStartCooking             = fmt.Errorf("failed to start cooking")
	ErrUserAlreadyCooking             = fmt.Errorf("user already cooking")
	ErrNoSuchRecipeWithID             = fmt.Errorf("no such recipe with id")
	ErrFailedToGetCurrentRecipe       = fmt.Errorf("failed to get current recipe")
	ErrFailedToUpdateRecipeStep       = fmt.Errorf("failed to update step cooking")
	ErrFailedToGetCurrentStepCooking  = fmt.Errorf("failed to get step cooking")
	ErrFailedToGetPrevStep            = fmt.Errorf("failed to get prev step")
	ErrFailedToGetNextStep            = fmt.Errorf("failed to get next step")
	ErrFailedToAddTimer               = fmt.Errorf("failed to add timer to recipe")
	ErrFailedToDeleteTimer            = fmt.Errorf("failed to delete timer")
	ErrFailedToGetTimers              = fmt.Errorf("failed to get timers")
	ErrFailedToGetRecipeStep          = fmt.Errorf("failed to get recipe step")
	ErrGetMaxMinCookingTime           = fmt.Errorf("err get max min cooking time")
	ErrToGetFilterValues              = fmt.Errorf("err to get filter values")
)
