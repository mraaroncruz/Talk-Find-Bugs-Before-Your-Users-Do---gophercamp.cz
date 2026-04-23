package cache

import (
	"log"

	"github.com/redis/go-redis/v9"
)

func Connect(redisURL string) *redis.Client {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("redis parse url: %v", err)
	}

	return redis.NewClient(opts)
}
