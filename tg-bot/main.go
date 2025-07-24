package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// Юзер стейты
const (
	StateStart                  = ""
	StateAwaitingLoginInput     = "awaiting_login_input"
	StateAwaitingOTP            = "awaiting_otp"
	StateLoggedIn               = "logged_in"
	StateGeneralInquiry         = "general_inquiry"
	StateAwaitingAgentIssuePre  = "awaiting_agent_issue_pre"
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
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("FATAL: TELEGRAM_TOKEN environment variable is not set")
	}

	store := NewRedisStore()
	defer store.Close()

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Основной луп всей логики бота
	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		ctx := context.Background()
		var chatID int64
		var userID int64
		var text string

		isCallback := update.CallbackQuery != nil

		if isCallback {
			// Обработка нажатий на инлайн клавиатуру
			callback := update.CallbackQuery
			chatID = callback.Message.Chat.ID
			userID = callback.From.ID
			text = callback.Data

			// Ответ, он нужен чтобы кнопка не переливалась (состояние загрузки)
			bot.Request(tgbotapi.NewCallback(callback.ID, ""))

		} else {
			// Обычные сообщения, без клавиатур == тупа текст
			chatID = update.Message.Chat.ID
			userID = update.Message.From.ID
			text = update.Message.Text
			// Каждое текстовое соо логируется
			store.SaveMessage(ctx, userID, text)
		}

		// Достаем куррент стейт с редиса
		userState, err := store.GetUserState(ctx, userID)
		if err != nil {
			log.Printf("ERROR getting user state for %d: %v", userID, err)
			continue
		}

		// Махина для обработки стейта
		switch userState {
		case StateStart:
			handleStartState(ctx, bot, store, chatID, userID, text)
		case StateAwaitingLoginInput:
			handleAwaitingLoginInput(ctx, bot, store, chatID, userID, text)
		case StateAwaitingOTP:
			handleAwaitingOTP(ctx, bot, store, chatID, userID, text)
		case StateLoggedIn:
			handleLoggedInState(ctx, bot, store, chatID, userID, text, isCallback)
		case StateGeneralInquiry:
			handleGeneralInquiryState(ctx, bot, store, chatID, userID, text)
		case StateAwaitingAgentIssuePre, StateAwaitingAgentIssuePost:
			handleAgentIssue(ctx, bot, store, chatID, userID, text)
		case StateAgentChat:
			handleAgentChat(ctx, bot, chatID, text)
		default:
			if text == "/start" {
				handleStartState(ctx, bot, store, chatID, userID, text)
			} else {
				// Незнакомый стейт = ошибка, но такого быть не должно
				msg := tgbotapi.NewMessage(chatID, "Произошла ошибка. Давайте начнем сначала.")
				msg.ReplyMarkup = welcomeKeyboard
				bot.Send(msg)
				store.SetUserState(ctx, userID, StateStart)
			}
		}
	}
}

// хэндлер изначального стейта
func handleStartState(ctx context.Context, bot *tgbotapi.BotAPI, store Store, chatID, userID int64, text string) {
	switch text {
	case "🔑 Вход", "/login":
		store.SetUserState(ctx, userID, StateAwaitingLoginInput)
		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, введите ваш телефон или email для входа.")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		bot.Send(msg)
	case "🔎 Общий запрос":
		store.SetUserState(ctx, userID, StateGeneralInquiry)
		msg := tgbotapi.NewMessage(chatID, "Вы можете задать любой общий вопрос. Я постараюсь на него ответить.\n\nИли вы можете войти в свой аккаунт для персональных опций.")
		bot.Send(msg)
	default:
		msg := tgbotapi.NewMessage(chatID, "Здравствуйте! Я ваш виртуальный помощник. Чем могу помочь?")
		msg.ReplyMarkup = welcomeKeyboard
		bot.Send(msg)
	}
}

// хэндлер когда боту необходим номер телефона или емаил
func handleAwaitingLoginInput(ctx context.Context, bot *tgbotapi.BotAPI, store Store, chatID, userID int64, text string) {
	// TODO: валидация юзер месаж и отправка (или не отправка) оне тайм пассворда
	// Пока пусть будет рандомный инт из 6 знаков.
	log.Printf("User %d submitted login identifier: %s", userID, text)

	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	store.SetUserData(ctx, userID, "otp", otp)
	store.SetUserState(ctx, userID, StateAwaitingOTP)

	log.Printf("STUB: Generated OTP for user %d: %s", userID, otp)
	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("На ваш номер был отправлен код подтверждения. Пожалуйста, введите его.\n\n(Для теста: ваш код - %s)", otp)))
}

// Когда боту нужен отп
func handleAwaitingOTP(ctx context.Context, bot *tgbotapi.BotAPI, store Store, chatID, userID int64, text string) {
	storedOTP, err := store.GetUserData(ctx, userID, "otp")
	if err != nil || storedOTP == "" {
		bot.Send(tgbotapi.NewMessage(chatID, "Произошла ошибка сессии. Попробуйте войти снова."))
		store.SetUserState(ctx, userID, StateStart)
		return
	}

	if text == storedOTP {
		// отп верный
		store.ClearUserData(ctx, userID)
		store.SetUserState(ctx, userID, StateLoggedIn)

		// TODO: тут нужно добавить апи второго data-сервиса, где мы имитируем БД реальной конторы с пользовательскими данными
		userName := "Пользователь"

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Вход выполнен!\n\nДобро пожаловать, %s!", userName))
		msg.ReplyMarkup = loggedInKeyboard
		bot.Send(msg)
	} else {
		// отп неверный 🤮
		bot.Send(tgbotapi.NewMessage(chatID, "🤮 Неверный код. Попробуйте еще раз или начните сначала /start."))
	}
}

