package main

import "context"

// Этот интерфейс я сделал для того чтобы можно было поменять редис на другие базы данных.
// Обеспечивает модулярность и легкую заменяемость компонентов.
type Store interface {
	SaveMessage(ctx context.Context, userID int64, message string) error
	GetHistory(ctx context.Context, userID int64, count int64) ([]string, error)
	Close()
}
