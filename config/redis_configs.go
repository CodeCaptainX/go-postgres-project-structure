package configs

import (
	"log"
	"os"
	env "snack-shop/pkg/utils/env"

	"github.com/joho/godotenv"
)

type RedisConfig struct {
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	RedisExpire   int
}

func InitRedis() *RedisConfig {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found, using system environment variables")
	}

	redis_host := os.Getenv("REDIS_HOST")
	redis_port := os.Getenv("REDIS_PORT")
	redis_password := os.Getenv("REDIS_PASSWORD")
	redis_db := env.GetenvInt("REDIS_DB", 0)
	redis_exprie := env.GetenvInt("REDIS_EXPIRE", 60)

	return &RedisConfig{
		RedisHost:     redis_host,
		RedisPort:     redis_port,
		RedisPassword: redis_password,
		RedisDB:       redis_db,
		RedisExpire:   redis_exprie,
	}
}
