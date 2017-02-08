package search

import (
	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/olebedev/config"
)

type Module struct {
	client  algoliasearch.Client
	indices map[string]algoliasearch.Index
}

func (module Module) Get(name string) algoliasearch.Index {

	if index, exists := module.indices[name]; exists {

		return index
	}

	panic("Invalid index name, not initialized")
}

func Boot(module_config *config.Config) *Module {

	config_application, err := module_config.String("application")

	if err != nil {
		panic(err)
	}

	config_api, err := module_config.String("api_key")

	if err != nil {
		panic(err)
	}

	// Initialize algolia client
	client := algoliasearch.NewClient(config_application, config_api)

	// Prepare available indices
	indices := make(map[string]algoliasearch.Index)
	config_indices, err := module_config.List("indexes")

	if err != nil {
		panic(err)
	}

	for _, index := range config_indices {

		// Type asset the index data
		index_data := index.(map[string]interface{})
		index_name, index_name_exists := index_data["name"]
		index_target, index_target_exists := index_data["index"]

		if !index_name_exists || !index_target_exists {
			panic("Invalid search module configuration.")
		}

		indices[index_name.(string)] = client.InitIndex(index_target.(string))
	}

	module := &Module{client, indices}

	return module
}
