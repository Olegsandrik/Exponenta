package elasticsearch

import (
	"github.com/Olegsandrik/Exponenta/config"
	"github.com/elastic/go-elasticsearch/v8"
)

type Adapter struct {
	ElasticClient *elasticsearch.Client
}

func NewElasticsearchAdapter(cfg *config.Config) (*Adapter, error) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.ElasticsearchAddress},
		Username:  cfg.ElasticsearchUsername,
		Password:  cfg.ElasticsearchPassword,
	})

	if err != nil {
		return nil, err
	}

	_, err = client.Ping()

	if err != nil {
		return nil, err
	}

	return &Adapter{
		ElasticClient: client,
	}, nil
}
