package deps

import (
	"github.com/xuyu/goredis"
)

var (
	RedisURL string = "tcp://127.0.0.1:6379"
)

func IgniteCache(container Deps) (Deps, error) {
	redis, err := goredis.DialURL(RedisURL)
	if err != nil {
		return container, err
	}

	container.CacheProvider = redis
	return container, nil
}
