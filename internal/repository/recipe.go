package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Olegsandrik/Exponenta/internal/adapters/postgres"
	"github.com/Olegsandrik/Exponenta/internal/repository/dao"
	"github.com/Olegsandrik/Exponenta/internal/repository/errors"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/logger"

	"github.com/jmoiron/sqlx"
)

type CookingRecipeRepo struct {
	storage *postgres.Adapter
}

func NewCookingRecipeRepo(storage *postgres.Adapter) *CookingRecipeRepo {
	return &CookingRecipeRepo{
		storage: storage,
	}
}

func (repo *CookingRecipeRepo) GetAllRecipe(ctx context.Context, num int) ([]models.RecipeModel, error) {
	q := `SELECT id, name, description, image FROM public.recipes LIMIT $1`

	recipeRows := make([]dao.RecipeTable, 0, num)

	err := repo.storage.Select(ctx, &recipeRows, q, num)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error getting recipe rows: %s with num: %d", err.Error(), num))
		return nil, errors.ErrFailToGetRecipes
	}

	recipeItems := dao.ConvertDaoToRecipe(recipeRows)

	if len(recipeItems) == 0 {
		logger.Error(ctx, fmt.Sprintf("error getting recipe zero row with num: %d", num))
		return nil, errors.ErrFailToGetRecipes
	}

	logger.Info(ctx, fmt.Sprintf("select %d recipes", len(recipeRows)))

	return recipeItems, nil
}

func (repo *CookingRecipeRepo) GetRecipeByID(ctx context.Context, id int) ([]models.RecipeModel, error) {
	tx, err := repo.storage.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed to begin transaction on getting recipe: %d, err: %e", id, err),
		)
		return nil, errors.ErrFailToGetRecipeByID
	}

	defer func() {
		if err != nil {
			if err = tx.Rollback(); err != nil {
				logger.Error(ctx, fmt.Sprintf(
					"failed to rollback transaction on getting recipe: %d, err: %e",
					id,
					err))
			}
		}
	}()

	recipe, err := repo.getRecipeByID(ctx, tx, id)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get recipe: %d, err: %e", id, err))
		return nil, err
	}

	ingredientsRows, err := repo.getIngredientsRecipeByID(ctx, tx, id)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get ingredients recipe: %d, err: %e", id, err))
		return nil, err
	}

	recipe[0].Ingredients = ingredientsRows

	return recipe, nil
}

func (repo *CookingRecipeRepo) getIngredientsRecipeByID(ctx context.Context, tx *sqlx.Tx, id int) ([]byte, error) {
	q := `SELECT ri.ingredient_id, i.name, i.image, ri.amount, ri.unit FROM public.recipe_ingredients AS ri
		  LEFT JOIN public.ingredients as i ON ri.ingredient_id = i.id
		  WHERE ri.recipe_id = $1`

	var ingredientRows []dao.IngredientTable

	err := tx.SelectContext(ctx, &ingredientRows, q, id)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error getting ingredients rows: %s with id: %d", err.Error(), id))
		return nil, errors.ErrFailToGetIngredientsRecipeByID
	}

	if len(ingredientRows) == 0 {
		logger.Error(ctx, fmt.Sprintf("error getting ingredients rows: no rows with id: %d", id))
		return nil, errors.ErrFailToGetIngredientsRecipeByID
	}

	return json.Marshal(ingredientRows)
}

func (repo *CookingRecipeRepo) getRecipeByID(ctx context.Context, tx *sqlx.Tx, id int) ([]models.RecipeModel, error) {
	q := `SELECT r.name, r.description, r.image, r.steps, r.healthscore, r.dish_types, r.diets, r.servings 
			FROM public.recipes as r WHERE id = $1`

	recipeRows := make([]dao.RecipeTable, 0, 1)

	err := tx.SelectContext(ctx, &recipeRows, q, id)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error getting recipe row: %s with id: %d", err.Error(), id))
		return []models.RecipeModel{}, errors.ErrFailToGetRecipeByID
	}

	recipeItem := dao.ConvertDaoToRecipe(recipeRows)

	if len(recipeItem) == 0 {
		logger.Error(ctx, fmt.Sprintf("getting zero recipe row with id: %d", id))
		return []models.RecipeModel{}, errors.ErrFailToGetRecipeByID
	}

	logger.Info(ctx, fmt.Sprintf("select recipe with id: %d", id))

	return recipeItem, nil
}

