package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Olegsandrik/Exponenta/internal/adapters/elasticsearch"
	"github.com/Olegsandrik/Exponenta/internal/adapters/postgres"
	"github.com/Olegsandrik/Exponenta/internal/repository/dao"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/logger"
	"github.com/Olegsandrik/Exponenta/utils"
)

type SearchRepository struct {
	AdapterElastic  *elasticsearch.Adapter
	AdapterPostgres *postgres.Adapter
}

func NewSearchRepository(adapter *elasticsearch.Adapter, adapterPostgres *postgres.Adapter) *SearchRepository {
	return &SearchRepository{
		AdapterElastic:  adapter,
		AdapterPostgres: adapterPostgres,
	}
}

func (repo *SearchRepository) Search(ctx context.Context, query string, diet string, dishType string,
	maxTime int) (models.SearchResponseModel, error) {
	q := ` 
    {
		"query": {
			"bool": {
				"must": [
					{
						"bool": {
							"should": [
								{
									"multi_match": {
										"query": "%s",
										"fields": ["name^5", "description^3"],
										"type": "best_fields",
										"operator": "or"
									}
								},
								{
									"match_phrase": {
										"name": {
											"query": "%s",
											"boost": 5
										}
									}
								},
								{
									"match_phrase": {
										"description": {
											"query": "%s",
											"boost": 5
										}
									}
								}
							],
							"minimum_should_match": 1
						}
					}
				],
				"filter": [
					%s,
					%s,
					%s
				]
			}
		}
	}`

	maxTimeFilter, dishTypeFilter, dietFilter := utils.FilterForElasticsearchRecipeIndex(maxTime, dishType, diet)

	res, err := repo.AdapterElastic.ElasticClient.Search(
		repo.AdapterElastic.ElasticClient.Search.WithContext(ctx),
		repo.AdapterElastic.ElasticClient.Search.WithIndex(elasticsearch.RecipeIndex),
		repo.AdapterElastic.ElasticClient.Search.WithBody(strings.NewReader(fmt.Sprintf(
			q,
			query,
			query,
			query,
			maxTimeFilter,
			dishTypeFilter,
			dietFilter))),
	)

	defer res.Body.Close()

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("search err: %e with query: %s", err, query))
		return models.SearchResponseModel{}, utils.ErrFailToSearch
	}

	var response dao.ResponseElasticRecipeIndex

	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("response decode error: %e with query: %s", err, query))
		return models.SearchResponseModel{}, utils.ErrFailToSearch
	}

	if len(response.Hits.Hits) == 0 {
		logger.Error(ctx, fmt.Sprintf("no results found with query: %s", query))
		return models.SearchResponseModel{}, utils.ErrNoFound
	}

	result := dao.ConvertResponseElasticRecipeIndexToModel(response)

	logger.Info(ctx, fmt.Sprintf("success query: %s", query))

	return models.SearchResponseModel{
		Recipes: result,
	}, nil
}

func (repo *SearchRepository) Suggest(ctx context.Context, query string) (models.SuggestResponseModel, error) {
	q := `{
	  "query": {
		"match": {
		  "name": {
			"query": "%s",
			"operator": "and"
		  }
		}
	  },
	  "size":5
	}`

	res, err := repo.AdapterElastic.ElasticClient.Search(
		repo.AdapterElastic.ElasticClient.Search.WithContext(ctx),
		repo.AdapterElastic.ElasticClient.Search.WithIndex(elasticsearch.SuggestIndex),
		repo.AdapterElastic.ElasticClient.Search.WithBody(strings.NewReader(fmt.Sprintf(q, query))),
	)

	defer res.Body.Close()

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("suggest err: %e with query: %s", err, query))
		return models.SuggestResponseModel{}, utils.ErrFailToGetSuggest
	}

	var response dao.ResponseElasticSuggestIndex

	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("suggest decode error: %e with query: %s", err, query))
		return models.SuggestResponseModel{}, utils.ErrFailToGetSuggest
	}

	if len(response.Hits.Hits) == 0 {
		logger.Info(ctx, fmt.Sprintf("success empty response query: %s", query))
		return models.SuggestResponseModel{}, nil
	}

	result := dao.ConvertResponseElasticSuggestIndexToModel(response)

	logger.Info(ctx, fmt.Sprintf("success query: %s", query))
	return result, nil
}

func (repo *SearchRepository) GetDiets(ctx context.Context) ([]string, error) {
	return repo.getFilter(ctx, "diets")
}

func (repo *SearchRepository) GetDishTypes(ctx context.Context) ([]string, error) {
	return repo.getFilter(ctx, "dish_types")
}

func (repo *SearchRepository) getFilter(ctx context.Context, filter string) ([]string, error) {
	var items []json.RawMessage

	q := "SELECT %s FROM recipes"

	err := repo.AdapterPostgres.Select(ctx, &items, fmt.Sprintf(q, filter))

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("query err: %e with query: %s", err, q))
		return nil, fmt.Errorf("err: %e with filter: %s", utils.ErrToGetFilterValues, filter)
	}

	if len(items) == 0 {
		logger.Error(ctx, fmt.Sprintf("no results found with query: %s", q))
		return nil, fmt.Errorf("err: %e with filter: %s", utils.ErrToGetFilterValues, filter)
	}

	hashMap, err := dao.MakeSet(items)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("query err: %e with query: %s", err, q))
		return nil, fmt.Errorf("err: %e with filter: %s", utils.ErrToGetFilterValues, filter)
	}

	result := make([]string, 0, len(hashMap))

	for diet := range hashMap {
		result = append(result, diet)
	}

	logger.Info(ctx, fmt.Sprintf("success query: %s", q))

	return result, nil
}

func (repo *SearchRepository) GetMaxMinCookingTime(ctx context.Context) (models.TimeModel, error) {
	q := "SELECT Min(ready_in_minutes), Max(ready_in_minutes) FROM public.recipes"

	time := make([]dao.TimeDao, 0, 1)

	err := repo.AdapterPostgres.Select(ctx, &time, q)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("query err: %e with query: %s", err, q))
		return models.TimeModel{}, utils.ErrGetMaxMinCookingTime
	}

	if len(time) == 0 {
		logger.Error(ctx, fmt.Sprintf("no row with query: %s", q))
		return models.TimeModel{}, utils.ErrGetMaxMinCookingTime
	}

	timeModel := dao.ConvertTimeDaoToModel(time[0])

	logger.Info(ctx, fmt.Sprintf("success query: %s", q))

	return timeModel, nil
}
