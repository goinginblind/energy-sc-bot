package bot

/*Вообще, стоит тут сказать о том, что в Handlers очень перекрывается с Bot
но это так и задумано, поскольку их функции отличаются и тут можно отметить, что
Bot == приемник, Handlers == раздатчик, мне кажется что такое разделение в принципе лучше*/

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goinginblind/energy-sc-bot/tg-bot/ragpb"
)

// Handlers содержит зависимости необходимые для обработки сообщений бота.
// Он включает в себя API бота, хранилище и gRPC клиент для RAG-сервиса.
type Handlers struct {
	api       *tgbotapi.BotAPI       // через апи происходит прием и отправка сообщений
	store     Store                  // редисный интерфейс
	ragClient ragpb.RAGServiceClient // gRPC клиент
}

// Новые хэндлеры
func NewHandlers(bot *tgbotapi.BotAPI, store Store, client ragpb.RAGServiceClient) *Handlers {
	return &Handlers{
		api:       bot,
		store:     store,
		ragClient: client,
	}
}

// хэндлер изначального стейта
func (h *Handlers) HandleStartState(ctx context.Context, chatID, userID int64, text string) {
	switch text {
	case "🔑 Вход", "/login":
		h.store.SetUserState(ctx, userID, StateAwaitingLoginInput)
		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, введите ваш телефон или email для входа.")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		h.api.Send(msg)
	case "🔎 Общий запрос":
		h.store.SetUserState(ctx, userID, StateGeneralInquiry)
		msg := tgbotapi.NewMessage(chatID, "Вы в режиме общего запроса. Просто напишите свой вопрос. Чтобы выйти, используйте кнопку ниже или команду /start.")
		h.api.Send(msg)
	case "/start":
		msg := tgbotapi.NewMessage(chatID, "Здравствуйте! Я ваш виртуальный помощник. Чем могу помочь?")
		msg.ReplyMarkup = welcomeKeyboard
		h.api.Send(msg)
	default:
		// Если юзер сразу пишет вопрос, переходим в режим РАГ
		h.store.SetUserState(ctx, userID, StateGeneralInquiry)
		h.HandleGeneralInquiryState(ctx, chatID, userID, text, false)
	}
}

// Ожидание логина
func (h *Handlers) HandleAwaitingLoginInput(ctx context.Context, chatID, userID int64, text string) {
	log.Printf("User %d submitted login identifier: %s", userID, text)
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	h.store.SetUserData(ctx, userID, "otp", otp)
	h.store.SetUserState(ctx, userID, StateAwaitingOTP)
	log.Printf("STUB: Generated OTP for user %d: %s", userID, otp)
	h.api.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("На ваш номер был отправлен код подтверждения. Пожалуйста, введите его.\n\n(Для теста: ваш код - %s)", otp)))
}

