package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

// Новый редисник
func NewRedisStore() *RedisStore {
	addr := os.Getenv("REDIS_ADDR")
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")
	db, _ := strconv.Atoi(dbStr)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisStore{client: client}
}

// Логирует одно сообщение
func (r *RedisStore) SaveMessage(ctx context.Context, userID int64, message string) error {
	key := fmt.Sprintf("user:%d:history", userID)
	if err := r.client.LPush(ctx, key, message).Err(); err != nil {
		return err
	}
	return r.client.LTrim(ctx, key, 0, 9).Err()
}

// Возвращает всю историю сообщений юзера по его телеграм ID в хронологическом порядке
func (r *RedisStore) GetHistory(ctx context.Context, userID int64, count int64) ([]string, error) {
	key := fmt.Sprintf("user:%d:history", userID)
	messages, err := r.client.LRange(ctx, key, 0, count-1).Result()
	if err != nil {
		return nil, err
	}

	// Реверс чтоб было в хронологическом порядке
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// Закрывает клиент
func (r *RedisStore) Close() {
	r.client.Close()
}
