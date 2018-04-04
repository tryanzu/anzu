package config

import (
	"io/ioutil"

	"github.com/hjson/hjson-go"
	"github.com/imdario/mergo"
)

func MergeUpdate(config map[string]interface{}) error {
	merged := make(map[string]interface{}, len(config))
	for k, v := range config {
		merged[k] = v
	}

	// Copy of current runtime config.
	current := C.Copy()
	if err := mergo.Merge(&merged, current, mergo.WithOverride); err != nil {
		return err
	}

	formatted, err := hjson.Marshal(merged)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("./config.hjson", formatted, 0644)
	if err != nil {
		return err
	}

	return nil
}