// юзер залогинился
func (h *Handlers) HandleLoggedInState(ctx context.Context, chatID, userID int64, text string, isCallback bool) {
	// Нажатие на инлайн клавы
	if isCallback {
		switch text {
		case "bill_pdf":
			// TODO: доставать пдф или генерировать его
			h.api.Send(tgbotapi.NewMessage(chatID, "Ваш PDF-счет генерируется и скоро будет отправлен..."))
		case "bill_pay":
			// TODO: оплата? Всм, я естественно этим заниматься не буду
			// у меня ваще нет денег как концепта
			h.api.Send(tgbotapi.NewMessage(chatID, "Перенаправляем на страницу оплаты..."))
		case "bill_agent":
			h.store.SetUserState(ctx, userID, StateAwaitingAgentIssuePost)
			h.api.Send(tgbotapi.NewMessage(chatID, "Пожалуйста, опишите вашу проблему, связанную с этим счетом. Вся информация будет передана сотруднику службы поддержки."))
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
		h.api.Send(msg)
	case "🧑‍💼 Связаться с агентом":
		h.store.SetUserState(ctx, userID, StateAwaitingAgentIssuePost)
		h.api.Send(tgbotapi.NewMessage(chatID, "Пожалуйста, опишите вашу проблему. Агент поддержки скоро с вами свяжется."))
	case "❓ Задать общий вопрос":
		h.store.SetUserState(ctx, userID, StateGeneralInquiry)
		h.api.Send(tgbotapi.NewMessage(chatID, "Вы можете задать любой общий вопрос. Чтобы вернуться в меню аккаунта, используйте кнопку ниже или команду /start."))
	case "🚪 Выход", "/logout":
		// **FIX: Очищаем все данные пользователя при выходе, включая флаг "logged_in"**
		h.store.ClearUserData(ctx, userID)
		h.store.SetUserState(ctx, userID, StateStart)
		msg := tgbotapi.NewMessage(chatID, "Вы успешно вышли из системы.")
		msg.ReplyMarkup = welcomeKeyboard
		h.api.Send(msg)
	default:
		h.api.Send(tgbotapi.NewMessage(chatID, "Пожалуйста, используйте кнопки меню."))
	}
}

// текстовые вопросы - сюдаа
func (h *Handlers) HandleGeneralInquiryState(ctx context.Context, chatID, userID int64, text string, isCallback bool) {
	// **FIX: Этот блок теперь проверяет, был ли юзер залогинен**
	if isCallback && text == "end_general_inquiry" {
		loginStatus, _ := h.store.GetUserData(ctx, userID, "logged_in")
		if loginStatus == "true" {
			// Если был залогинен, возвращаем в меню аккаунта
			h.store.SetUserState(ctx, userID, StateLoggedIn)
			msg := tgbotapi.NewMessage(chatID, "Диалог завершен. Возвращаемся в меню вашего аккаунта.")
			msg.ReplyMarkup = loggedInKeyboard
			h.api.Send(msg)
		} else {
			// Если не был, возвращаем в главное меню
			h.store.SetUserState(ctx, userID, StateStart)
			msg := tgbotapi.NewMessage(chatID, "Диалог завершен. Возвращаемся в главное меню.")
			msg.ReplyMarkup = welcomeKeyboard
			h.api.Send(msg)
		}
		return
	}

	// Не даем командам с клавиатуры уходить в РАГ
	if !isCallback && (text == "🔎 Общий запрос" || text == "🔑 Вход") {
		h.store.SetUserState(ctx, userID, StateStart)
		h.HandleStartState(ctx, chatID, userID, text)
		return
	}

	// TODO: вызов РАГ ч/з gRPC..
	var msg tgbotapi.MessageConfig
	ragLabelStruct, err := h.ragClient.ClassifyQuery(ctx, &ragpb.ClassifyRequest{Query: text})
	if err != nil {
		log.Printf("Failed to call RAG service and classify request: %d", err)
		msg = tgbotapi.NewMessage(chatID, "Извините, я не смог обработать ваш запрос, пожалуйста, попробуйте позже")
		msg.ReplyMarkup = generalInquiryKeyboard
		h.api.Send(msg)
		return
	}

	userHistory, _ := h.store.GetHistory(ctx, userID, 10)
	ragAnsStruct, err := h.ragClient.GetAnswerToQuery(ctx, &ragpb.AnswerRequest{Query: text, History: userHistory, Label: ragLabelStruct.Label})
	if err != nil {
		log.Printf("Failed to call RAG service and getting a query answer: %d", err)
		msg = tgbotapi.NewMessage(chatID, "Извините, я не смог обработать ваш запрос, пожалуйста, попробуйте позже")
		msg.ReplyMarkup = generalInquiryKeyboard
		h.api.Send(msg)
		return
	}

	msg = tgbotapi.NewMessage(chatID, ragAnsStruct.Answer)
	msg.ReplyMarkup = generalInquiryKeyboard
	h.api.Send(msg)
}

// Когда бот ждет описания проблемы для передачи мяясному мешку (службе п.)
func (h *Handlers) HandleAgentIssue(ctx context.Context, chatID, userID int64, text string) {
	// TODO: здесь должен будет появиться тикет для жиры или чего-то вроде того
	log.Printf("STUB: Creating support ticket for user %d. Issue: %s", userID, text)

	h.store.SetUserState(ctx, userID, StateAgentChat)
	msg := tgbotapi.NewMessage(chatID, "Спасибо! Ваше обращение передано агенту. Вы вошли в режим чата с поддержкой. Все последующие сообщения будут направлены агенту.\n\nЧтобы завершить чат, используйте кнопку ниже.")
	msg.ReplyMarkup = agentChatKeyboard
	h.api.Send(msg)
}

// Вот наш активный чат с поддержкой
func (h *Handlers) HandleAgentChat(ctx context.Context, chatID, userID int64, text string, isCallback bool) {
	if isCallback && text == "end_agent_chat" {
		// Тут надо бы передать поддержке что чат завершили (но по-моему оверкилл)
		h.api.Send(tgbotapi.NewMessage(chatID, "Чат с агентом завершен. Возвращаемся в меню."))

		// Проверяем, был ли юзер залогинен, чтобы вернуть его в правильное меню
		loginStatus, _ := h.store.GetUserData(ctx, userID, "logged_in")
		if loginStatus == "true" {
			h.store.SetUserState(ctx, userID, StateLoggedIn)
			msg := tgbotapi.NewMessage(chatID, "Чем я могу помочь?")
			msg.ReplyMarkup = loggedInKeyboard
			h.api.Send(msg)
		} else {
			// Этот случай теоретически невозможен по текущей логике, но для надежности
			h.store.SetUserState(ctx, userID, StateStart)
			msg := tgbotapi.NewMessage(chatID, "Чем я могу помочь?")
			msg.ReplyMarkup = welcomeKeyboard
			h.api.Send(msg)
		}
		return
	}

	// Если это не коллбэк, значит это сообщение для агента
	if !isCallback {
		// TODO: передача сообщения в чат поддержки...
		log.Printf("STUB: Forwarding message to agent from chat %d: %s", chatID, text)
	}
}

// Когда боту нужен отп
func (h *Handlers) HandleAwaitingOTP(ctx context.Context, chatID, userID int64, text string) {
	storedOTP, err := h.store.GetUserData(ctx, userID, "otp")
	if err != nil || storedOTP == "" {
		h.api.Send(tgbotapi.NewMessage(chatID, "Произошла ошибка сессии. Попробуйте войти снова."))
		h.store.SetUserState(ctx, userID, StateStart)
		return
	}

	if text == storedOTP {
		// отп верный
		h.store.ClearUserData(ctx, userID)
		h.store.SetUserState(ctx, userID, StateLoggedIn)
		// **FIX: Устанавливаем флаг, что пользователь залогинен**
		h.store.SetUserData(ctx, userID, "logged_in", "true")

		// TODO: тут нужно добавить апи второго data-сервиса, где мы имитируем БД реальной конторы с пользовательскими данными
		userName := "Пользователь"

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Вход выполнен!\n\nДобро пожаловать, %s!", userName))
		msg.ReplyMarkup = loggedInKeyboard
		h.api.Send(msg)
	} else {
		// отп неверный 🤮
		h.api.Send(tgbotapi.NewMessage(chatID, "🤮 Неверный код. Попробуйте еще раз или начните сначала /start."))
	}
}
