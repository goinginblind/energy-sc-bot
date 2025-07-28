package handlers

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/common"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/states"
)

var (
	WelcomeKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔎 General Inquiry"),
			tgbotapi.NewKeyboardButton("🔑 Login"),
		),
	)
)

func HandleStartState(ctx context.Context, env *common.HandlerEnv, chatID, userID int64, text string) {
	switch text {
	case "🔑 Login", "/login":
		env.Store.SetUserState(ctx, userID, states.StateAwaitingLoginInput)
		msg := tgbotapi.NewMessage(chatID, "Please enter your phone or email to log in.")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		env.API.Send(msg)
	case "🔎 General Inquiry":
		env.Store.SetUserState(ctx, userID, states.StateGeneralInquiry)
		msg := tgbotapi.NewMessage(chatID, "You are in general inquiry mode. Just write your question. To exit, use the button below or the /start command.")
		env.API.Send(msg)
	case "/start":
		msg := tgbotapi.NewMessage(chatID, "Hello! I am your virtual assistant. How can I help you?")
		msg.ReplyMarkup = WelcomeKeyboard
		env.API.Send(msg)
	default:
		// If the user immediately writes a question, we switch to RAG mode
		env.Store.SetUserState(ctx, userID, states.StateGeneralInquiry)
		HandleGeneralInquiryState(ctx, env, chatID, userID, text, false)
	}
}