// юзер залогинился
func handleLoggedInState(ctx context.Context, bot *tgbotapi.BotAPI, store Store, chatID, userID int64, text string, isCallback bool) {
	// Нажатие на инлайн клавы
	if isCallback {
		switch text {
		case "bill_pdf":
			// TODO: доставать пдф или генерировать его
			bot.Send(tgbotapi.NewMessage(chatID, "Ваш PDF-счет генерируется и скоро будет отправлен..."))
		case "bill_pay":
			// TODO: оплата? Всм, я естественно этим заниматься не буду
			// пусть они уже в продакшене это делают у меня ваще нет денег как концепта лол
			bot.Send(tgbotapi.NewMessage(chatID, "Перенаправляем на страницу оплаты..."))
		case "bill_agent":
			store.SetUserState(ctx, userID, StateAwaitingAgentIssuePost)
			bot.Send(tgbotapi.NewMessage(chatID, "Пожалуйста, опишите вашу проблему, связанную с этим счетом. Вся информация будет передана сотруднику службы поддержки."))
		}
		return
	}

	// Другие клавы или текст месаж
	switch text {
	case "🧾 Мои счета":
		// TODO: тут нужно фетчить с апи data-сервиса.
		summary := "Счет #12345 от 24.07.2025\nСумма: $420.69\nСтатус: Не оплачен"
		msg := tgbotapi.NewMessage(chatID, summary)
		msg.ReplyMarkup = billOptionsKeyboard
		bot.Send(msg)
	case "🧑‍💼 Связаться с агентом":
		store.SetUserState(ctx, userID, StateAwaitingAgentIssuePost)
		bot.Send(tgbotapi.NewMessage(chatID, "Пожалуйста, опишите вашу проблему. Агент поддержки скоро с вами свяжется."))
	case "❓ Задать общий вопрос":
		store.SetUserState(ctx, userID, StateGeneralInquiry)
		bot.Send(tgbotapi.NewMessage(chatID, "Вы можете задать любой общий вопрос. Чтобы вернуться в меню аккаунта, нажмите /start."))
	case "🚪 Выход", "/logout":
		store.SetUserState(ctx, userID, StateStart)
		msg := tgbotapi.NewMessage(chatID, "Вы успешно вышли из системы.")
		msg.ReplyMarkup = welcomeKeyboard
		bot.Send(msg)
	default:
		bot.Send(tgbotapi.NewMessage(chatID, "Пожалуйста, используйте кнопки меню."))
	}
}

// Общие вопросы без херни и логинов (Антоха?)
func handleGeneralInquiryState(ctx context.Context, bot *tgbotapi.BotAPI, store Store, chatID, userID int64, text string) {
	if text == "/start" {
		store.SetUserState(ctx, userID, StateStart)
		msg := tgbotapi.NewMessage(chatID, "Возвращаемся в главное меню...")
		msg.ReplyMarkup = welcomeKeyboard
		bot.Send(msg)
		return
	}

	// TODO: вызов РАГ ч/з gRPC..
	log.Printf("STUB: RAG query from user %d: %s", userID, text)
	ragAnswer := "Это мог бы быть ответ, сгенерированный RAG-моделью, но бот пока так не умеет (я так и не начал gRPC штуку)."

	bot.Send(tgbotapi.NewMessage(chatID, ragAnswer))
	bot.Send(tgbotapi.NewMessage(chatID, "Могу ли я помочь чем-то еще? Вы всегда можете вернуться в главное меню с помощью /start."))
}

// Когда бот ждет описания проблемы для передачи мяясному мешку (службе п.)
func handleAgentIssue(ctx context.Context, bot *tgbotapi.BotAPI, store Store, chatID, userID int64, text string) {
	// TODO: здесь должен будет появиться тикет для жиры или чего-то вроде того
	log.Printf("STUB: Creating support ticket for user %d. Issue: %s", userID, text)

	store.SetUserState(ctx, userID, StateAgentChat)
	bot.Send(tgbotapi.NewMessage(chatID, "Спасибо! Ваше обращение передано агенту. Вы вошли в режим чата с поддержкой. Все последующие сообщения будут направлены агенту.\n\nЧтобы завершить чат, отправьте команду /endchat."))
}

// Вот наш активный чат с поддержкой
func handleAgentChat(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, text string) {
	if text == "/endchat" {
		// Тут надо бы передать поддержке что чат завершили (но по-моему оверкилл)
		bot.Send(tgbotapi.NewMessage(chatID, "Чат с агентом завершен. Возвращаемся в главное меню."))
		// И отправка юзера в старт меню потому что я не знаю что делать ещё
		msg := tgbotapi.NewMessage(chatID, "Чем я могу помочь?")
		msg.ReplyMarkup = welcomeKeyboard
		bot.Send(msg)
		return
	}

	// TODO: передача сообщения в чат поддержки, это вообще в теории можно сделать через бота, но нужно тогда несколько
	// типов юзеров, а это черезчур сейчас
	// либо в теории это мог бы быть отдельный веб-апп, либо внешний серврер
	log.Printf("STUB: Forwarding message to agent from chat %d: %s", chatID, text)
	// И по-хорошему юзеру бы сообщить, что его сообщение передано в поддержку реальному человеку (скорее всего юзер зол и
	// был бы рад увидеть что ему наконец-то поможет не бот)
	bot.Send(tgbotapi.NewMessage(chatID, "Ваше сообщение передано работнику службы поддержки."))
}
