package deps

import (
	"github.com/tidwall/buntdb"
)

func IgniteBuntDB(container Deps) (Deps, error) {
	db, err := buntdb.Open(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()

	container.BuntProvider = db
	return container, nil
}
