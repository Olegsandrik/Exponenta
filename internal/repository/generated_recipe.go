package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/Olegsandrik/Exponenta/config"
	"github.com/Olegsandrik/Exponenta/internal/adapters/postgres"
	"github.com/Olegsandrik/Exponenta/internal/repository/dao"
	"github.com/Olegsandrik/Exponenta/internal/repository/repoErrors"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/logger"
	"github.com/Olegsandrik/Exponenta/utils"
)

const (
	promptChoiceGeneration = `You are a professional chef assistant. Provide detailed cooking recipes in Russian with 
	next json format. Send me only json od recipe.
	"name": "str",
	"description": "str",
	"servingsNum": int,
	"dishTypes": [
		"str",
		"str",
	],
	"diets": [
		"str",
		"str"
	],
	"ingredients": [
		{
			"id": 1, // 1... inf
			"name": "str",
			"image": "str",
			"amount": 0.5,
			"unit": "ч. л."
		},
		{
			"id": 2,
			"name": "chocolate",
			"image": "milk-chocolate.jpg",
			"amount": 8,
			"unit": "унций"
		}
	],
	"totalSteps": int, 
	"readyInMinutes": int,
	"steps": [
		{
			"number": int,
			"step": "description movement step",
			"ingredients": [
				{
					"name": "молотый эспрессо",
					"localizedName": "молотый эспрессо"
				},
				{
					"name": "взбитые сливки",
					"localizedName": "взбитые сливки"
					
				}
			],
			"equipment": [
				{
					"name": "пергамент для выпечки",
					"localizedName": "пергамент для выпечки",
				},
				{
					"name": "водяная баня",
					"localizedName": "водяная баня",
				}
			],
			"length": {
				"number": 5,
				"unit": "минут"
			}
		},
		{
			"number": int,
			"step": "description movement step",
			"ingredients": [
				{
					"name": "молотый эспрессо",
					"localizedName": "молотый эспрессо"
				},
				{
					"name": "взбитые сливки",
					"localizedName": "взбитые сливки"
				}
			],
			"equipment": [
				{
					"name": "пергамент для выпечки",
					"localizedName": "пергамент для выпечки",
				},
				{
					"name": "водяная баня",
					"localizedName": "водяная баня",
				}
			],
		}
	]
	You must use products, that i will send you`

	promptChoiceModernization = `You are a professional chef assistant. I will send you my recipe and you should reform
	my recipe with my promise. Send me only json of my new recipe.`
)

type Key struct {
	Value  string
	IsUsed bool
}

type KeysPool struct {
	Keys []Key
	Mu   sync.Mutex
}

type GeneratedRecipeRepo struct {
	storage  *postgres.Adapter
	config   *config.Config
	keysPool *KeysPool
}

func NewGeneratedRecipeRepo(storage *postgres.Adapter, config *config.Config) *GeneratedRecipeRepo {
	keysPool := NewKeysPool([]string{
		config.DeepSeekAPIKey2,
		config.DeepSeekAPIKey3,
		config.DeepSeekAPIKey4,
		config.DeepSeekAPIKey5,
	})

	return &GeneratedRecipeRepo{
		storage:  storage,
		config:   config,
		keysPool: keysPool,
	}
}

func NewKeysPool(Keys []string) *KeysPool {
	keys := make([]Key, 0, len(Keys))

	for _, key := range Keys {
		keys = append(keys, Key{Value: key, IsUsed: false})
	}

	return &KeysPool{
		Keys: keys,
		Mu:   sync.Mutex{},
	}
}

func (repo *GeneratedRecipeRepo) RefreshKeyByID(ID int) {
	repo.keysPool.Mu.Lock()
	defer repo.keysPool.Mu.Unlock()
	repo.keysPool.Keys[ID].IsUsed = false
}

func (repo *GeneratedRecipeRepo) GetKey() (string, int, error) {
	repo.keysPool.Mu.Lock()
	defer repo.keysPool.Mu.Unlock()

	for idx := range repo.keysPool.Keys {
		if !repo.keysPool.Keys[idx].IsUsed {
			repo.keysPool.Keys[idx].IsUsed = true
			return repo.keysPool.Keys[idx].Value, idx, nil
		}
	}

	return "", 0, repoErrors.ErrAllKeysAreUsing

}

