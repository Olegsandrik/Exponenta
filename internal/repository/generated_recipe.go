package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Olegsandrik/Exponenta/config"
	"github.com/Olegsandrik/Exponenta/internal/adapters/postgres"
	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/repository/dao"
	"github.com/Olegsandrik/Exponenta/internal/repository/errors"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/logger"
	"github.com/Olegsandrik/Exponenta/utils"
	"github.com/jmoiron/sqlx"
	"net/http"
	"strings"
)

type GeneratedRecipeRepo struct {
	storage *postgres.Adapter
	config  *config.Config
}

func NewGeneratedRecipeRepo(storage *postgres.Adapter, config *config.Config) *GeneratedRecipeRepo {
	return &GeneratedRecipeRepo{
		storage: storage,
		config:  config,
	}
}

func (repo *GeneratedRecipeRepo) GetAllRecipes(ctx context.Context, num int, userID uint) ([]models.RecipeModel, error) {
	q := `SELECT id, name, description FROM public.generated_recipes WHERE user_id = $1 LIMIT $2`

	recipeRows := make([]dao.RecipeTable, 0, num)

	err := repo.storage.Select(ctx, &recipeRows, q, userID, num)

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

func (repo *GeneratedRecipeRepo) GetRecipeByID(ctx context.Context, recipeID int, userID uint) ([]models.RecipeModel, error) {
	q := `SELECT r.name, r.description, r.steps, r.healthscore, r.dish_types, r.diets, r.servings 
			FROM public.generated_recipes as r WHERE r.user_id = $1 AND r.id = $2`

	recipeRows := make([]dao.RecipeTable, 0, 1)

	err := repo.storage.Select(ctx, &recipeRows, q, userID, recipeID)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error getting recipe row: %s with rid: %d, uid: %d",
			err.Error(), recipeID, userID))
		return []models.RecipeModel{}, errors.ErrFailToGetRecipeByID
	}

	recipeItem := dao.ConvertDaoToRecipe(recipeRows)

	if len(recipeItem) == 0 {
		logger.Error(ctx, fmt.Sprintf("getting zero recipe row with rid: %d, uid: %d", recipeID, userID))
		return []models.RecipeModel{}, errors.ErrFailToGetRecipeByID
	}

	logger.Info(ctx, fmt.Sprintf("select recipe with id: %d, uid: %d", recipeID, userID))

	return recipeItem, nil
}

func (repo *GeneratedRecipeRepo) StartCookingByRecipeID(ctx context.Context, recipeID int, uID uint) error {
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

func (repo *GeneratedRecipeRepo) insertCurrentRecipe(
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

func (repo *GeneratedRecipeRepo) insertRecipeSteps(
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

func (repo *GeneratedRecipeRepo) getRecipe(ctx context.Context, tx *sqlx.Tx, recipeID int) (*dao.RecipeTable, error) {
	q := `SELECT r.name, r.steps, r.total_steps FROM public.generated_recipes as r WHERE id = $1`

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

func (repo *GeneratedRecipeRepo) GetHistoryByID(ctx context.Context, recipeID int, userID uint) ([]models.RecipeModel, error) {
	q := `SELECT r.name, r.description, r.steps, r.healthscore, r.dish_types, r.diets, r.servings, r.total_steps
		  FROM public.generated_recipes_versions as r WHERE user_id = $1 AND id = $2`

	var recipeRows []dao.RecipeTable

	err := repo.storage.Select(ctx, &recipeRows, q, userID, recipeID)

	if err != nil {

		return nil, err
	}

	recipeModels := dao.ConvertDaoToRecipe(recipeRows)

	return recipeModels, nil
}

func (repo *GeneratedRecipeRepo) SetNewMainVersion(ctx context.Context, recipeID int, versionID int, userID uint) error {
	q := `SELECT r.name, r.description, r.steps, r.healthscore, r.dish_types, r.diets, r.servings, r.total_steps
		  FROM public.generated_recipes_versions as r WHERE user_id = $1 AND id = $2`

	var recipeRows []dao.RecipeTable

	err := repo.storage.Select(ctx, &recipeRows, q, userID, recipeID)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed 1: %+v for userId: %d, recipeId: %d",
				err,
				userID,
				recipeID),
		)
		return err
	}

	q = `UPDATE public.generated_recipes SET name = $1, 
                                         description = $2, 
                                         steps = $3, 
                                         healthscore = $4, 
                                         dish_types = $5, 
                                         diets = $6, 
                                         servings = $7, 
                                         total_steps = $8
		 WHERE user_id = $9 AND id = $10`

	result, err := repo.storage.Exec(ctx, q,
		recipeRows[0].Name,
		recipeRows[0].Desc,
		recipeRows[0].Steps,
		recipeRows[0].HealthScore,
		recipeRows[0].DishTypes,
		recipeRows[0].Diets,
		recipeRows[0].ServingsNum,
		recipeRows[0].TotalSteps)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed 2: %+v for userId: %d, recipeId: %d",
				err,
				userID,
				recipeID),
		)
		return errors.ErrWithGenerating
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed 3: %+v for userId: %d, recipeId: %d",
				err,
				userID,
				recipeID),
		)
		return errors.ErrWithGenerating
	}

	if rowsAffected == 0 {
		logger.Error(ctx,
			fmt.Sprintf("failed 4: %+v for userId: %d, recipeId: %d",
				err,
				userID,
				recipeID),
		)
		return errors.ErrWithGenerating
	}

	return nil
}

func (repo *GeneratedRecipeRepo) CreateRecipe(ctx context.Context, products []string, query string,
	userID uint) ([]models.RecipeModel, error) {
	APIURL := repo.config.DeepSeekAPIURL
	APIKey := repo.config.DeepSeekAPIKey

	q := query + " products " + strings.Join(products, " ")

	req, err := utils.BuildRequestGeneration(ctx, q, APIURL, APIKey)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed 1: %+v for userId: %d, uId: %d, query: %s, products: %s",
				err,
				userID, query, products),
		)
		return nil, errors.ErrWithGenerating
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed 2: %+v for userId: %d, uId: %d, query: %s, products: %s",
				err,
				userID, query, products),
		)
		return nil, errors.ErrWithGenerating
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error(ctx,
			fmt.Sprintf("failed 3: %+v for userId: %d, uId: %d, query: %s, products: %s",
				err,
				userID, query, products),
		)
		return nil, errors.ErrWithGenerating
	}

	respData, err := dto.ConvertGenerationDataTest(ctx, resp.Body)

	logger.Info(ctx, fmt.Sprintf("RESP: %s", respData))

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed 4: %+v for userId: %d, uId: %d, query: %s, products: %s",
				err,
				userID, query, products),
		)
		return nil, errors.ErrWithGenerating
	}

	recipeTable, err := dao.ParseGeneratedRecipe(json.RawMessage(respData))

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed 5: %+v for userId: %d, uId: %d, query: %s, products: %s",
				err,
				userID, query, products),
		)
		return nil, errors.ErrWithGenerating
	}

	// TODO потом сохранить в двух таблицах и вернуть созданный рецепт и отослать данные

	recipeModel := dao.ConvertDaoToRecipe([]dao.RecipeTable{recipeTable})

	return recipeModel, nil
}

func (repo *GeneratedRecipeRepo) UpdateRecipe(ctx context.Context, query string, recipeID int, versionID int,
	userID uint) ([]models.RecipeModel, error) {
	// отослать прошлые данные в deepseek и после чего получить новые данные
	// обновить в таблицах версиях данные
	// TODO fix
	return nil, nil
}
