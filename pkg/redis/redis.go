package redis_util

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisUtil struct {
	Client *redis.Client
	Ctx    context.Context
}

type CacheData struct {
	Exp          int    `json:"exp"`
	LoginSession string `json:"login_session"`
	UserID       int    `json:"user_id"`
	Username     string `json:"username"`
}

type Cards struct {
	ID            int        `json:"id" db:"id"`
	SuitName      string     `db:"suit_name"`
	SuitSymbol    string     `db:"suit_symbol"`
	CardName      string     `json:"card_name" db:"card_name"`
	CardNumber    int        `json:"card_number" db:"card_number"`
	IsAllowRandom bool       `json:"is_allow_random" db:"is_allow_random"`
	CardValue     int        `json:"card_value" db:"card_value"`
	CardTypeID    int        `json:"card_type_id" db:"card_type_id"`
	CardImage     *string    `json:"card_image" db:"card_image"`
	StatusID      int        `json:"status_id" db:"status_id"`
	Order         int        `json:"order" db:"order"`
	CreatedBy     int        `json:"-" db:"created_by"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedBy     *int       `json:"-" db:"updated_by"`
	UpdatedAt     time.Time  `json:"-" db:"updated_at"`
	DeletedBy     *int       `json:"-" db:"deleted_by"`
	DeletedAt     *time.Time `json:"-" db:"deleted_at"`
}

func NewRedisUtil(rdb *redis.Client) *RedisUtil {
	return &RedisUtil{
		Client: rdb,
		Ctx:    context.Background(),
	}
}

func (r *RedisUtil) SetCacheKey(key string, value interface{}, ctx context.Context) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	duration := time.Duration(0)
	return r.Client.Set(ctx, key, data, duration).Err()
}

func (r *RedisUtil) SetEx(key string, data interface{}, expiredTime time.Duration, ctx context.Context) error {
	jsonData, errData := json.Marshal(data)
	if errData != nil {
		return errData
	}

	err := r.Client.Set(ctx, key, jsonData, expiredTime).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisUtil) GetCache(key string, result interface{}) error {
	value, err := r.Client.Get(r.Ctx, key).Result()
	if err == redis.Nil {
		return nil
	} else if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), result)
}

func Get[T any](r *RedisUtil, key string, ctx context.Context) (*T, error) {
	// Get data from Redis
	result, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// Unmarshal into generic type
	var data T
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	return &data, nil
}

func (r *RedisUtil) GetCacheKey(key string, ctx context.Context) (*CacheData, error) {
	result, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("key %s does not exist", key)
	} else if err != nil {
		return nil, err
	}
	var data CacheData
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return &data, nil
}

func (r *RedisUtil) DeleteCache(key string) error {
	deleted, err := r.Client.Del(r.Ctx, key).Result()
	if err != nil {
		return err
	}
	if deleted == 0 {
		return fmt.Errorf("key %s does not exist", key)
	}
	return nil
}

func (r *RedisUtil) AddToBlockList(token string, expiration time.Duration) error {
	return r.Client.Set(r.Ctx, "blokclist:"+token, "revoked", expiration).Err()
}

func (r *RedisUtil) IsTokenRevoked(token string) (bool, error) {
	val, err := r.Client.Get(r.Ctx, "blocklist:"+token).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, nil
	}
	return val == "revoked", nil
}

func (r *RedisUtil) RateLimit(key string, limit int, window time.Duration) (bool, error) {
	count, err := r.Client.Incr(r.Ctx, key).Result()
	if err != nil {
		return false, nil
	}
	if count == 1 {
		_ = r.Client.Expire(r.Ctx, key, window).Err()
	}
	return count <= int64(limit), nil
}

func (r *RedisUtil) CloseConnection() error {
	return r.Client.Close()
}

func StoreCardData(redisClient *redis.Client, key string, data []byte) error {
	duration := time.Duration(0)
    ctx := context.Background()
    err := redisClient.Set(ctx, key, data, duration).Err()
    if err != nil {
        return fmt.Errorf("failed to store cards in Redis: %w", err)
    }
    return nil
}

func GetCardData(redisClient *redis.Client, key string) ([]Cards, error) {
	ctx := context.Background()

	// Get from Redis
	cardBytes, err := redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get cards from Redis: %w", err)
	}

	// Unmarshal JSON
	var cards []Cards
	err = json.Unmarshal(cardBytes, &cards)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cards: %w", err)
	}

	return cards, nil
}

