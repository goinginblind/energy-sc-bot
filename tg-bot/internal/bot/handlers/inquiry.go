package handlers

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/common"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/states"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/metrics"
	"github.com/goinginblind/energy-sc-bot/tg-bot/ragpb"
)

var (
	GeneralInquiryKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏁 End Dialogue", "end_general_inquiry"),
		),
	)
)

func HandleGeneralInquiryState(ctx context.Context, env *common.HandlerEnv, chatID, userID int64, text string, isCallback bool) {
	// Measure handler duration
	defer func(start time.Time) {
		metrics.TelegramBotHandlerDuration.WithLabelValues("HandleGeneralInquiryState").Observe(time.Since(start).Seconds())
	}(time.Now())

	if isCallback && text == "end_general_inquiry" {
		loginStatus, _ := env.Store.GetUserData(ctx, userID, "logged_in")
		if loginStatus == "true" {
			env.Store.SetUserState(ctx, userID, states.StateLoggedIn)
			msg := tgbotapi.NewMessage(chatID, "The dialogue is over. Returning to your account menu.")
			msg.ReplyMarkup = LoggedInKeyboard
			env.API.Send(msg)
		} else {
			env.Store.SetUserState(ctx, userID, states.StateStart)
			msg := tgbotapi.NewMessage(chatID, "The dialogue is over. Returning to the main menu.")
			msg.ReplyMarkup = WelcomeKeyboard
			env.API.Send(msg)
		}
		env.Logger.Info("General inquiry dialogue ended",
			zap.Int64("user_id", userID),
			zap.Bool("is_logged_in", loginStatus == "true"),
		)
		return
	}

	if !isCallback && (text == "🔎 General Inquiry" || text == "🔑 Login") {
		env.Store.SetUserState(ctx, userID, states.StateStart)
		HandleStartState(ctx, env, chatID, userID, text)
		return
	}

	// Fetch user account data if logged in
	var accData string
	loginStatus, _ := env.Store.GetUserData(ctx, userID, "logged_in")
	if loginStatus == "true" {
		dbUserIDstr, err := env.Store.GetUserData(ctx, userID, "user_id")
		if err == nil {
			dbUserID, _ := strconv.ParseInt(dbUserIDstr, 10, 64)
			bills, err := env.DataClient.GetBillsByUserID(dbUserID)
			if err != nil {
				env.Logger.Error("Could not fetch bills for RAG context",
					zap.Int64("user_id", userID),
					zap.Error(err),
				)
				metrics.DataServiceCallsTotal.WithLabelValues("GetBillsByUserID", "error").Inc()
			} else {
				metrics.DataServiceCallsTotal.WithLabelValues("GetBillsByUserID", "success").Inc()
				billsJSON, err := json.Marshal(bills)
				if err != nil {
					env.Logger.Error("Could not marshal bills for RAG context",
						zap.Int64("user_id", userID),
						zap.Error(err),
					)
				} else {
					accData = string(billsJSON)
				}
			}
		}
	}

	var msg tgbotapi.MessageConfig
	env.Logger.Info("Calling RAG service for query classification",
		zap.Int64("user_id", userID),
		zap.String("query", text),
	)
	ragLabelStruct, err := env.RAGClient.ClassifyQuery(ctx, &ragpb.ClassifyRequest{Query: text})
	if err != nil {
		metrics.RAGServiceCallsTotal.WithLabelValues("ClassifyQuery", "error").Inc()
		env.Logger.Error("Failed to call RAG service and classify request",
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		msg = tgbotapi.NewMessage(chatID, "Sorry, I could not process your request, please try again later")
		msg.ReplyMarkup = GeneralInquiryKeyboard
		env.API.Send(msg)
		return
	} else {
		metrics.RAGServiceCallsTotal.WithLabelValues("ClassifyQuery", "success").Inc()
	}

	env.Logger.Info("RAG query classified",
		zap.Int64("user_id", userID),
		zap.String("query", text),
		zap.String("label", ragLabelStruct.Label),
	)

	userHistory, _ := env.Store.GetHistory(ctx, userID, 10)
	env.Logger.Info("Calling RAG service for answer generation",
		zap.Int64("user_id", userID),
		zap.String("query", text),
		zap.String("label", ragLabelStruct.Label),
		zap.String("acc_data_len", strconv.Itoa(len(accData))),
	)
	ragAnsStruct, err := env.RAGClient.GetAnswerToQuery(ctx, &ragpb.AnswerRequest{Query: text, History: userHistory, Label: ragLabelStruct.Label, AccData: accData})
	if err != nil {
		metrics.RAGServiceCallsTotal.WithLabelValues("GetAnswerToQuery", "error").Inc()
		env.Logger.Error("Failed to call RAG service and getting a query answer",
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		msg = tgbotapi.NewMessage(chatID, "Sorry, I could not process your request, please try again later")
		msg.ReplyMarkup = GeneralInquiryKeyboard
		env.API.Send(msg)
		return
	} else {
		metrics.RAGServiceCallsTotal.WithLabelValues("GetAnswerToQuery", "success").Inc()
	}

	env.Logger.Info("RAG answer received",
		zap.Int64("user_id", userID),
		zap.String("query", text),
		zap.String("answer_len", strconv.Itoa(len(ragAnsStruct.Answer))),
	)
	msg = tgbotapi.NewMessage(chatID, ragAnsStruct.Answer)
	msg.ReplyMarkup = GeneralInquiryKeyboard
	env.API.Send(msg)
}
