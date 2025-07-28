package handlers

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/common"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/states"
)

var (
	AgentChatKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏁 End Chat", "end_agent_chat"),
		),
	)
)

func HandleAgentIssue(ctx context.Context, env *common.HandlerEnv, chatID, userID int64, text string) {
	// TODO: a ticket for Jira or something similar should be created here
	log.Printf("STUB: Creating support ticket for user %d. Issue: %s", userID, text)

	env.Store.SetUserState(ctx, userID, states.StateAgentChat)
	msg := tgbotapi.NewMessage(chatID, "Thank you! Your request has been forwarded to an agent. You have entered chat mode with support. All subsequent messages will be sent to the agent.\n\nTo end the chat, use the button below.")
	msg.ReplyMarkup = AgentChatKeyboard
	env.API.Send(msg)
}

func HandleAgentChat(ctx context.Context, env *common.HandlerEnv, chatID, userID int64, text string, isCallback bool) {
	if isCallback && text == "end_agent_chat" {
		// Here we should probably notify support that the chat has been terminated
		env.API.Send(tgbotapi.NewMessage(chatID, "Chat with the agent has been terminated. Returning to the menu."))

		// Check if the user was logged in to return them to the correct menu
		loginStatus, _ := env.Store.GetUserData(ctx, userID, "logged_in")
		if loginStatus == "true" {
			env.Store.SetUserState(ctx, userID, states.StateLoggedIn)
			msg := tgbotapi.NewMessage(chatID, "How can I help you?")
			msg.ReplyMarkup = LoggedInKeyboard
			env.API.Send(msg)
		} else {
			// This case is theoretically impossible with the current logic, but its here for reliability
			env.Store.SetUserState(ctx, userID, states.StateStart)
			msg := tgbotapi.NewMessage(chatID, "How can I help you?")
			msg.ReplyMarkup = WelcomeKeyboard
			env.API.Send(msg)
		}
		return
	}

	// If it's not a callback, then it's a message for the agent
	if !isCallback {
		// TODO: forward the message to the support chat...
		log.Printf("STUB: Forwarding message to agent from chat %d: %s", chatID, text)
	}
}
