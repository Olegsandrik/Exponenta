package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/Olegsandrik/Exponenta/config"
	"github.com/Olegsandrik/Exponenta/internal/adapters/postgres"
	"github.com/Olegsandrik/Exponenta/internal/repository/dao"
	"github.com/Olegsandrik/Exponenta/internal/repository/errors"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/logger"
	"github.com/Olegsandrik/Exponenta/utils"
)

const (
	promptChoiceGeneration    = "Gen"
	promptChoiceModernization = "Mod"
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

func (repo *GeneratedRecipeRepo) GetAllRecipes(ctx context.Context, num int,
	userID uint) ([]models.RecipeModel, error) {
	q := `SELECT id, name, description, ready_in_minutes FROM public.generated_recipes WHERE user_id = $1 LIMIT $2`

	recipeRows := make([]dao.RecipeTable, 0, num)

	err := repo.storage.Select(ctx, &recipeRows, q, userID, num)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error getting recipe rows: %+v with num: %d", err, num))
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

func (repo *GeneratedRecipeRepo) getRecipeByIDAndVersion(ctx context.Context, recipeID int, userID uint,
	versionID int) ([]dao.RecipeTable, error) {
	q := `SELECT r.name, r.description, r.ingredients, r.steps, r.dish_types, r.diets, r.servings, r.ready_in_minutes
			FROM public.generated_recipes_versions as r WHERE r.user_id = $1 AND r.id = $2 AND r.version = $3`

	recipeRows := make([]dao.RecipeTable, 0, 1)

	err := repo.storage.Select(ctx, &recipeRows, q, userID, recipeID, versionID)

	if err != nil {
		return []dao.RecipeTable{}, errors.ErrFailToGetRecipeByIDAndVersion
	}

	return recipeRows, nil
}

func (repo *GeneratedRecipeRepo) GetRecipeByID(ctx context.Context, recipeID int,
	userID uint) ([]models.RecipeModel, error) {
	q := `SELECT r.name, r.description, r.ingredients, r.steps, r.dish_types, r.diets, r.servings, r.ready_in_minutes
			FROM public.generated_recipes as r WHERE r.user_id = $1 AND r.id = $2`

	recipeRows := make([]dao.RecipeTable, 0, 1)

	err := repo.storage.Select(ctx, &recipeRows, q, userID, recipeID)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error getting recipe row: %s with rid: %d, uid: %d",
			err.Error(), recipeID, userID))
		return []models.RecipeModel{}, errors.ErrFailToGetRecipeByID
	}

	recipeItem := dao.ConvertGenRecipeToRecipeModel(recipeRows)

	if len(recipeItem) == 0 {
		logger.Error(ctx, fmt.Sprintf("getting zero recipe row with rid: %d, uid: %d", recipeID, userID))
		return []models.RecipeModel{}, errors.ErrFailToGetRecipeByID
	}

	logger.Info(ctx, fmt.Sprintf("select recipe with id: %d, uid: %d", recipeID, userID))

	return recipeItem, nil
}

func (repo *GeneratedRecipeRepo) GetHistoryByID(ctx context.Context, recipeID int,
	userID uint) ([]models.RecipeModel, error) {
	q := `SELECT r.id, r.version, r.name, r.description, r.steps, r.dish_types, r.diets, r.servings, r.total_steps
		  FROM public.generated_recipes_versions as r WHERE user_id = $1 AND id = $2`

	var recipeRows []dao.RecipeTable

	err := repo.storage.Select(ctx, &recipeRows, q, userID, recipeID)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("err getting history rows: %+v with rid: %d, uid: %d",
			err, recipeID, userID))
		return nil, err
	}

	recipeModels := dao.ConvertDaoToRecipe(recipeRows)

	logger.Info(ctx, fmt.Sprintf("select recipe history with id: %d, uid: %d", recipeID, userID))

	return recipeModels, nil
}

