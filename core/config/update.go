package config

import (
	"bytes"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/divideandconquer/go-merge/merge"
)

func MergeUpdate(config map[string]interface{}) error {

	// Copy of current runtime config.
	current := C.UserCopy()
	merged := merge.Merge(current, config)
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)

	if err := encoder.Encode(merged); err != nil {
		return err
	}

	if err := os.WriteFile("./config.toml", buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}
