package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/common"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/handlers"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/states"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/storage"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/client"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/metrics"
	"github.com/goinginblind/energy-sc-bot/tg-bot/ragpb"
)

// Bot struct to hold dependencies
type Bot struct {
	env *common.HandlerEnv
}

// New creates a new bot instance.
func New(api common.BotAPI, store storage.Storage, ragClient ragpb.RAGServiceClient, dataClient client.DataClientInterface, logger *zap.Logger) *Bot {
	return &Bot{
		env: &common.HandlerEnv{
			API:        api,
			Store:      store,
			RAGClient:  ragClient,
			DataClient: dataClient,
			Logger:     logger,
		},
	}
}

// Start begins the main bot update loop.
func (b *Bot) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.env.API.GetUpdatesChan(u)

	self, err := b.env.API.GetMe()
	if err != nil {
		b.env.Logger.Fatal("Failed to get bot info", zap.Error(err))
	}

	b.env.Logger.Info("Authorized on account", zap.String("username", self.UserName), zap.String("status", "starting update loop"))

	// Main loop for all bot logic
	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		ctx := context.Background()
		var chatID int64
		var userID int64
		var text string
		var isCallback bool

		if update.CallbackQuery != nil {
			// Handling inline keyboard button presses
			isCallback = true
			callback := update.CallbackQuery
			chatID = callback.Message.Chat.ID
			userID = callback.From.ID
			text = callback.Data

			// Response is needed to stop the button from glowing (loading state)
			b.env.API.Request(tgbotapi.NewCallback(callback.ID, ""))
			b.env.Logger.Info("Callback query received",
				zap.Int64("chat_id", chatID),
				zap.Int64("user_id", userID),
				zap.String("data", text),
			)
			metrics.TelegramBotCallbackQueriesReceived.Inc()
		} else {
			// Regular messages, without keyboards == plain text
			chatID = update.Message.Chat.ID
			userID = update.Message.From.ID
			text = update.Message.Text
			// Every text message is logged
			b.env.Store.SaveMessage(ctx, userID, text)
			b.env.Logger.Info("Message received",
				zap.Int64("chat_id", chatID),
				zap.Int64("user_id", userID),
				zap.String("text", text),
			)
			metrics.TelegramBotMessagesReceived.Inc()
		}

		// Get the current state from Redis
		userState, err := b.env.Store.GetUserState(ctx, userID)
		if err != nil {
			b.env.Logger.Error("Error getting user state",
				zap.Int64("user_id", userID),
				zap.Error(err),
			)
			userState = states.StateStart
		}

		// State handling machine
		switch userState {
		case states.StateStart:
			handlers.HandleStartState(ctx, b.env, chatID, userID, text)
		case states.StateAwaitingLoginInput:
			handlers.HandleAwaitingLoginInput(ctx, b.env, chatID, userID, text)
		case states.StateAwaitingOTP:
			handlers.HandleAwaitingOTP(ctx, b.env, chatID, userID, text)
		case states.StateLoggedIn:
			handlers.HandleLoggedInState(ctx, b.env, chatID, userID, text, isCallback)
		case states.StateGeneralInquiry:
			handlers.HandleGeneralInquiryState(ctx, b.env, chatID, userID, text, isCallback)
		case states.StateAwaitingAgentIssuePost:
			handlers.HandleAgentIssue(ctx, b.env, chatID, userID, text)
		case states.StateAgentChat:
			handlers.HandleAgentChat(ctx, b.env, chatID, userID, text, isCallback)
		default:
			// Unknown state = error, but this should not happen
			b.env.Logger.Warn("Unknown user state",
				zap.Int64("user_id", userID),
				zap.String("state", userState),
			)
			b.env.Store.SetUserState(ctx, userID, states.StateStart)
			msg := tgbotapi.NewMessage(chatID, "An error occurred. Let's start over.")
			msg.ReplyMarkup = handlers.WelcomeKeyboard
			b.env.API.Send(msg)
		}
	}
}
