package main

import (
	"context"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/storage"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/client"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/logger"
)

func main() {
	logger.InitLogger()
	defer logger.Sync()
	log := logger.L()

	// Загрузка средовых переменных
	if err := godotenv.Load(); err != nil {
		log.Info("No .env file found")
	}
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	grpcAddr := os.Getenv("GRPC_SERVICE_ADDR")
	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "8081"
	}

	// Инициализация Редис ДБ и gRPC клиента
	ragClient := client.New(grpcAddr)
	log.Info("gRPC client initialized.")

	dataClient := client.NewDataServiceClient()
	log.Info("Data service client initialized.")

	redisStore := storage.NewRedisStore()
	defer redisStore.Close()

	log.Info("Redis store initialized.")

	api, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Fatal("Failed to initialize Telegram Bot API", zap.Error(err))
	}
	log.Info("Telegram Bot API initialized.")

	// Инициализация хэндлеров, бота и клиентов (РАГ клиента и ДАТА клиента)
	telegramBot := bot.New(api, redisStore, ragClient, dataClient, log)

	// Start Prometheus metrics server
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Info("Starting metrics server", zap.String("port", metricsPort))
		if err := http.ListenAndServe(":"+metricsPort, nil); err != nil {
			log.Error("Metrics server failed", zap.Error(err))
		}
	}()

	// Стартуем
	telegramBot.Start(context.Background())
}