func (repo *GeneratedRecipeRepo) CreateRecipe(ctx context.Context, products []string, query string,
	userID uint) ([]models.RecipeModel, error) {
	APIURL := repo.config.DeepSeekAPIURL
	APIKey := repo.config.DeepSeekAPIKey

	q := fmt.Sprintf("my promise: %s, products: %s", query, strings.Join(products, ", "))

	respData, err := utils.GetResponseData(ctx, q, APIURL, APIKey, promptChoiceGeneration)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed to get resp data: %+v for userId: %d, query: %s, promptChoice: %s",
				err, userID, query, promptChoiceGeneration),
		)
		return nil, errors.ErrWithGenerating
	}

	generatedRecipe, err := dao.ParseGeneratedRecipe(json.RawMessage(respData))

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf(`RESP from DeepSeek API: %s, 
				failed to parse recipe: %+v for userId: %d, query: %s, products: %s`,
				respData, err, userID, query, products),
		)
		return nil, errors.ErrWithGenerating
	}

	tx, err := repo.storage.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed to begin transaction on generating recipe: %d, err: %e", userID, err),
		)
		return nil, errors.ErrWithGenerating
	}

	defer func() {
		if err != nil {
			if err = tx.Rollback(); err != nil {
				logger.Error(ctx, fmt.Sprintf("Failed to rollback transaction: %e for userId: %d",
					err, userID))
			}
		}
	}()

	generateRecipeID, err := repo.insertGeneratedRecipe(ctx, repo.storage, generatedRecipe, userID)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("fail to insert generated recipe: %+v for userId: %d, genRecipe: %+v",
				err, userID, generatedRecipe),
		)
		return nil, errors.ErrWithGenerating
	}

	recipeVersion, err := repo.insertVersionGeneratedRecipe(ctx,
		repo.storage, generatedRecipe, userID, generateRecipeID)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf(
				"fail to insert version generated recipe: %+v for userId: %d, genRecipe: %+v",
				err, userID, generatedRecipe),
		)
		return nil, errors.ErrWithGenerating
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.ErrWithGenerating
	}

	generatedRecipe.ID = generateRecipeID
	generatedRecipe.Version = recipeVersion

	recipeModel := dao.ConvertGeneratedRecipeToRecipeModels([]dao.GeneratedRecipe{generatedRecipe})

	return recipeModel, nil
}

func (repo *GeneratedRecipeRepo) insertVersionGeneratedRecipe(ctx context.Context, queryer sqlx.QueryerContext,
	generatedRecipe dao.GeneratedRecipe, userID uint, generateRecipeID int) (int, error) {
	var generateVersion int
	q := `INSERT INTO public.generated_recipes_versions (
	        user_id, 
	        name,
		    description,
			dish_types,
		    servings, 
	        diets,
			ingredients,
			ready_in_minutes, 
			steps, 
			total_steps,
            id,
            version
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, (
	SELECT COALESCE(MAX(version), 0) + 1 AS next_version
 		FROM public.generated_recipes_versions
 		WHERE id = $12
 	)) RETURNING version;`

	err := queryer.QueryRowxContext(ctx, q,
		userID,
		generatedRecipe.Name,
		generatedRecipe.Desc,
		generatedRecipe.DishTypes,
		generatedRecipe.ServingsNum,
		generatedRecipe.Diets,
		generatedRecipe.Ingredients,
		generatedRecipe.ReadyInMinutes,
		generatedRecipe.Steps,
		generatedRecipe.TotalSteps,
		generateRecipeID,
		generateRecipeID,
	).Scan(&generateVersion)

	if err != nil {
		return 0, err
	}

	return generateVersion, err
}

func (repo *GeneratedRecipeRepo) insertGeneratedRecipe(ctx context.Context, queryer sqlx.QueryerContext,
	generatedRecipe dao.GeneratedRecipe, userID uint) (int, error) {
	var generateRecipeID int
	q := `INSERT INTO public.generated_recipes (
    user_id,name,description,dish_types,servings, diets, ingredients, ready_in_minutes, steps, total_steps)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`

	err := queryer.QueryRowxContext(ctx, q,
		userID,
		generatedRecipe.Name,
		generatedRecipe.Desc,
		generatedRecipe.DishTypes,
		generatedRecipe.ServingsNum,
		generatedRecipe.Diets,
		generatedRecipe.Ingredients,
		generatedRecipe.ReadyInMinutes,
		generatedRecipe.Steps,
		generatedRecipe.TotalSteps).Scan(&generateRecipeID)

	if err != nil {
		return 0, err
	}

	return generateRecipeID, nil
}

