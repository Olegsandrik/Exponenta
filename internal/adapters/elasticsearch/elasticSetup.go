package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/Olegsandrik/Exponenta/internal/adapters/postgres"
	"github.com/Olegsandrik/Exponenta/internal/repository/dao"
	"github.com/elastic/go-elasticsearch/v8/esutil"

	"github.com/kozhurkin/pipers"
)

const (
	RecipeIndex = "recipes"
)

const (
	mappingRecipe = `{
		"mappings": {
			"properties": {
				"name": {
					"type": "text",
					"analyzer": "english"
				},
				"description": {
                	"type": "text",
					"analyzer": "english"
                },
                "image": {
                	"type": "text",
                    "index": false
                },
                "id": {
                     "type": integer,
                     "index": false
                }
			}
		}
	}`
)

func InitElasticSearchData(ctx context.Context, elasticSearchAdapter *Adapter,
	postgresAdapter *postgres.Adapter) error {
	pp := pipers.FromFuncs(func() (struct{}, error) {
		return struct{}{}, initRecipeIndex(ctx, elasticSearchAdapter, postgresAdapter)
	})

	_, err := pp.Resolve()

	return err
}

func initRecipeIndex(ctx context.Context, elasticSearchAdapter *Adapter,
	postgresAdapter *postgres.Adapter) error {
	if err := createRecipeIndex(elasticSearchAdapter); err != nil {
		return err
	}

	return insertRecipeDataInIndex(ctx, elasticSearchAdapter, postgresAdapter)
}

func createRecipeIndex(elasticSearchAdapter *Adapter) error {
	err := deleteIfExistIndex(elasticSearchAdapter, RecipeIndex)

	if err != nil {
		return err
	}

	res, err := elasticSearchAdapter.ElasticClient.Indices.Create(
		RecipeIndex,
		elasticSearchAdapter.ElasticClient.Indices.Create.WithBody(strings.NewReader(mappingRecipe)),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	return nil
}

func deleteIfExistIndex(elasticSearchAdapter *Adapter, indexName string) error {
	res, err := elasticSearchAdapter.ElasticClient.Indices.Exists([]string{indexName})
	defer res.Body.Close()

	if err != nil {
		return err
	}

	if res.Status() == "200 OK" {
		deleteRes, err := elasticSearchAdapter.ElasticClient.Indices.Delete([]string{indexName})

		defer deleteRes.Body.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func insertRecipeDataInIndex(ctx context.Context, elasticsearchAdapter *Adapter,
	postgresAdapter *postgres.Adapter) error {
	var recipes []dao.RecipeTable
	q := "SELECT id, description, name, image FROM public.recipes"

	err := postgresAdapter.Select(ctx, &recipes, q)
	if err != nil {
		return err
	}

	indexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: elasticsearchAdapter.ElasticClient,
		Index:  RecipeIndex,
	})

	if err != nil {
		return err
	}

	for _, recipe := range recipes {
		JSONrecipe, err := json.Marshal(recipe)

		if err != nil {
			return err
		}

		err = indexer.Add(
			ctx,
			esutil.BulkIndexerItem{
				Action: "index",
				Body:   bytes.NewReader(JSONrecipe),
			},
		)

		if err != nil {
			return err
		}
	}

	return indexer.Close(ctx)
}