func (repo *CookingRecipeRepo) EndCooking(ctx context.Context, uID uint) error {
	q := "DELETE FROM public.current_recipe WHERE user_id = $1"

	result, err := repo.storage.Exec(ctx, q, uID)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error deleting recipe row: %e with id: %d", err, uID))
		return errors.ErrFailToEndCooking
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get rows affected by delete with userId: %d, err: %e", uID, err))
		return errors.ErrFailToEndCooking
	}

	if rowsAffected == 0 {
		logger.Error(ctx, fmt.Sprintf("recipe not delete with userId: %d", uID))
		return errors.ErrNoCurrentRecipe
	}

	logger.Info(ctx, fmt.Sprintf("delete row into current_recipe with userId %d", uID))

	return nil
}

func (repo *CookingRecipeRepo) StartCooking(ctx context.Context, uID uint, recipeID int) error {
	tx, err := repo.storage.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed to begin transaction on start cooking for userId: %d, err: %e", uID, err),
		)
		return errors.ErrFailToStartCooking
	}

	defer func() {
		if err != nil {
			if err = tx.Rollback(); err != nil {
				logger.Error(ctx, fmt.Sprintf("Failed to rollback transaction: %e for userId: %d", err, uID))
			}
		}
	}()

	recipe, err := repo.getRecipe(ctx, tx, recipeID)

	if err != nil {
		return err
	}

	if err = repo.insertCurrentRecipe(ctx, tx, uID, recipeID, recipe.Name, recipe.TotalSteps); err != nil {
		return err
	}

	if err = repo.insertRecipeSteps(ctx, tx, uID, recipeID, recipe.Steps); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	logger.Info(ctx,
		fmt.Sprintf("successfully started cooking recipe for userId: %d, recipeId: %d",
			uID,
			recipeID),
	)

	return nil
}

func (repo *CookingRecipeRepo) getRecipe(ctx context.Context, tx *sqlx.Tx, recipeID int) (*dao.RecipeTable, error) {
	q := `SELECT r.name, r.steps, r.total_steps FROM public.recipes as r WHERE id = $1`

	recipeRows := make([]dao.RecipeTable, 0, 1)

	if err := tx.SelectContext(ctx, &recipeRows, q, recipeID); err != nil {
		logger.Error(ctx, fmt.Sprintf("error getting recipe row: %e with recipeId: %d", err, recipeID))
		return nil, errors.ErrFailToGetRecipeByID
	}

	if len(recipeRows) == 0 {
		logger.Error(ctx, fmt.Sprintf("recipe not found with recipeId: %d", recipeID))
		return nil, errors.ErrNoSuchRecipeWithID
	}

	return &recipeRows[0], nil
}

func (repo *CookingRecipeRepo) insertCurrentRecipe(
	ctx context.Context, tx *sqlx.Tx, uID uint, recipeID int, name string, totalSteps int) error {
	q := "INSERT INTO public.current_recipe (user_id, recipe_id, name, total_steps) VALUES ($1, $2, $3, $4)"

	result, err := tx.ExecContext(ctx, q, uID, recipeID, name, totalSteps)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed insert recipe row: %d for userId: %d, recipeId: %d",
				err,
				uID,
				recipeID),
		)
		return errors.ErrUserAlreadyCooking
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("err with get rows affected by insert: %d for userId: %d, recipeId: %d",
				err,
				uID,
				recipeID),
		)
		return errors.ErrFailToStartCooking
	}

	logger.Info(ctx, fmt.Sprintf("inserted %d rows into current_recipe", rowsAffected))

	return nil
}

