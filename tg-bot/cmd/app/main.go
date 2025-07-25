package main

import (
	"context"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/client"
)

func main() {
	// Энв
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	grpcAddr := os.Getenv("GRPC_SERVICE_ADDR")

	// запуск клиента и редис дб
	serviceClient := client.New(grpcAddr)
	log.Println("gRPC client initialized.")

	redisStore := bot.NewRedisStore() // Вообще по-хорошему бы отделить как-то хранилище от бот-пакета, но пока так
	defer redisStore.Close()

	log.Println("Redis store initialized.")

	api, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Panic(err)
	}
	log.Println("Telegram Bot API initialized.")

	// Хэндлеры + бот, внедрение зависимостей
	botHandlers := bot.NewHandlers(api, redisStore, serviceClient)
	telegramBot := bot.New(api, botHandlers, redisStore)

	// Запуск бота
	telegramBot.Start(context.Background())
}
