package bot

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Юзер стейты
const (
	StateStart                  = ""
	StateAwaitingLoginInput     = "awaiting_login_input"
	StateAwaitingOTP            = "awaiting_otp"
	StateLoggedIn               = "logged_in"
	StateGeneralInquiry         = "general_inquiry"
	StateAwaitingAgentIssuePost = "awaiting_agent_issue_post"
	StateAgentChat              = "agent_chat"
)

// Объявление глобаль клавиатур (реюзабилити агаа)
var (
	// `Приветствие: Общий запрос или Вход?`
	welcomeKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔎 Общий запрос"),
			tgbotapi.NewKeyboardButton("🔑 Вход"),
		),
	)

	// `Основные опции аккаунта`
	loggedInKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🧾 Мои счета"),
			tgbotapi.NewKeyboardButton("🧑‍💼 Связаться с агентом"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❓ Задать общий вопрос"),
			tgbotapi.NewKeyboardButton("🚪 Выход"),
		),
	)

	// `Инлайн: PDF, Оплата, Агент`
	billOptionsKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📄 Скачать PDF", "bill_pdf"),
			tgbotapi.NewInlineKeyboardButtonData("💳 Оплатить", "bill_pay"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧑‍💼 Связаться с оператором", "bill_agent"),
		),
	)

	// Инлайн кнопка для выхода из режима РАГ-запросов
	generalInquiryKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏁 Завершить диалог", "end_general_inquiry"),
		),
	)

	// Инлайн кнопка для выхода из чата с агентом
	agentChatKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏁 Завершить чат", "end_agent_chat"),
		),
	)
)

// Бот стракт для хранения депенденсиес
type Bot struct {
	api      *tgbotapi.BotAPI
	handlers *Handlers
	store    Store
}

// Нью бот создает новый экземпляр бота.
// Он принимает API, обработчики и хранилище в качестве зависимостей
func New(api *tgbotapi.BotAPI, handlers *Handlers, store Store) *Bot {
	return &Bot{
		api:      api,
		handlers: handlers,
		store:    store,
	}
}

// Начинает основной цикл обновлений бота
func (b *Bot) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	log.Printf("Authorized on account %s, starting update loop...", b.api.Self.UserName)

	// Основной луп всей логики бота
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
			// Обработка нажатий на инлайн клавиатуру
			isCallback = true
			callback := update.CallbackQuery
			chatID = callback.Message.Chat.ID
			userID = callback.From.ID
			text = callback.Data

			// Ответ, он нужен чтобы кнопка не переливалась (состояние загрузки)
			b.api.Request(tgbotapi.NewCallback(callback.ID, ""))
		} else {
			// Обычные сообщения, без клавиатур == тупа текст
			chatID = update.Message.Chat.ID
			userID = update.Message.From.ID
			text = update.Message.Text
			// Каждое текстовое соо логируется
			b.store.SaveMessage(ctx, userID, text)
		}

		// Достаем куррент стейт с редиса
		userState, err := b.store.GetUserState(ctx, userID)
		if err != nil {
			log.Printf("ERROR getting user state for %d: %v", userID, err)
			continue
		}

		// Махина для обработки стейта
		switch userState {
		case StateStart:
			b.handlers.HandleStartState(ctx, chatID, userID, text)
		case StateAwaitingLoginInput:
			b.handlers.HandleAwaitingLoginInput(ctx, chatID, userID, text)
		case StateAwaitingOTP:
			b.handlers.HandleAwaitingOTP(ctx, chatID, userID, text)
		case StateLoggedIn:
			b.handlers.HandleLoggedInState(ctx, chatID, userID, text, isCallback)
		case StateGeneralInquiry:
			b.handlers.HandleGeneralInquiryState(ctx, chatID, userID, text, isCallback)
		case StateAwaitingAgentIssuePost:
			b.handlers.HandleAgentIssue(ctx, chatID, userID, text)
		case StateAgentChat:
			b.handlers.HandleAgentChat(ctx, chatID, userID, text, isCallback)
		default:
			// Незнакомый стейт = ошибка, но такого быть не должно
			b.store.SetUserState(ctx, userID, StateStart)
			msg := tgbotapi.NewMessage(chatID, "Произошла ошибка. Давайте начнем сначала.")
			msg.ReplyMarkup = welcomeKeyboard
			b.api.Send(msg)
		}
	}
}
