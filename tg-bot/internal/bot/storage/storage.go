package storage

import "context"

type Storage interface {
	SetUserState(ctx context.Context, userID int64, state string) error
	GetUserState(ctx context.Context, userID int64) (string, error)
	SetUserData(ctx context.Context, userID int64, key, value string) error
	GetUserData(ctx context.Context, userID int64, key string) (string, error)
	ClearUserData(ctx context.Context, userID int64) error
	SaveMessage(ctx context.Context, userID int64, text string) error
	GetHistory(ctx context.Context, userID int64, limit int) (string, error)
	Close() error
}
