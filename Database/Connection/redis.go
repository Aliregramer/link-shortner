package Connection

import (
	h "shorjiga/Helper"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
)

func RedisClient() *redis.Client {
	godotenv.Load()
	return redis.NewClient(&redis.Options{
		Addr:     h.Getenv("REDIS_HOST", "127.0.0.1") + ":" + h.Getenv("REDIS_PORT", "6379"),
		Password: "",
		DB:       0,
	})
}
