package redis

import (
	"context"
	"log"
	config "snack-shop/config"
	custom_log "snack-shop/pkg/logs"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	once   sync.Once
	client *redis.Client
)

func NewRedisClient() *redis.Client {

	redis_config := config.InitRedis()

	once.Do(func() {
		client = redis.NewClient(&redis.Options{
			Addr:     redis_config.RedisHost + ":" + redis_config.RedisPort,
			Password: redis_config.RedisPassword,
			DB:       redis_config.RedisDB,
		})
		pong, err := client.Ping(context.Background()).Result()
		if err != nil {
			custom_log.NewCustomLog("connect_redis_failed", err.Error(), "error")
			log.Fatalf("Could not connect to Redis: %v", err)
		}
		log.Printf("Connected to Redis successfully: %s", pong)
	})
	return client
}