func (repo *CookingRecipeRepo) insertRecipeSteps(
	ctx context.Context, tx *sqlx.Tx, uID uint, recipeID int, stepsJSON string) error {
	var steps []dao.CurrentRecipeStepTable

	if err := json.Unmarshal([]byte(stepsJSON), &steps); err != nil {
		logger.Error(ctx, fmt.Sprintf("unmarshal error %s with recipe %d", err, recipeID))
		return errors.ErrFailToStartCooking
	}

	q := "INSERT INTO public.current_recipe_step " +
		"(user_id, recipe_id, step_num, step, ingredients, equipment, length) VALUES "

	args := make([]interface{}, 0, 7*len(steps))

	for i, step := range steps {
		if string(step.Length) == "" {
			step.Length = json.RawMessage("null")
		}

		if i > 0 {
			q += ", "
		}

		q += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			7*i+1, 7*i+2, 7*i+3, 7*i+4, 7*i+5, 7*i+6, 7*i+7)

		args = append(args, uID, recipeID, step.NumStep, step.Step, step.Ingredients, step.Equipment, step.Length)
	}

	result, err := tx.ExecContext(ctx, q, args...)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed to insert steps: %e for userId: %d, recipeId: %d",
				err,
				uID,
				recipeID),
		)
		return errors.ErrFailToStartCooking
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("err with get rows affected by insert: %d for userId: %d", err, uID))
		return errors.ErrFailToStartCooking
	}

	if rowsAffected == 0 {
		logger.Error(ctx, fmt.Sprintf("failed to insert steps for userId: %d, recipeId: %d", uID, recipeID))
		return errors.ErrFailToStartCooking
	}

	logger.Info(ctx, fmt.Sprintf("Inserted %d rows into current_recipe_step", rowsAffected))

	return nil
}

func (repo *CookingRecipeRepo) GetCurrentRecipe(ctx context.Context, uID uint) (models.CurrentRecipeModel, error) {
	q := "SELECT cr.recipe_id, cr.name, cr.total_steps, cs.step_num, cs.step, cs.ingredients, cs.equipment, cs.length " +
		"FROM public.current_recipe as cr " +
		"LEFT JOIN public.current_recipe_step as cs ON cs.user_id = cr.user_id AND cr.current_step_num=cs.step_num " +
		"WHERE cr.user_id = $1"

	recipeRows := make([]dao.CurrentRecipeTable, 0, 1)

	err := repo.storage.Select(ctx, &recipeRows, q, uID)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("error getting current recipe row: %s with recipeId: %d",
				err.Error(),
				uID),
		)
		return models.CurrentRecipeModel{}, errors.ErrFailedToGetCurrentRecipe
	}

	if len(recipeRows) == 0 {
		logger.Error(ctx, fmt.Sprintf("recipe not found with userId: %d", uID))
		return models.CurrentRecipeModel{}, errors.ErrFailedToGetCurrentRecipe
	}

	currentRecipeItem := dao.ConvertDaoToCurrentRecipe(recipeRows[0])

	logger.Info(ctx, fmt.Sprintf("got current recipe row for userId: %d", uID))

	return currentRecipeItem, nil
}

func (repo *CookingRecipeRepo) updateCurrentStepTx(ctx context.Context, tx *sqlx.Tx, uID uint, sign string) error {
	q := fmt.Sprintf("UPDATE public.current_recipe SET current_step_num = current_step_num %s 1 "+
		"WHERE user_id = $1", sign)

	result, err := tx.ExecContext(ctx, q, uID)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to update current_step: %e for userId: %d", err, uID))
		return errors.ErrFailedToUpdateRecipeStep
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("err with get rows affected by update: %d for userId: %d", err, uID))
		return errors.ErrFailedToUpdateRecipeStep
	}

	if rowsAffected == 0 {
		logger.Error(ctx, fmt.Sprintf("now found row to update current_step for userId: %d", uID))
		return errors.ErrFailedToUpdateRecipeStep
	}

	return nil
}

