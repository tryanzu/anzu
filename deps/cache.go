package deps

import (
	"github.com/xuyu/goredis"
)

func IgniteCache(container Deps) (Deps, error) {
	address, err := container.Config().String("cache.redis")
	if err != nil {
		return container, err
	}

	redis, err := goredis.Dial(&goredis.DialConfig{Address: address})
	if err != nil {
		return container, err
	}

	container.CacheProvider = redis
	return container, nil
}
