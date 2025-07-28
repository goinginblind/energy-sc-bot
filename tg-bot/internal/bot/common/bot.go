package common

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// BotAPI defines the interface for the Telegram Bot API,
// allowing for mock implementations for testing.
type BotAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	GetMe() (tgbotapi.User, error)
}