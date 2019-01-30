package deps

import (
	lediscfg "github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
)

func IgniteLedisDB(container Deps) (Deps, error) {
	conf := lediscfg.NewConfigDefault()
	conn, err := ledis.Open(conf)
	if err != nil {
		log.Fatal(err)
	}

	db, err := conn.Select(0)
	if err != nil {
		log.Fatal(err)
	}

	container.LedisProvider = db
	return container, nil
}
