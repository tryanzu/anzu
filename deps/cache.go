package deps

import (
	"github.com/go-redis/redis/v8"
)

var (
	RedisURL string = "redis://127.0.0.1:6379"
)

func IgniteCache(container Deps) (Deps, error) {
	url, err := redis.ParseURL(RedisURL)
	if err != nil {
		return container, err
	}
	client := redis.NewClient(url)
	container.CacheProvider = client
	return container, nil
}