func (repo *GeneratedRecipeRepo) UpdateRecipe(ctx context.Context, query string, recipeID int, versionID int,
	userID uint) ([]models.RecipeModel, error) {
	APIURL := repo.config.DeepSeekAPIURL
	APIKey := repo.config.DeepSeekAPIKey

	recipeDao, err := repo.getRecipeByIDAndVersion(ctx, recipeID, userID, versionID)

	if err != nil {
		return nil, err
	}

	if len(recipeDao) == 0 {
		return nil, errors.ErrWithGenerating
	}

	jsonRecipe, err := json.Marshal(recipeDao[0])

	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf(`my recipe in correct json format (only change cookingTime to readyInMinutes) 
	and add totalSteps - int as count of steps %s reformat: %s`, string(jsonRecipe), query)

	/*
		q := `my recipe in correct json format (only change cookingTime to readyInMinutes)
		and add totalSteps - int as count of steps` + string(jsonRecipe) + "reformat: " + query
	*/

	respData, err := utils.GetResponseData(ctx, q, APIURL, APIKey, promptChoiceModernization)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf(`failed get resp data: %+v for userId: %d, query: %s, recipeID: %d, respData: %+v,
				prompt: %s`,
				err, userID, query, recipeID, respData, promptChoiceModernization),
		)
		return nil, errors.ErrWithModernization
	}

	generatedRecipe, err := dao.ParseGeneratedRecipe(json.RawMessage(respData))

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf(`failed to parse recipe: %+v for respData: %+v,`,
				err, respData),
		)
		return nil, errors.ErrWithModernization
	}

	generateVersion, err := repo.insertVersionGeneratedRecipe(ctx, repo.storage, generatedRecipe, userID, recipeID)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf(`failed to insert version: %+v for genRecipe: %+v, uID: %d, recipeID: %d`,
				err, generatedRecipe, userID, recipeID),
		)
		return nil, errors.ErrWithModernization
	}

	generatedRecipe.Version = generateVersion

	recipeModel := dao.ConvertGeneratedRecipeToRecipeModels([]dao.GeneratedRecipe{generatedRecipe})

	return recipeModel, nil
}

func (repo *GeneratedRecipeRepo) GetVersionByID(ctx context.Context, userID uint,
	recipeID int, versionID int) ([]dao.RecipeTable, error) {
	q := `SELECT r.name, r.description, r.steps, r.ingredients, 
       	  r.ready_in_minutes ,r.dish_types, r.diets, r.servings, r.total_steps
		  FROM public.generated_recipes_versions as r WHERE user_id = $1 AND id = $2 AND version = $3`

	var recipeRows []dao.RecipeTable

	err := repo.storage.Select(ctx, &recipeRows, q, userID, recipeID, versionID)

	if err != nil {
		return []dao.RecipeTable{}, err
	}
	return recipeRows, nil
}

func (repo *GeneratedRecipeRepo) SetNewMainVersion(ctx context.Context, recipeID int,
	versionID int, userID uint) error {
	recipeRows, err := repo.GetVersionByID(ctx, userID, recipeID, versionID)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf(`failed to get recipe version: %+v for uID: %d, recipeID: %d, versionID: %d`,
				err, userID, recipeID, versionID),
		)
		return errors.ErrWithUpdateVersion
	}

	q := `UPDATE public.generated_recipes SET 
			 name = $1, 
			 description = $2, 
			 dish_types = $3, 
			 servings = $4, 
			 diets = $5, 
			 ingredients = $6,
			 ready_in_minutes = $7,
			 steps = $8,
			 total_steps = $9
		 WHERE user_id = $10 AND id = $11`

	result, err := repo.storage.Exec(ctx, q,
		recipeRows[0].Name,
		recipeRows[0].Desc,
		recipeRows[0].DishTypes,
		recipeRows[0].ServingsNum,
		recipeRows[0].Diets,
		recipeRows[0].Ingredients,
		recipeRows[0].CookingTime,
		recipeRows[0].Steps,
		recipeRows[0].TotalSteps,
		userID,
		recipeID,
	)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf(
				`failed to update recipe version: %+v for uID: %d, recipeID: %d, versionID: %d, rows: %+v`,
				err, userID, recipeID, versionID, recipeRows),
		)
		return errors.ErrWithUpdateVersion
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf(
				`failed to get affected rows: %+v for uID: %d, recipeID: %d, versionID: %d, rows: %+v`,
				err, userID, recipeID, versionID, recipeRows),
		)
		return errors.ErrWithUpdateVersion
	}

	if rowsAffected == 0 {
		logger.Error(ctx,
			fmt.Sprintf(
				`zero affected rows: %+v for uID: %d, recipeID: %d, versionID: %d, rows: %+v`,
				err, userID, recipeID, versionID, recipeRows),
		)
		return errors.ErrWithGenerating
	}

	return nil
}
