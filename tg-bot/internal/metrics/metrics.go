package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// TelegramBotMessagesReceived counts the total number of messages received by the bot.
	TelegramBotMessagesReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "telegram_bot_messages_received_total",
			Help: "Total number of messages received by the Telegram bot.",
		},
	)

	// TelegramBotCallbackQueriesReceived counts the total number of callback queries received by the bot.
	TelegramBotCallbackQueriesReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "telegram_bot_callback_queries_received_total",
			Help: "Total number of callback queries received by the Telegram bot.",
		},
	)

	// TelegramBotHandlerDuration measures the duration of handler executions.
	TelegramBotHandlerDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "telegram_bot_handler_duration_seconds",
			Help:    "Duration of Telegram bot handler executions in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"handler"},
	)

	// RAGServiceCallsTotal counts the total number of calls to the RAG service.
	RAGServiceCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rag_service_calls_total",
			Help: "Total number of calls to the RAG service.",
		},
		[]string{"method", "status"},
	)

	// DataServiceCallsTotal counts the total number of calls to the Data service.
	DataServiceCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "data_service_calls_total",
			Help: "Total number of calls to the Data service.",
		},
		[]string{"method", "status"},
	)
)
