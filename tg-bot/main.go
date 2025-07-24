package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_TOKEN is not set")
	}

	// Редис
	store := NewRedisStore()
	defer store.Close()

	// Бот
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		userID := update.Message.From.ID
		text := update.Message.Text

		switch text {
		case "/start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Отправь сообщение, и раздастся эхо. Напиши /gethist чтобы получить историю.")
			bot.Send(msg)

		case "/gethist":
			messages, err := store.GetHistory(context.Background(), userID, 10)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении истории"))
				continue
			}

			if len(messages) == 0 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "История пуста"))
				continue
			}

			var response strings.Builder
			for i, m := range messages {
				response.WriteString(fmt.Sprintf("%d. %s\n", i+1, m))
			}

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, response.String()))

		default:
			// Эхо
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
			bot.Send(msg)

			// Лог в редис
			if err := store.SaveMessage(context.Background(), userID, text); err != nil {
				log.Printf("Redis save error: %v", err)
			}
		}
	}
}
