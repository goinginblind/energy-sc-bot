package handlers

import (
	"context"
	"fmt"
	"math/rand"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/common"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/states"
)

func HandleAwaitingLoginInput(ctx context.Context, env *common.HandlerEnv, chatID, userID int64, text string) {
	env.Logger.Info("User submitted login identifier",
		zap.Int64("user_id", userID),
		zap.String("identifier", text),
	)
	// In a real scenario, we would validate the phone/email format here.
	// For this project, we'll assume the input is a phone number.
	env.Store.SetUserData(ctx, userID, "login_identifier", text)

	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	env.Store.SetUserData(ctx, userID, "otp", otp)
	env.Store.SetUserState(ctx, userID, states.StateAwaitingOTP)

	env.Logger.Info("Generated OTP for user",
		zap.Int64("user_id", userID),
		zap.String("otp", otp),
		zap.String("status", "STUB: For testing purposes"),
	)
	env.API.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("A confirmation code has been sent to your number. Please enter it.\n\n(For testing: your code is %s)", otp)))
}

func HandleAwaitingOTP(ctx context.Context, env *common.HandlerEnv, chatID, userID int64, text string) {
	storedOTP, err := env.Store.GetUserData(ctx, userID, "otp")
	if err != nil || storedOTP == "" {
		env.Logger.Error("Session error or OTP not found",
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		env.API.Send(tgbotapi.NewMessage(chatID, "A session error has occurred. Please try to log in again."))
		env.Store.SetUserState(ctx, userID, states.StateStart)
		return
	}

	if text == storedOTP {
		loginIdentifier, _ := env.Store.GetUserData(ctx, userID, "login_identifier")

		// Fetch user from the data service
		user, err := env.DataClient.GetUserByPhone(loginIdentifier)
		if err != nil {
			env.Logger.Error("Failed to get user by phone",
				zap.String("phone", loginIdentifier),
				zap.Error(err),
			)
			env.API.Send(tgbotapi.NewMessage(chatID, "Could not find your account. Please check your login details or contact support."))
			env.Store.SetUserState(ctx, userID, states.StateStart)
			return
		}

		// Clear temporary login data and set the user as logged in
		env.Store.ClearUserData(ctx, userID)
		env.Store.SetUserState(ctx, userID, states.StateLoggedIn)
		env.Store.SetUserData(ctx, userID, "logged_in", "true")
		env.Store.SetUserData(ctx, userID, "user_id", fmt.Sprintf("%d", user.ID)) // Store database ID

		userName := user.Email.String // Assuming name is not available, use email
		if !user.Email.Valid {
			userName = "Valued Customer"
		}

		env.Logger.Info("User logged in successfully",
			zap.Int64("user_id", userID),
			zap.String("username", userName),
		)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Login successful!\n\nWelcome, %s!", userName))
		msg.ReplyMarkup = LoggedInKeyboard
		env.API.Send(msg)
	} else {
		env.Logger.Warn("Incorrect OTP entered",
			zap.Int64("user_id", userID),
			zap.String("entered_otp", text),
		)
		env.API.Send(tgbotapi.NewMessage(chatID, "🤮 Incorrect code. Please try again or start over /start."))
		env.Store.SetUserState(ctx, userID, states.StateAwaitingOTP)
	}
}


