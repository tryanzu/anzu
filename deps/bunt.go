package deps

import (
	"github.com/tidwall/buntdb"
)

func IgniteBuntDB(container Deps) (Deps, error) {
	db, err := buntdb.Open("cache.db")
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()

	db.CreateIndex("usernames", "user:*:names", buntdb.IndexString)
	db.CreateIndex("assets", "asset:*:url", buntdb.IndexString)
	container.BuntProvider = db
	return container, nil
}
