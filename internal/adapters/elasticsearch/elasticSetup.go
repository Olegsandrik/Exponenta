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
	RecipeIndex  = "recipes"
	SuggestIndex = "suggest"
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
	mappingSuggest = `{
	  "settings": {
		"analysis": {
		  "tokenizer": {
			"ngram_tokenizer": {
			  "type": "edge_ngram", 
			  "min_gram": 1, 
			  "max_gram": 20,
			  "token_chars": [
				"letter",
				"digit", 
				"custom" 
			  ],
			  "custom_token_chars": ".,-_"
			}
		  },
		  "analyzer": {
			"custom_analyzer": {
			  "type": "custom", 
			  "tokenizer": "ngram_tokenizer",
			  "filter": [
				"lowercase"
			  ]
			}
		  }
		}
	  },
	  "mappings" : {
		  "properties" : {
			"name" : {
			  "type" : "text",
			  "analyzer": "custom_analyzer",
			  "search_analyzer": "custom_analyzer",
			  "fields" : {
				"keyword" : {
				  "type" : "keyword",
				  "ignore_above" : 256
				}
			  }
			}
		  }
	  }
	}`
)

func InitElasticSearchData(ctx context.Context, elasticSearchAdapter *Adapter,
	postgresAdapter *postgres.Adapter) error {
	pp := pipers.FromFuncs(
		func() (struct{}, error) {
			return struct{}{}, initRecipeIndex(ctx, elasticSearchAdapter, postgresAdapter)
		},
		func() (struct{}, error) {
			return struct{}{}, initSuggestIndex(ctx, elasticSearchAdapter, postgresAdapter)
		})

	_, err := pp.Resolve()

	return err
}

func initRecipeIndex(ctx context.Context, elasticSearchAdapter *Adapter,
	postgresAdapter *postgres.Adapter) error {
	if err := createIndex(elasticSearchAdapter, RecipeIndex, mappingRecipe); err != nil {
		return err
	}

	return insertDataInIndex(ctx, elasticSearchAdapter, postgresAdapter,
		"SELECT id, description, name, image FROM public.recipes", RecipeIndex)
}

func initSuggestIndex(ctx context.Context, elasticSearchAdapter *Adapter,
	postgresAdapter *postgres.Adapter) error {
	if err := createIndex(elasticSearchAdapter, SuggestIndex, mappingSuggest); err != nil {
		return err
	}

	return insertDataInIndex(ctx, elasticSearchAdapter, postgresAdapter,
		"SELECT name FROM public.recipes", SuggestIndex)
}

func insertDataInIndex(ctx context.Context, elasticsearchAdapter *Adapter,
	postgresAdapter *postgres.Adapter, q string, index string) error {
	var recipes []dao.RecipeTable

	err := postgresAdapter.Select(ctx, &recipes, q)
	if err != nil {
		return err
	}

	indexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: elasticsearchAdapter.ElasticClient,
		Index:  index,
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

func createIndex(elasticSearchAdapter *Adapter, index string, mapping string) error {
	err := deleteIfExistIndex(elasticSearchAdapter, index)

	if err != nil {
		return err
	}

	res, err := elasticSearchAdapter.ElasticClient.Indices.Create(
		index,
		elasticSearchAdapter.ElasticClient.Indices.Create.WithBody(strings.NewReader(mapping)),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	return nil
}