func (repo *CookingRecipeRepo) getCurrentStep(
	ctx context.Context, queryer sqlx.QueryerContext, uID uint) (models.CurrentStepRecipeModel, error) {
	currentStep := make([]dao.CurrentRecipeStepTable, 0, 1)

	q := "SELECT cs.step_num, cs.step, cs.ingredients, cs.equipment, cs.length " +
		"FROM public.current_recipe as cr LEFT JOIN public.current_recipe_step as cs " +
		"ON cr.current_step_num = cs.step_num AND cr.user_id=cs.user_id " +
		"WHERE cr.user_id = $1;"

	err := sqlx.SelectContext(ctx, queryer, &currentStep, q, uID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error getting recipe step row: %s with recipeId: %d", err.Error(), uID))
		return models.CurrentStepRecipeModel{}, errors.ErrFailedToGetCurrentStepCooking
	}

	if len(currentStep) == 0 {
		logger.Error(ctx, fmt.Sprintf("recipe step not found with userId: %d", uID))
		return models.CurrentStepRecipeModel{}, errors.ErrFailedToGetCurrentStepCooking
	}

	currentStepModel := dao.ConvertDaoToCurrentStepRecipe(currentStep[0])

	return currentStepModel, nil
}

func (repo *CookingRecipeRepo) GetPrevRecipeStep(ctx context.Context, uID uint) (models.CurrentStepRecipeModel, error) {
	tx, err := repo.storage.BeginTx(ctx, nil)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed to begin transaction on get prev step userId: %d, err: %d", uID, err),
		)
		return models.CurrentStepRecipeModel{}, errors.ErrFailedToGetPrevStep
	}

	defer func() {
		if err != nil {
			if err = tx.Rollback(); err != nil {
				logger.Error(ctx,
					fmt.Sprintf("Failed to rollback transaction on get prev step: %d for userId: %d",
						err,
						uID),
				)
			}
		}
	}()

	err = repo.updateCurrentStepTx(ctx, tx, uID, "-")

	if err != nil {
		return models.CurrentStepRecipeModel{}, err
	}

	currentStepModel, err := repo.getCurrentStep(ctx, tx, uID)

	if err != nil {
		return models.CurrentStepRecipeModel{}, err
	}

	if err = tx.Commit(); err != nil {
		return models.CurrentStepRecipeModel{}, err
	}

	logger.Info(ctx, fmt.Sprintf("get prev current step for userId: %d", uID))

	return currentStepModel, nil
}

func (repo *CookingRecipeRepo) GetNextRecipeStep(ctx context.Context, uID uint) (models.CurrentStepRecipeModel, error) {
	tx, err := repo.storage.BeginTx(ctx, nil)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed to begin transaction on get next step userId: %d, err: %d", uID, err),
		)
		return models.CurrentStepRecipeModel{}, errors.ErrFailedToGetNextStep
	}

	defer func() {
		if err != nil {
			if err = tx.Rollback(); err != nil {
				logger.Error(ctx,
					fmt.Sprintf("failed to rollback transaction on get next step: %d for userId: %d",
						err,
						uID),
				)
			}
		}
	}()

	err = repo.updateCurrentStepTx(ctx, tx, uID, "+")

	if err != nil {
		return models.CurrentStepRecipeModel{}, err
	}

	currentStepModel, err := repo.getCurrentStep(ctx, tx, uID)

	if err != nil {
		return models.CurrentStepRecipeModel{}, err
	}

	if err = tx.Commit(); err != nil {
		return models.CurrentStepRecipeModel{}, err
	}

	logger.Info(ctx, fmt.Sprintf("get next current step for userId: %d", uID))

	return currentStepModel, nil
}

func (repo *CookingRecipeRepo) GetCurrentStep(ctx context.Context, uID uint) (models.CurrentStepRecipeModel, error) {
	return repo.getCurrentStep(ctx, repo.storage, uID)
}