func (repo *GeneratedRecipeRepo) GetAllRecipes(ctx context.Context, num int,
	userID uint) ([]models.RecipeModel, error) {
	q := `SELECT id, name, description, ready_in_minutes FROM public.generated_recipes WHERE user_id = $1 LIMIT $2`

	recipeRows := make([]dao.RecipeTable, 0, num)

	err := repo.storage.Select(ctx, &recipeRows, q, userID, num)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error getting recipe rows: %+v with num: %d", err, num))
		return nil, repoErrors.ErrFailToGetRecipes
	}

	recipeItems := dao.ConvertDaoToRecipe(recipeRows)

	if len(recipeItems) == 0 {
		logger.Error(ctx, fmt.Sprintf("error getting recipe zero row with num: %d", num))
		return nil, repoErrors.ErrZeroRowsGet
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
		return []dao.RecipeTable{}, repoErrors.ErrFailToGetRecipeByIDAndVersion
	}

	return recipeRows, nil
}

func (repo *GeneratedRecipeRepo) GetRecipeByID(ctx context.Context, recipeID int,
	userID uint) ([]models.RecipeModel, error) {
	q := `SELECT r.name, r.version, r.user_ingredients, r.query, r.description, r.ingredients, r.steps, r.dish_types, r.diets, r.servings, 
       r.ready_in_minutes FROM public.generated_recipes as r WHERE r.user_id = $1 AND r.id = $2`

	recipeRows := make([]dao.RecipeTable, 0, 1)

	err := repo.storage.Select(ctx, &recipeRows, q, userID, recipeID)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error getting recipe row: %s with rid: %d, uid: %d",
			err.Error(), recipeID, userID))
		return []models.RecipeModel{}, repoErrors.ErrFailToGetRecipeByID
	}

	recipeItem := dao.ConvertGenRecipeToRecipeModel(recipeRows)

	if len(recipeItem) == 0 {
		logger.Error(ctx, fmt.Sprintf("getting zero recipe row with rid: %d, uid: %d", recipeID, userID))
		return []models.RecipeModel{}, repoErrors.ErrFailToGetRecipeByID
	}

	logger.Info(ctx, fmt.Sprintf("select recipe with id: %d, uid: %d", recipeID, userID))

	return recipeItem, nil
}

func (repo *GeneratedRecipeRepo) GetHistoryByID(ctx context.Context, recipeID int,
	userID uint) ([]models.RecipeModel, error) {
	q := `SELECT r.id, r.ready_in_minutes, r.version, r.name, r.description, r.steps, r.dish_types, r.diets, r.servings, r.total_steps, 
       r.query FROM public.generated_recipes_versions as r WHERE user_id = $1 AND id = $2`

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
	APIKey, APIKeyID, err := repo.GetKey()

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("no free keys now: %v", err))
		return nil, err
	}

	defer repo.RefreshKeyByID(APIKeyID)

	q := fmt.Sprintf("my promise: %s, products: %s", query, strings.Join(products, ", "))

	respData, err := utils.GetResponseData(ctx, q, APIURL, APIKey, promptChoiceGeneration)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed to get resp data: %+v for userId: %d, query: %s, promptChoice: %s",
				err, userID, query, promptChoiceGeneration),
		)
		return nil, repoErrors.ErrWithGenerating
	}

	generatedRecipe, err := dao.ParseGeneratedRecipe(json.RawMessage(respData))

	generatedRecipe.Query = query
	jsonProducts, _ := json.Marshal(products)
	generatedRecipe.UserIngredients = jsonProducts

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf(`RESP from DeepSeek API: %s, 
				failed to parse recipe: %+v for userId: %d, query: %s, products: %s`,
				respData, err, userID, query, products),
		)
		return nil, repoErrors.ErrWithGenerating
	}

	tx, err := repo.storage.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf("failed to begin transaction on generating recipe: %d, err: %e", userID, err),
		)
		return nil, repoErrors.ErrWithGenerating
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
		return nil, repoErrors.ErrWithGenerating
	}

	recipeVersion, err := repo.insertVersionGeneratedRecipe(ctx,
		repo.storage, generatedRecipe, userID, generateRecipeID)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf(
				"fail to insert version generated recipe: %+v for userId: %d, genRecipe: %+v",
				err, userID, generatedRecipe),
		)
		return nil, repoErrors.ErrWithGenerating
	}

	if err = tx.Commit(); err != nil {
		return nil, repoErrors.ErrWithGenerating
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
            version,
            query
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, (
	SELECT COALESCE(MAX(version), 0) + 1 AS next_version
 		FROM public.generated_recipes_versions
 		WHERE id = $12
 	), $13) RETURNING version;`

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
		generatedRecipe.Query,
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
    user_id,name,description,dish_types,servings, diets, ingredients, ready_in_minutes, steps, total_steps, query, 
                                      user_ingredients)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id`

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
		generatedRecipe.Query,
		generatedRecipe.UserIngredients).Scan(&generateRecipeID)

	if err != nil {
		return 0, err
	}

	return generateRecipeID, nil
}

