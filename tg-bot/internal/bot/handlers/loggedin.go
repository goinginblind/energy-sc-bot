package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/common"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/states"
)

var (
	LoggedInKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🧾 My Bills"),
			tgbotapi.NewKeyboardButton("🧑‍💼 Contact Agent"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❓ Ask a General Question"),
			tgbotapi.NewKeyboardButton("🚪 Logout"),
		),
	)

	BillOptionsKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📄 Download PDF", "bill_pdf"),
			tgbotapi.NewInlineKeyboardButtonData("💳 Pay Bill", "bill_pay"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧑‍💼 Contact Agent about Bill", "bill_agent"),
		),
	)
)

func HandleLoggedInState(ctx context.Context, env *common.HandlerEnv, chatID, userID int64, text string, isCallback bool) {
	// Handling inline keyboard presses
	if isCallback {
		switch text {
		case "bill_pdf":
			// TODO: get a pdf or generate it
			env.API.Send(tgbotapi.NewMessage(chatID, "Your PDF bill is being generated and will be sent shortly..."))
		case "bill_pay":
			// TODO: payment? I mean, I'm not going to do that, of course.
			// I don't even have money as a concept.
			env.API.Send(tgbotapi.NewMessage(chatID, "Redirecting to the payment page..."))
		case "bill_agent":
			env.Store.SetUserState(ctx, userID, states.StateAwaitingAgentIssuePost)
			env.API.Send(tgbotapi.NewMessage(chatID, "Please describe your problem with this bill. All information will be passed on to a support employee."))
		}
		return
	}

	// Other keyboards or text messages
	switch text {
	case "🧾 My Bills":
		dbUserIDstr, err := env.Store.GetUserData(ctx, userID, "user_id")
		if err != nil {
			log.Printf("Failed to get user_id from store for telegram user %d: %v", userID, err)
			env.API.Send(tgbotapi.NewMessage(chatID, "Could not retrieve your session information. Please try logging in again."))
			return
		}
		dbUserID, _ := strconv.ParseInt(dbUserIDstr, 10, 64)

		bills, err := env.DataClient.GetBillsByUserID(dbUserID)
		if err != nil {
			log.Printf("Failed to get bills for user %d: %v", dbUserID, err)
			env.API.Send(tgbotapi.NewMessage(chatID, "Could not retrieve your bills at this time. Please try again later."))
			return
		}

		if len(bills) == 0 {
			env.API.Send(tgbotapi.NewMessage(chatID, "You have no bills on record."))
			return
		}

		var summary strings.Builder
		summary.WriteString("Here are your recent bills:\n\n")
		for _, bill := range bills {
			summary.WriteString(fmt.Sprintf("Bill #%d from %s\n", bill.ID, bill.IssuedAt.Format("02 Jan 2006")))
			summary.WriteString(fmt.Sprintf("Amount: $%.2f\n", bill.Amount))
			summary.WriteString(fmt.Sprintf("Status: %s\n\n", bill.Status))
		}

		msg := tgbotapi.NewMessage(chatID, summary.String())
		msg.ReplyMarkup = BillOptionsKeyboard
		env.API.Send(msg)
	case "🧑‍💼 Contact Agent":
		env.Store.SetUserState(ctx, userID, states.StateAwaitingAgentIssuePost)
		env.API.Send(tgbotapi.NewMessage(chatID, "Please describe your problem. A support agent will contact you shortly."))
	case "❓ Ask a General Question":
		env.Store.SetUserState(ctx, userID, states.StateGeneralInquiry)
		env.API.Send(tgbotapi.NewMessage(chatID, "You can ask any general question. To return to the account menu, use the button below or the /start command."))
	case "🚪 Logout", "/logout":
		// **FIX: Clear all user data on logout, including the "logged_in" flag**
		env.Store.ClearUserData(ctx, userID)
		env.Store.SetUserState(ctx, userID, states.StateStart)
		msg := tgbotapi.NewMessage(chatID, "You have been successfully logged out.")
		msg.ReplyMarkup = WelcomeKeyboard
		env.API.Send(msg)
	default:
		env.API.Send(tgbotapi.NewMessage(chatID, "Please use the menu buttons."))
	}
}