func (repo *CookingRecipeRepo) AddTimerToRecipe(
	ctx context.Context, uID uint, StepNum int, timeSec int, description string) error {
	q := "INSERT INTO public.timers (user_id,step_num,description,end_time) " +
		"VALUES($1, $2, $3, $4);"

	endTime := time.Now().Add(time.Duration(timeSec) * time.Second)

	result, err := repo.storage.Exec(ctx, q, uID, StepNum, description, endTime)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"failed to insert timer for userId: %d, step: %d, description: %s, endTime: %s, err: %s",
			uID,
			StepNum,
			description,
			endTime,
			err))
		return errors.ErrFailedToAddTimer
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"err with get rows affected by insert: for userId: %d, "+
				"step: %d, description: %s, endTime: %s, err: %s",
			uID,
			StepNum,
			description,
			endTime,
			err),
		)
		return errors.ErrFailedToAddTimer
	}

	if rowsAffected == 0 {
		logger.Error(ctx, fmt.Sprintf(
			"no row affected by insert timer for userId: %d, step: %d, description: %s, endTime: %s, err: %s",
			uID,
			StepNum,
			description,
			endTime,
			err),
		)
		return errors.ErrFailedToAddTimer
	}

	return nil
}

func (repo *CookingRecipeRepo) DeleteTimerFromRecipe(ctx context.Context, uID uint, StepNum int) error {
	q := "DELETE FROM public.timers WHERE user_id=$1 AND step_num=$2"

	result, err := repo.storage.Exec(ctx, q, uID, StepNum)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"failed to delete timer for userId: %d, step: %d, err: %s",
			uID,
			StepNum,
			err),
		)
		return errors.ErrFailedToDeleteTimer
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("err with get rows affected by delete: for userId: %d, step: %d",
			uID,
			StepNum),
		)
		return errors.ErrFailedToDeleteTimer
	}

	if rowsAffected == 0 {
		logger.Error(ctx, fmt.Sprintf("no row affected by delete: for userId: %d, step: %d",
			uID,
			StepNum),
		)
		return errors.ErrFailedToDeleteTimer
	}

	return nil
}

func (repo *CookingRecipeRepo) GetTimersRecipe(ctx context.Context, uID uint) ([]models.TimerRecipeModel, error) {
	q := "SELECT description, end_time, step_num FROM public.timers WHERE user_id=$1"

	var timers []dao.TimerTable

	err := repo.storage.Select(ctx, &timers, q, uID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error getting timers: %s with userId: %d", err.Error(), uID))
		return []models.TimerRecipeModel{}, errors.ErrFailedToGetTimers
	}

	timersModel, err := dao.ConvertTimerToDAO(timers)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"error converting timers table to models: %s with userId: %d", err.Error(), uID))
		return []models.TimerRecipeModel{}, errors.ErrFailedToGetTimers
	}

	logger.Info(ctx, fmt.Sprintf("get timers for userId: %d", uID))

	return timersModel, nil
}

func (repo *CookingRecipeRepo) GetCurrentRecipeStepByNum(
	ctx context.Context, uID uint, stepNum int) (models.CurrentStepRecipeModel, error) {
	q := "SELECT step, step_num, ingredients, equipment, length FROM public.current_recipe_step " +
		"WHERE user_id=$1 AND step_num=$2"

	recipeStepRow := make([]dao.CurrentRecipeStepTable, 0, 1)

	err := repo.storage.Select(ctx, &recipeStepRow, q, uID, stepNum)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"error getting recipe step: %s with userId: %d, stepNum: %d",
			err.Error(),
			uID,
			stepNum),
		)
		return models.CurrentStepRecipeModel{}, errors.ErrFailedToGetRecipeStep
	}

	if len(recipeStepRow) == 0 {
		logger.Error(ctx, fmt.Sprintf("not found recipe step with userId: %d, stepNum: %d", uID, stepNum))
		return models.CurrentStepRecipeModel{}, errors.ErrFailedToGetRecipeStep
	}

	logger.Info(ctx, fmt.Sprintf("get recipe step for userId: %d, stepNum: %d", uID, stepNum))

	recipeStep := dao.ConvertDaoToCurrentStepRecipe(recipeStepRow[0])
	return recipeStep, nil
}
