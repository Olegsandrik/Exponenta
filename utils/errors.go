package utils

import (
	"errors"
	"fmt"
)

var (
	NoFoundErr                       = errors.New("results not found")
	FailToSearchErr                  = errors.New("failed to search")
	FailToGetSuggestErr              = errors.New("failed to get suggest")
	FailToGetRecipesErr              = fmt.Errorf("failed to get recipes")
	FailToGetRecipeByIDErr           = fmt.Errorf("failed to get recipe by id")
	FailToEndCookingErr              = fmt.Errorf("failed to end cooking")
	NoCurrentRecipeErr               = fmt.Errorf("no current recipe was found")
	FailToStartCookingErr            = fmt.Errorf("failed to start cooking")
	UserAlreadyCookingErr            = fmt.Errorf("user already cooking")
	NoSuchRecipeWithIDErr            = fmt.Errorf("no such recipe with id")
	FailedToGetCurrentRecipeErr      = fmt.Errorf("failed to get current recipe")
	FailedToUpdateRecipeStepErr      = fmt.Errorf("failed to update step cooking")
	FailedToGetCurrentStepCookingErr = fmt.Errorf("failed to get step cooking")
	FailedToGetPrevStepErr           = fmt.Errorf("failed to get prev step")
	FailedToGetNextStepErr           = fmt.Errorf("failed to get next step")
	FailedToAddTimerErr              = fmt.Errorf("failed to add timer to recipe")
	FailedToDeleteTimerErr           = fmt.Errorf("failed to delete timer")
	FailedToGetTimersErr             = fmt.Errorf("failed to get timers")
	FailedToGetRecipeStep            = fmt.Errorf("failed to get recipe step")
)
