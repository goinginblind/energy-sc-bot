package storage

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore() *RedisStore {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &RedisStore{client: rdb}
}

func (s *RedisStore) SetUserState(ctx context.Context, userID int64, state string) error {
	key := fmt.Sprintf("user:%d:state", userID)
	return s.client.Set(ctx, key, state, 0).Err()
}

func (s *RedisStore) GetUserState(ctx context.Context, userID int64) (string, error) {
	key := fmt.Sprintf("user:%d:state", userID)
	return s.client.Get(ctx, key).Result()
}

func (s *RedisStore) SetUserData(ctx context.Context, userID int64, key, value string) error {
	userKey := fmt.Sprintf("user:%d:data", userID)
	return s.client.HSet(ctx, userKey, key, value).Err()
}

func (s *RedisStore) GetUserData(ctx context.Context, userID int64, key string) (string, error) {
	userKey := fmt.Sprintf("user:%d:data", userID)
	return s.client.HGet(ctx, userKey, key).Result()
}

func (s *RedisStore) ClearUserData(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("user:%d:data", userID)
	return s.client.Del(ctx, key).Err()
}

func (s *RedisStore) SaveMessage(ctx context.Context, userID int64, text string) error {
	key := fmt.Sprintf("user:%d:history", userID)
	member := fmt.Sprintf("%d:%s", time.Now().Unix(), text)
	return s.client.LPush(ctx, key, member).Err()
}

func (s *RedisStore) GetHistory(ctx context.Context, userID int64, limit int) (string, error) {
	key := fmt.Sprintf("user:%d:history", userID)
	history, err := s.client.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return "", err
	}
	// History is joined into a string
	var result string
	for _, msg := range history {
		result += msg + "\n"
	}
	return result, nil
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}
