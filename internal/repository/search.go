package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Olegsandrik/Exponenta/utils"
	"strings"

	"github.com/Olegsandrik/Exponenta/internal/adapters/elasticsearch"
	"github.com/Olegsandrik/Exponenta/internal/repository/dao"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/logger"
)

type SearchRepository struct {
	Adapter *elasticsearch.Adapter
}

func NewSearchRepository(adapter *elasticsearch.Adapter) *SearchRepository {
	return &SearchRepository{
		Adapter: adapter,
	}
}

func (repo *SearchRepository) Search(ctx context.Context, query string) (models.SearchResponseModel, error) {
	q := `{
    "query": {
        "bool": {
            "should": [
                {
                    "multi_match": {
                        "query": "%s",
                        "fields": ["name^5", "description^3"],
                        "type": "best_fields",
                        "operator": "or",
                        "fuzziness": 2
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
            ]
        }
    }
	}`

	res, err := repo.Adapter.ElasticClient.Search(
		repo.Adapter.ElasticClient.Search.WithContext(ctx),
		repo.Adapter.ElasticClient.Search.WithIndex(elasticsearch.RecipeIndex),
		repo.Adapter.ElasticClient.Search.WithBody(strings.NewReader(fmt.Sprintf(q, query, query, query))),
	)

	defer res.Body.Close()

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("search err: %e with query: %s", err, query))
		return models.SearchResponseModel{}, utils.FailToSearchErr
	}

	var response dao.ResponseElasticRecipeIndex

	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("response decode error: %e with query: %s", err, query))
		return models.SearchResponseModel{}, utils.FailToSearchErr
	}

	if len(response.Hits.Hits) == 0 {
		logger.Error(ctx, fmt.Sprintf("no results found with query: %s", query))
		return models.SearchResponseModel{}, utils.NoFoundErr
	}

	result := dao.ConvertResponseElasticRecipeIndexToModel(response)

	logger.Info(ctx, fmt.Sprintf("success query: %s", query))

	return models.SearchResponseModel{
		Recipes: result,
	}, nil
}
