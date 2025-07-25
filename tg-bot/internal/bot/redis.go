package bot

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

// Ключи для редиса
const (
	userHistoryKey = "user:%d:history"
	userStateKey   = "user:%d:state"
	userDataKey    = "user:%d:data"
	// Время очистки юзер-стейта
	keyTTL = 3 * time.Hour
)

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
	key := fmt.Sprintf(userHistoryKey, userID)
	pipe := r.client.Pipeline()
	pipe.LPush(ctx, key, message)
	pipe.LTrim(ctx, key, 0, 9) // Тут кст при нужде меняем сколько сообщений сохраняется
	pipe.Expire(ctx, key, keyTTL)
	_, err := pipe.Exec(ctx)
	return err
}

// Возвращает всю историю сообщений юзера по его телеграм ID в хронологическом порядке
func (r *RedisStore) GetHistory(ctx context.Context, userID int64, count int64) ([]string, error) {
	key := fmt.Sprintf(userHistoryKey, userID)
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

// Устанавливает юзер стейт для диалога
func (r *RedisStore) SetUserState(ctx context.Context, userID int64, state string) error {
	key := fmt.Sprintf(userStateKey, userID)
	return r.client.Set(ctx, key, state, keyTTL).Err()
}

// Возвращает юзер стейт, если юзер стейта нет, возвращает пустой стринг
func (r *RedisStore) GetUserState(ctx context.Context, userID int64) (string, error) {
	key := fmt.Sprintf(userStateKey, userID)
	state, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // нет стейта != ошибка
	}
	return state, err
}

// Сохраняет key-val пару для юзера (прим., "otp", "123456").
// Тут исползуется ключ 'data', по большому счёту я пока планирую использовать его только как хэш OTP.
func (r *RedisStore) SetUserData(ctx context.Context, userID int64, field, value string) error {
	key := fmt.Sprintf(userDataKey, userID)
	pipe := r.client.Pipeline()
	pipe.HSet(ctx, key, field, value)
	pipe.Expire(ctx, key, keyTTL)
	_, err := pipe.Exec(ctx)
	return err
}

// Возвращает хэш для юзера из поля 'data'.
func (r *RedisStore) GetUserData(ctx context.Context, userID int64, field string) (string, error) {
	key := fmt.Sprintf(userDataKey, userID)
	return r.client.HGet(ctx, key, field).Result()
}

// Весь хэш юзера очищается.
func (r *RedisStore) ClearUserData(ctx context.Context, userID int64) error {
	key := fmt.Sprintf(userDataKey, userID)
	return r.client.Del(ctx, key).Err()
}

// Закрывает клиент
func (r *RedisStore) Close() {
	r.client.Close()
}
