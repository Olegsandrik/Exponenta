package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	DB "github.com/Olegsandrik/Exponenta/internal/adapters/postgres"
	"github.com/Olegsandrik/Exponenta/internal/repository/dao"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/logger"
	"github.com/jmoiron/sqlx"
)

type CookingRecipeRepo struct {
	storage *DB.Adapter
}

func NewCookingRecipeRepo(storage *DB.Adapter) *CookingRecipeRepo {
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
		return nil, fmt.Errorf("failed to get recipes")
	}

	recipeItems := dao.ConvertDaoToRecipe(recipeRows)

	if len(recipeItems) == 0 {
		logger.Error(ctx, fmt.Sprintf("error getting recipe zero row with num: %d", num))
		return nil, fmt.Errorf("failed to get recipes")
	}

	logger.Info(ctx, fmt.Sprintf("select %d recipes", len(recipeRows)))

	return recipeItems, nil
}

func (repo *CookingRecipeRepo) GetRecipeByID(ctx context.Context, id int) ([]models.RecipeModel, error) {
	q := `SELECT r.name, r.description, r.image, r.steps FROM public.recipes as r WHERE id = $1`

	recipeRows := make([]dao.RecipeTable, 0, 1)

	err := repo.storage.Select(ctx, &recipeRows, q, id)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error getting recipe row: %s with id: %d", err.Error(), id))
		return []models.RecipeModel{}, fmt.Errorf("failed to get recipe by id")
	}

	recipeItem := dao.ConvertDaoToRecipe(recipeRows)

	if len(recipeItem) == 0 {
		logger.Error(ctx, fmt.Sprintf("getting zero recipe row with id: %d", id))
		return []models.RecipeModel{}, fmt.Errorf("failed to get recipe by id: %d", id)
	}

	logger.Info(ctx, fmt.Sprintf("select recipe with id: %d", id))

	return recipeItem, nil
}

func (repo *CookingRecipeRepo) EndCooking(ctx context.Context, uID uint) error {
	q := "DELETE FROM public.current_recipe WHERE user_id = $1"

	result, err := repo.storage.Exec(ctx, q, uID)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error deleting recipe row: %e with id: %d", err, uID))
		return fmt.Errorf("failed to end cooking")
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get rows affected by delete with userId: %d, err: %e", uID, err))
		return fmt.Errorf("failed to end cooking")
	}

	if rowsAffected == 0 {
		logger.Error(ctx, fmt.Sprintf("recipe not delete with userId: %d", uID))
		return fmt.Errorf("no current recipe was found")
	}

	logger.Info(ctx, fmt.Sprintf("delete row into current_recipe with userId %d", uID))

	return nil
}

func (repo *CookingRecipeRepo) StartCooking(ctx context.Context, uID uint, recipeID int) error {
	tx, err := repo.storage.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed to begin transaction on start cooking for userId: %d, err: %d", uID, err),
		)
		return fmt.Errorf("failed to start cooking")
	}

	defer func() {
		if err != nil {
			if err = tx.Rollback(); err != nil {
				logger.Error(ctx, fmt.Sprintf("Failed to rollback transaction: %d for userId: %d", err, uID))
			}
		}
	}()

	recipe, err := repo.getRecipe(ctx, tx, recipeID)

	if err != nil {
		return err
	}

	if err = repo.insertCurrentRecipe(ctx, tx, uID, recipeID, recipe.Name); err != nil {
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
	q := `SELECT r.name, r.steps FROM public.recipes as r WHERE id = $1`

	recipeRows := make([]dao.RecipeTable, 0, 1)

	if err := tx.SelectContext(ctx, &recipeRows, q, recipeID); err != nil {
		logger.Error(ctx, fmt.Sprintf("error getting recipe row: %e with recipeId: %d", err, recipeID))
		return nil, fmt.Errorf("failed to get recipe")
	}

	if len(recipeRows) == 0 {
		logger.Error(ctx, fmt.Sprintf("recipe not found with recipeId: %d", recipeID))
		return nil, fmt.Errorf("no such recipe with id")
	}

	return &recipeRows[0], nil
}

func (repo *CookingRecipeRepo) insertCurrentRecipe(
	ctx context.Context, tx *sqlx.Tx, uID uint, recipeID int, name string) error {
	q := "INSERT INTO public.current_recipe (user_id, recipe_id, name) VALUES ($1, $2, $3)"

	result, err := tx.ExecContext(ctx, q, uID, recipeID, name)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed insert recipe row: %d for userId: %d, recipeId: %d",
				err,
				uID,
				recipeID),
		)
		return fmt.Errorf("user already cooking")
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("err with get rows affected by insert: %d for userId: %d, recipeId: %d",
				err,
				uID,
				recipeID),
		)
		return fmt.Errorf("failed to start cooking")
	}

	logger.Info(ctx, fmt.Sprintf("inserted %d rows into current_recipe", rowsAffected))

	return nil
}