func (repo *GeneratedRecipeRepo) UpdateRecipe(ctx context.Context, query string, recipeID int, versionID int,
	userID uint) ([]models.RecipeModel, error) {
	APIURL := repo.config.DeepSeekAPIURL
	APIKey, APIKeyID, err := repo.GetKey()

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("no free keys now"))
		return nil, err
	}

	defer repo.RefreshKeyByID(APIKeyID)

	recipeDao, err := repo.getRecipeByIDAndVersion(ctx, recipeID, userID, versionID)

	if err != nil {
		return nil, err
	}

	if len(recipeDao) == 0 {
		return nil, repoErrors.ErrWithGenerating
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
		return nil, repoErrors.ErrWithModernization
	}

	generatedRecipe, err := dao.ParseGeneratedRecipe(json.RawMessage(respData))

	generatedRecipe.Query = query

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf(`failed to parse recipe: %+v for respData: %+v,`,
				err, respData),
		)
		return nil, repoErrors.ErrWithModernization
	}

	generateVersion, err := repo.insertVersionGeneratedRecipe(ctx, repo.storage, generatedRecipe, userID, recipeID)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf(`failed to insert version: %+v for genRecipe: %+v, uID: %d, recipeID: %d`,
				err, generatedRecipe, userID, recipeID),
		)
		return nil, repoErrors.ErrWithModernization
	}

	generatedRecipe.Version = generateVersion

	recipeModel := dao.ConvertGeneratedRecipeToRecipeModels([]dao.GeneratedRecipe{generatedRecipe})

	return recipeModel, nil
}

func (repo *GeneratedRecipeRepo) GetVersionByID(ctx context.Context, userID uint,
	recipeID int, versionID int) ([]dao.RecipeTable, error) {
	q := `SELECT r.name, r.query, r.description, r.steps, r.ingredients, 
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

	if len(recipeRows) == 0 {
		return repoErrors.ErrVersionNotFound
	}

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf(`failed to get recipe version: %+v for uID: %d, recipeID: %d, versionID: %d`,
				err, userID, recipeID, versionID),
		)
		return repoErrors.ErrWithUpdateVersion
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
			 total_steps = $9,
             version = $10
		 WHERE user_id = $11 AND id = $12`

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
		versionID,
		userID,
		recipeID,
	)

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf(
				`failed to update recipe version: %+v for uID: %d, recipeID: %d, versionID: %d, rows: %+v`,
				err, userID, recipeID, versionID, recipeRows),
		)
		return repoErrors.ErrWithUpdateVersion
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		logger.Error(ctx,
			fmt.Sprintf(
				`failed to get affected rows: %+v for uID: %d, recipeID: %d, versionID: %d, rows: %+v`,
				err, userID, recipeID, versionID, recipeRows),
		)
		return repoErrors.ErrWithUpdateVersion
	}

	if rowsAffected == 0 {
		logger.Error(ctx,
			fmt.Sprintf(
				`zero affected rows: %+v for uID: %d, recipeID: %d, versionID: %d, rows: %+v`,
				err, userID, recipeID, versionID, recipeRows),
		)
		return repoErrors.ErrWithGenerating
	}

	return nil
}
