package bot

import "context"

// Этот интерфейс я сделал для того чтобы можно было поменять редис на другие базы данных.
// Обеспечивает модулярность и легкую заменяемость компонентов.
type Store interface {
	// Методы для истории (сохранение + ретривал)
	SaveMessage(ctx context.Context, userID int64, message string) error
	GetHistory(ctx context.Context, userID int64, count int64) ([]string, error)

	// Отслеживание юзер стейта
	SetUserState(ctx context.Context, userID int64, state string) error
	GetUserState(ctx context.Context, userID int64) (string, error)

	// Методы для временных частей юзер стейта, это например сохранение OTP
	SetUserData(ctx context.Context, userID int64, field, value string) error
	GetUserData(ctx context.Context, userID int64, field string) (string, error)
	ClearUserData(ctx context.Context, userID int64) error

	// Закрытие
	Close()
}