func (repo *CookingRecipeRepo) insertRecipeSteps(
	ctx context.Context, tx *sqlx.Tx, uID uint, recipeID int, stepsJSON string) error {
	var steps []dao.CurrentStepRecipeTable

	if err := json.Unmarshal([]byte(stepsJSON), &steps); err != nil {
		logger.Error(ctx, fmt.Sprintf("unmarshal error %s with recipe %d", err, recipeID))
		return fmt.Errorf("failed to start cooking")
	}

	q := "INSERT INTO public.current_recipe_step " +
		"(user_id, recipe_id, step_num, step, ingredients, equipment, length) VALUES "

	args := make([]interface{}, 0, 7*len(steps))

	for i, step := range steps {
		if string(step.Length) == "" {
			step.Length = json.RawMessage("[]")
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
		return fmt.Errorf("failed to start cooking")
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("err with get rows affected by insert: %d for userId: %d", err, uID))
		return fmt.Errorf("failed to start cooking")
	}

	if rowsAffected == 0 {
		logger.Error(ctx, fmt.Sprintf("failed to insert steps for userId: %d, recipeId: %d", uID, recipeID))
		return fmt.Errorf("failed to start cooking")
	}

	logger.Info(ctx, fmt.Sprintf("Inserted %d rows into current_recipe_step", rowsAffected))

	return nil
}

func (repo *CookingRecipeRepo) GetCurrentRecipe(ctx context.Context, uID uint) (models.CurrentRecipeModel, error) {
	q := "SELECT cr.recipe_id, cr.name, cs.step_num, cs.step, cs.ingredients, cs.equipment, cs.length " +
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
		return models.CurrentRecipeModel{}, fmt.Errorf("failed to get current recipe")
	}

	if len(recipeRows) == 0 {
		logger.Error(ctx, fmt.Sprintf("recipe not found with userId: %d", uID))
		return models.CurrentRecipeModel{}, fmt.Errorf("failed to get current recipe")
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
		return fmt.Errorf("failed to update step cooking")
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("err with get rows affected by update: %d for userId: %d", err, uID))
		return fmt.Errorf("failed to update step cooking")
	}

	if rowsAffected == 0 {
		logger.Error(ctx, fmt.Sprintf("now found row to update current_step for userId: %d", uID))
		return fmt.Errorf("failed to update step cooking")
	}

	return nil
}

func (repo *CookingRecipeRepo) getCurrentStep(
	ctx context.Context, queryer sqlx.QueryerContext, uID uint) (models.CurrentStepRecipeModel, error) {
	CurrentStep := make([]dao.CurrentStepRecipeTable, 0, 1)

	q := "SELECT cs.step_num, cs.step, cs.ingredients, cs.equipment, cs.length " +
		"FROM public.current_recipe as cr LEFT JOIN public.current_recipe_step as cs " +
		"ON cr.current_step_num = cs.step_num AND cr.user_id=cs.user_id " +
		"WHERE cr.user_id = $1;"

	err := sqlx.SelectContext(ctx, queryer, &CurrentStep, q, uID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error getting recipe step row: %s with recipeId: %d", err.Error(), uID))
		return models.CurrentStepRecipeModel{}, fmt.Errorf("failed to get step cooking")
	}

	if len(CurrentStep) == 0 {
		logger.Error(ctx, fmt.Sprintf("recipe step not found with userId: %d", uID))
		return models.CurrentStepRecipeModel{}, fmt.Errorf("failed to get step cooking")
	}

	currentStepModel := dao.ConvertDaoToCurrentStepRecipe(CurrentStep[0])

	return currentStepModel, nil
}

func (repo *CookingRecipeRepo) GetPrevRecipeStep(ctx context.Context, uID uint) (models.CurrentStepRecipeModel, error) {
	tx, err := repo.storage.BeginTx(ctx, nil)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed to begin transaction on get prev step userId: %d, err: %d", uID, err),
		)
		return models.CurrentStepRecipeModel{}, fmt.Errorf("failed to get prev step")
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
		return models.CurrentStepRecipeModel{}, fmt.Errorf("failed to get next step")
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
		return fmt.Errorf("failed to add timer to recipe")
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
		return fmt.Errorf("failed to add timer to recipe")
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
		return fmt.Errorf("failed to add timer to recipe")
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
		return fmt.Errorf("failed to delete timer")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("err with get rows affected by delete: for userId: %d, step: %d",
			uID,
			StepNum),
		)
		return fmt.Errorf("failed to delete timer")
	}

	if rowsAffected == 0 {
		logger.Error(ctx, fmt.Sprintf("no row affected by delete: for userId: %d, step: %d",
			uID,
			StepNum),
		)
		return fmt.Errorf("failed to delete timer")
	}

	return nil
}

func (repo *CookingRecipeRepo) GetTimersRecipe(ctx context.Context, uID uint) ([]models.TimerRecipeModel, error) {
	q := "SELECT description, end_time FROM public.timers WHERE user_id=$1"

	var timers []dao.TimerTable

	err := repo.storage.Select(ctx, &timers, q, uID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error getting timers: %s with userId: %d", err.Error(), uID))
		return []models.TimerRecipeModel{}, fmt.Errorf("failed to get timers")
	}

	timersModel, err := dao.ConvertTimerToDAO(timers)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"error converting timers table to models: %s with userId: %d", err.Error(), uID))
		return []models.TimerRecipeModel{}, fmt.Errorf("failed to get timers")
	}

	logger.Info(ctx, fmt.Sprintf("get timers for userId: %d", uID))

	return timersModel, nil
}
