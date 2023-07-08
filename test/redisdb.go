package test

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func NewRedisDB() (*miniredis.Miniredis, redis.UniversalClient) {
	mredis, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: mredis.Addr()})

	return mredis, redisClient
}
