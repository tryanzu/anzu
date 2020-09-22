package deps

import (
	"github.com/xuyu/goredis"
)

var (
	RedisURL string = "127.0.0.1:6379"
)

func IgniteCache(container Deps) (Deps, error) {
	redis, err := goredis.Dial(&goredis.DialConfig{Address: RedisURL})
	if err != nil {
		return container, err
	}

	container.CacheProvider = redis
	return container, nil
}
